package config

import (
	"sync"

	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/util/conf"
	"github.com/gochenzl/chess/util/log"
)

type gameServer struct {
	index         int
	ClientVersion int      `json:"client_version"`
	Addrs         []string `json:"addrs"`
}

var gameServerGroup struct {
	Servers []gameServer `json:"servers"`
	mu      *sync.RWMutex
}

func Init(confPath string) bool {
	if err := common.InitConfig(confPath + "/login.conf"); err != nil {
		log.Error("init common config fail")
		return false
	}

	confFile := confPath + "/game_server_group.json"
	gameServerGroup.mu = conf.NewMutableConfig(confFile, conf.ConfigTypeJson, &gameServerGroup)
	if gameServerGroup.mu == nil {
		return false
	}

	return true
}

func FindGameServer(clientVersion int) string {
	gameServerGroup.mu.RLock()
	defer gameServerGroup.mu.RUnlock()

	if len(gameServerGroup.Servers) == 0 {
		return ""
	}

	var server *gameServer
	for i := 0; i < len(gameServerGroup.Servers); i++ {

		if gameServerGroup.Servers[i].ClientVersion == clientVersion {
			server = &(gameServerGroup.Servers[i])
			break
		}
	}

	if server == nil {
		server = &(gameServerGroup.Servers[0])
	}

	if len(server.Addrs) == 0 {
		return ""
	}

	index := server.index % len(server.Addrs)
	server.index = (server.index + 1) % len(server.Addrs)

	return server.Addrs[index]

}
