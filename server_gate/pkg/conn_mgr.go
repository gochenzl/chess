package pkg

import (
	"io"
	"sync"
)

var connMgr struct {
	conns map[uint32]io.Writer
	mu    sync.RWMutex
}

func init() {
	connMgr.conns = make(map[uint32]io.Writer)
}

func getConn(connid uint32) io.Writer {
	connMgr.mu.RLock()
	defer connMgr.mu.RUnlock()

	if conn, ok := connMgr.conns[connid]; ok {
		return conn
	}

	return nil
}

func getAllConns() []io.Writer {
	connMgr.mu.RLock()
	defer connMgr.mu.RUnlock()

	if len(connMgr.conns) == 0 {
		return nil
	}

	conns := make([]io.Writer, 0, len(connMgr.conns))
	for _, v := range connMgr.conns {
		conns = append(conns, v)
	}

	return conns
}

func putConn(connid uint32, w io.Writer) {
	connMgr.mu.Lock()
	connMgr.conns[connid] = w
	connMgr.mu.Unlock()
}

func delConn(connid uint32) {
	connMgr.mu.Lock()
	delete(connMgr.conns, connid)
	connMgr.mu.Unlock()
}
