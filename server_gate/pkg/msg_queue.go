package pkg

import (
	"strings"
	"time"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/server_gate/config"
	"github.com/gochenzl/chess/util/log"
	"gopkg.in/redis.v3"
)

func processMsgQueue() {
	redisCli := redis.NewClient(&redis.Options{
		Addr:        config.GetQueueAddr(),
		MaxRetries:  3,
		PoolSize:    10,
		PoolTimeout: time.Millisecond * 300,
	})

	key := common.GenGateQueueKey(common.GetGateid())
	for {
		stringSliceCmd := redisCli.BLPop(0, key)
		if err := stringSliceCmd.Err(); err != nil {
			log.Error("%s", err.Error())
			continue
		}

		log.Info("pop")

		values := stringSliceCmd.Val()
		if len(values) != 2 {
			log.Error("length of values is %d", len(values))
			continue
		}

		var bg codec.BackendGate
		if err := bg.Decode(strings.NewReader(values[1])); err != nil {
			log.Error("decode BackendGate fail:%s", err.Error())
			continue
		}

		proccessBg(bg)
	}
}
