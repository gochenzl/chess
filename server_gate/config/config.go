package config

import (
	"sync"

	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/util/conf"
	"github.com/gochenzl/chess/util/log"
)

var gateConfig struct {
	QueueAddr string `ini:"queue_addr"`
}

var backendConfig struct {
	List string `ini:"list"`
	mu   *sync.RWMutex
}

func GetQueueAddr() string {
	return gateConfig.QueueAddr
}

func GetBackendConfig() string {
	backendConfig.mu.RLock()
	defer backendConfig.mu.RUnlock()

	return backendConfig.List
}

func Init(confPath string) bool {
	if err := common.InitConfig(confPath + "/gate.conf"); err != nil {
		log.Error("init common config fail:%s", err.Error())
		return false
	}

	if err := conf.LoadIniFromFile(confPath+"/gate.conf", &gateConfig); err != nil {
		log.Error("init gate config fail:%s", err.Error())
		return false
	}

	backendConfig.mu = conf.NewMutableConfig(confPath+"/backend.conf", conf.ConfigTypeIni, &backendConfig)
	if backendConfig.mu == nil {
		return false
	}

	return true
}
