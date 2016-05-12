package pkg

import (
	"sync"
	"time"

	"github.com/gochenzl/chess/util/conf"
	"github.com/gochenzl/chess/util/log"
)

type roomConfig struct {
	Roomid       int32
	NumOfPlayers int
}

var roomConfigs struct {
	Configs []roomConfig
	mu      *sync.RWMutex
}

func InitRoomConfig(confFile string) bool {
	roomConfigs.mu = conf.NewMutableConfig(confFile, conf.ConfigTypeCsv, &roomConfigs)
	if roomConfigs.mu == nil {
		return false
	}

	go refreshRoomConfig()

	return true
}

func refreshRoomConfig() {
	roomConfigMap := make(map[roomConfig]bool)

	for {
		roomConfigs.mu.RLock()
		for i := 0; i < len(roomConfigs.Configs); i++ {
			c := roomConfigs.Configs[i]
			if _, present := roomConfigMap[c]; present {
				continue
			}

			if addRoomInfo(c.Roomid, c.NumOfPlayers) {
				log.Info("new room %d %d", c.Roomid, c.NumOfPlayers)
			}
		}
		roomConfigs.mu.RUnlock()

		time.Sleep(time.Second * 10)
	}
}
