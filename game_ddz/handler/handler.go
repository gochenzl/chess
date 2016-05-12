package handler

import (
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/game/server"
	"github.com/golang/protobuf/proto"
)

// msgid
const (
	MsgidLoginReq  = 100
	MsgidLoginResp = 101

	MsgidEchoReq  = 102
	MsgidEchoResp = 103

	MsgidSendMessageReq  = 104
	MsgidSendMessageResp = 105

	MsgidMessageNotify = 800
)

// result code
const (
	ResultFailInvalidToken = 100 // 无效token
	ResultFailNotLogined   = 101 // 没有登录
	ResultFailUserNotExist = 102 // 玩家不存在
)

var msgidName map[uint16]string = make(map[uint16]string)
var resultName map[uint16]string = make(map[uint16]string)

func init() {
	server.RegisterHandler(MsgidLoginReq, HandleLogin)
	server.RegisterHandler(MsgidEchoReq, HandleEcho)
	server.RegisterHandler(MsgidSendMessageReq, HandleSendMessage)
	server.SetLoginReqMsgid(MsgidLoginReq)

	msgidName[MsgidLoginReq] = "LoginReq"
	msgidName[MsgidLoginResp] = "LoginResp"
	msgidName[MsgidEchoReq] = "EchoReq"
	msgidName[MsgidEchoResp] = "EchoResp"
	msgidName[MsgidSendMessageReq] = "SendMessageReq"
	msgidName[MsgidSendMessageResp] = "SendMessageResp"

	msgidName[MsgidMessageNotify] = "MessageNotify"

	resultName[common.ResultSuccess] = "Success"
	resultName[common.ResultFail] = "SystemFail"
	resultName[ResultFailInvalidToken] = "InvalidToken"
	resultName[ResultFailNotLogined] = "NotLogined"
	resultName[ResultFailUserNotExist] = "UserNotExist"

}

func MsgName(msgid uint16) string {
	name, present := msgidName[msgid]
	if present {
		return name
	}

	return "unknown"
}

func ResultName(result uint16) string {
	name, present := resultName[result]
	if present {
		return name
	}

	return "unknown"
}

func exitFunc(userid uint32, msgid uint16, result uint16, resp proto.Message) {

	if result != common.ResultSuccess || resp == nil {

		server.SendResp(userid, msgid, result, nil)
		return
	}

	buf, _ := proto.Marshal(resp)
	server.SendResp(userid, msgid, result, buf)

}
