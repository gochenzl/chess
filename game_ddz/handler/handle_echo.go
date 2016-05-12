package handler

import (
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/game/server"
)

func HandleEcho(userid uint32, connid uint32, msgBody []byte) {
	server.SendResp(userid, MsgidEchoResp, common.ResultSuccess, msgBody)
}
