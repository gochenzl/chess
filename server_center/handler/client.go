package handler

import (
	"io"
	"sync"

	"github.com/gochenzl/chess/util/log"
)

var clients []io.Writer = make([]io.Writer, 0, 100)
var clientMu sync.RWMutex

func addClient(client io.Writer) {
	clientMu.Lock()
	clients = append(clients, client)
	clientMu.Unlock()
}

func RemoveClient(client io.Writer) {
	clientMu.Lock()
	for i := 0; i < len(clients); i++ {
		if clients[i] == client {
			clients[i] = clients[len(clients)-1]
			clients = clients[:len(clients)-1]
			break
		}
	}
	clientMu.Unlock()
}

func sendClientNotify(buf []byte, excludeClient io.Writer) {
	clientMu.RLock()
	for i := 0; i < len(clients); i++ {
		if clients[i] == excludeClient {
			continue
		}

		if _, err := clients[i].Write(buf); err != nil {
			log.Error("sendNotify fail:%s", err.Error())
		}
	}
	clientMu.RUnlock()
}

func exist(client io.Writer) bool {
	clientMu.RLock()
	clientMu.RUnlock()
	for i := 0; i < len(clients); i++ {
		if client == clients[i] {
			return true
		}
	}

	return false
}
