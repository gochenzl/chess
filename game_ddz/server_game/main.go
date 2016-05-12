package main

import (
	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/game/config"
	"github.com/gochenzl/chess/game/server"
	"github.com/gochenzl/chess/game/session"
	_ "github.com/gochenzl/chess/game_ddz/handler"
	"github.com/gochenzl/chess/game_ddz/user"
	"github.com/gochenzl/chess/util/log"
	"github.com/gochenzl/chess/util/redis_cli"
	"github.com/gochenzl/chess/util/rpc"
	"github.com/gochenzl/chess/util/services"

	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s conf_path\n", os.Args[0])
		return
	}

	log.Info("server start, pid = %d", os.Getpid())

	if !config.Init(os.Args[1]) {
		return
	}

	if !user.Init(common.GetUserAddr()) {
		return
	}

	if !redis_cli.Init(common.GetRedisAddr(), 500) {
		return
	}

	// change it
	key := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
		0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f}
	iv := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f}
	codec.Init(key, iv)

	rpc.Add(services.Center, common.GetCenterAddr(), 100)
	rpc.Add(services.Table, common.GetTableAddr(), 1000)

	session.Init(common.GetCenterAddr())
	server.Run(common.GetListenPort())

}
