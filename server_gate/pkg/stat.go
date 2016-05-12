package pkg

import (
	"sync/atomic"
	"time"

	"github.com/gochenzl/chess/util/log"
)

var recvMsgCounter int64

func incRecvMsgCounter() {
	atomic.AddInt64(&recvMsgCounter, 1)
}

func printStat() {
	for {
		time.Sleep(time.Second * 5)

		counter := atomic.LoadInt64(&recvMsgCounter)
		log.Info("recv msg %d", counter)
	}
}
