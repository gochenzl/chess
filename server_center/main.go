package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gochenzl/chess/server_center/conn_info"
	"github.com/gochenzl/chess/server_center/handler"
	"github.com/gochenzl/chess/util/conf"
	"github.com/gochenzl/chess/util/log"
	"github.com/gochenzl/chess/util/rpc"
	"github.com/golang/protobuf/proto"
)

var config struct {
	ListenPort int    `ini:"listen_port"`
	DataPath   string `ini:"data_path"`
}

func initConfig(confPath string) bool {
	if err := conf.LoadIniFromFile(confPath+"/center.conf", &config); err != nil {
		log.Error("init config fail:%s", err.Error())
		return false
	}

	return true
}

var server *rpc.Server

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s conf_path\n", os.Args[0])
		return
	}

	log.Info("server start, pid = %d", os.Getpid())

	if !initConfig(os.Args[1]) {
		return
	}

	if !conn_info.Init(config.DataPath) {
		return
	}

	server = rpc.NewServer(config.ListenPort)
	server.SetConnHandler(handleConn)

	go doSignal()

	if err := server.Run(); err != nil {
		log.Error("run server fail:%s", err.Error())
		return
	}

	conn_info.Close()
	log.Info("exit graceful")
}

func doSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	signal := <-ch
	log.Info("receive signal %s", signal.String())
	server.Stop()
}

func handleConn(conn net.Conn) {
	log.Info("new connection from %s", conn.RemoteAddr().String())

	br := bufio.NewReaderSize(conn, 1024)

	defer conn.Close()
	defer handler.RemoveClient(conn)
	defer server.Done()

	for {
		if server.CheckStop() {
			return
		}

		conn.SetDeadline(time.Now().Add(time.Second * 3))

		req, err := rpc.DecodePb(br)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}

			log.Error("connection from %s error: %s", conn.RemoteAddr().String(), err.Error())
			return
		}

		name := proto.MessageName(req)
		log.Info("receive request %s: %s", name, req.String())

		switch name {
		case "center.AddConnInfoReq":
			handler.HandleAddConnInfo(conn, req)
		case "center.DelConnInfoReq":
			handler.HandleDelConnInfo(conn, req)
		case "center.DelConnInfoByGateidReq":
			handler.HandleDelConnInfoByGateid(conn, req)
		case "center.GetAllConnInfoReq":
			handler.HandleGetAllConnInfo(conn, req)
		default:
			log.Info("invalid message name:%s", name)
		}
	}
}
