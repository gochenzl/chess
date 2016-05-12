package config

import (
	"sync"

	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/util/conf"
	"github.com/gochenzl/chess/util/log"
)

type GateQueueAddr struct {
	Gateid    uint32
	RedisAddr string
}

var gateQueueAddrs struct {
	Addrs []GateQueueAddr
	mu    *sync.RWMutex
}

func Init(confPath string) bool {
	if err := common.InitConfig(confPath + "/game.conf"); err != nil {
		log.Error("init common config fail:%s", err.Error())
		return false
	}

	confFile := confPath + "/gate_queue.csv"
	gateQueueAddrs.mu = conf.NewMutableConfig(confFile, conf.ConfigTypeCsv, &gateQueueAddrs)
	if gateQueueAddrs.mu == nil {
		return false
	}

	return true
}

func GetGateQueueAddrs() map[uint32]string {
	addrs := make(map[uint32]string)
	gateQueueAddrs.mu.RLock()
	defer gateQueueAddrs.mu.RUnlock()
	for _, addr := range gateQueueAddrs.Addrs {
		addrs[addr.Gateid] = addr.RedisAddr
	}

	return addrs
}
