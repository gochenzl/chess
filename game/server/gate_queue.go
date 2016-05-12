package server

import (
	"bytes"
	"time"

	"gopkg.in/redis.v3"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/game/config"
	"github.com/gochenzl/chess/util/hack"
	"github.com/gochenzl/chess/util/log"
)

type gateQueueInfo struct {
	gateid uint32
	connid uint32
	msgBuf []byte
}

var gateQueueInfoQ chan gateQueueInfo = make(chan gateQueueInfo, 10000)

func pushGateQueue() {
	redisClis := make(map[uint32]*redis.Client)
	buf := bytes.Buffer{}

	refreshGateQueue(redisClis)
	for {
		info := <-gateQueueInfoQ

		var cli *redis.Client
		var present bool
		cli, present = redisClis[info.gateid]
		if !present {
			refreshGateQueue(redisClis)
			cli, present = redisClis[info.gateid]
		}

		if cli == nil {
			log.Warn("can not find gate queue addr %d", info.gateid)
			continue
		}
		var bg codec.BackendGate
		bg.Connid = info.connid
		bg.MsgBuf = info.msgBuf
		bg.Encode(&buf)

		key := common.GenGateQueueKey(info.gateid)
		cli.RPush(key, hack.String(buf.Bytes()))
		log.Info("push msg to gate %d", info.gateid)

		buf.Reset()
	}
}

func refreshGateQueue(redisClis map[uint32]*redis.Client) {
	addrs := config.GetGateQueueAddrs()
	for gateid, addr := range addrs {
		if _, present := redisClis[gateid]; present {
			continue
		}

		redisClis[gateid] = redis.NewClient(&redis.Options{
			Addr:        addr,
			MaxRetries:  3,
			ReadTimeout: time.Millisecond * 1000,
			PoolSize:    1000,
			PoolTimeout: time.Millisecond * 300,
		})
	}
}

func sendToGateQ(gateid uint32, connid uint32, msgBuf []byte) {

	gateQueueInfoQ <- gateQueueInfo{gateid, connid, msgBuf}
}
