package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/server_login/account"
	"github.com/gochenzl/chess/server_login/config"
	"github.com/gochenzl/chess/server_login/handler"
	"github.com/gochenzl/chess/util/log"
	"github.com/gochenzl/chess/util/redis_cli"
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

	// change it
	key := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
		0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f}
	iv := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f}
	codec.Init(key, iv)

	// for test
	account.Init()

	mux := http.NewServeMux()
	mux.HandleFunc("/login", handler.HandleLogin)

	redis_cli.Init(common.GetRedisAddr(), 100)

	var server http.Server
	server.Addr = ":" + strconv.Itoa(common.GetListenPort())
	server.ReadTimeout = time.Second * 5
	server.WriteTimeout = time.Second * 8
	server.Handler = mux

	if err := server.ListenAndServe(); err != nil {
		log.Error("ListenAndServe fail: %s", err.Error())
		return
	}
}
