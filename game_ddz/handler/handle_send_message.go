package handler

import (
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/game/server"
	"github.com/gochenzl/chess/game/session"
	"github.com/gochenzl/chess/game_ddz/pb_client"
	"github.com/gochenzl/chess/game_ddz/user"
	"github.com/gochenzl/chess/util/log"
	"github.com/golang/protobuf/proto"
)

func HandleSendMessage(userid uint32, connid uint32, msgBody []byte) {
	var req pb_client.SendMessageReq
	var resp pb_client.SendMessageResp

	if err := proto.Unmarshal(msgBody, &req); err != nil {
		log.Warn("unmarshal SendMessageReq fail:%s", err.Error())
		return
	}

	log.Info("receive SendMessageReq:%s", req.String())

	var result uint16
	result = common.ResultFail

	defer func() {
		exitFunc(userid, MsgidSendMessageResp, result, &resp)
	}()

	ui := user.LoadUserInfo(req.Receiver, []int{user.FlagBasicInfo})
	if ui == nil {
		result = ResultFailUserNotExist
		return
	}

	if !session.Exist(req.Receiver) {
		result = ResultFailNotLogined
		return
	}

	var notify pb_client.MessageNotify
	notify.Sender = userid
	notify.Content = req.Content

	buf, _ := proto.Marshal(&notify)
	server.SendResp(req.Receiver, MsgidMessageNotify, 0, buf)

	result = common.ResultSuccess
}
