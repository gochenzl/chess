package rpc

import (
	"bufio"
	"net"
	"sync"
	"time"

	"github.com/gochenzl/chess/util/log"
)

type myConn struct {
	net.Conn
	br *bufio.Reader
	//bw *bufio.Writer
}

type connPoolInfo struct {
	mutex         sync.Mutex
	conns         []net.Conn
	connNumber    int
	maxConnNumber int
	hostAndPort   string
	name          string
}

func (conn *myConn) Read(b []byte) (n int, err error) {
	n, err = conn.br.Read(b)
	return
}

//func (conn *myConn) Write(b []byte) (n int, err error) {
//	n, err = conn.bw.Write(b)
//	conn.bw.Flush()
//	return
//}

func newConnPool(maxConn int, hostAndPort string, name string) *connPoolInfo {
	var info connPoolInfo
	info.maxConnNumber = maxConn
	info.hostAndPort = hostAndPort
	info.name = name
	info.conns = make([]net.Conn, 0, maxConn)
	return &info
}

func (pool *connPoolInfo) createConn() net.Conn {
	conn, err := net.Dial("tcp", pool.hostAndPort)
	if err != nil {
		log.Error("%s", err.Error())
		return nil
	}

	return &myConn{conn, bufio.NewReader(conn)}
}

func (pool *connPoolInfo) get() net.Conn {
	for i := 0; i < 10; i++ {
		pool.mutex.Lock()

		if len(pool.conns) == 0 {
			if pool.connNumber < pool.maxConnNumber {
				pool.mutex.Unlock()
				c := pool.createConn()
				if c == nil {
					return nil
				}

				pool.mutex.Lock()
				pool.connNumber++
				pool.mutex.Unlock()
				return c
			} else {
				pool.mutex.Unlock()
				log.Warn("user connection pool is empty")
				time.Sleep(time.Millisecond * 50)
				continue
			}

		}

		c := pool.conns[len(pool.conns)-1]
		pool.conns = pool.conns[:len(pool.conns)-1]
		pool.mutex.Unlock()
		return c
	}

	return nil
}

func (pool *connPoolInfo) release(c net.Conn) {
	pool.mutex.Lock()
	pool.conns = append(pool.conns, c)
	pool.mutex.Unlock()
}

func (pool *connPoolInfo) decrConnNum(c net.Conn) {
	pool.mutex.Lock()
	pool.connNumber--
	if pool.connNumber < 0 {
		pool.connNumber = 0
	}
	pool.mutex.Unlock()
}
