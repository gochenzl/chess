package redis

import (
	"bufio"
	"net"
	"sync"

	"github.com/gochenzl/chess/util/log"
)

type Server struct {
	listenHostAndPort string
	store             Store
	listener          net.Listener
	allConns          map[net.Conn]bool
	connMu            sync.Mutex
}

func NewServer(hostAndPort string, store Store) *Server {
	var server Server
	server.listenHostAndPort = hostAndPort
	server.store = store
	server.allConns = make(map[net.Conn]bool)

	return &server
}

func (server *Server) addConn(conn net.Conn) {
	server.connMu.Lock()
	server.allConns[conn] = true
	server.connMu.Unlock()
}

func (server *Server) delConn(conn net.Conn) {
	server.connMu.Lock()
	delete(server.allConns, conn)
	server.connMu.Unlock()
}

func (server *Server) Run() error {
	var err error
	server.listener, err = net.Listen("tcp", server.listenHostAndPort)
	if err != nil {
		return err
	}

	go server.listenLoop()

	return nil
}

func (server *Server) Close() {
	server.listener.Close()

	server.connMu.Lock()
	for conn, _ := range server.allConns {
		conn.Close()
		delete(server.allConns, conn)
	}
	server.connMu.Unlock()
}

func (server *Server) listenLoop() {

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			log.Error("[redis server]accept fail:%s", err.Error())
			return
		}

		server.addConn(conn)
		go server.handleConn(conn)
	}
}

func (server *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	defer server.delConn(conn)

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	for {
		req, err := Parse(r)
		if err != nil {
			log.Error("[redis server]parse fail:%s", err.Error())
			return
		}

		rsp := processCmd(req, server.store)
		if !rsp.Pack(w) {
			return
		}

		if err := w.Flush(); err != nil {
			log.Error("flush fail:%s", err.Error())
			return
		}
	}
}
