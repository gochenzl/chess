package handler

import (
	"encoding/json"

	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/game/server"
	"github.com/gochenzl/chess/game/session"
	"github.com/gochenzl/chess/game_ddz/pb_client"
	"github.com/gochenzl/chess/game_ddz/user"
	"github.com/gochenzl/chess/util/log"
	"github.com/gochenzl/chess/util/redis_cli"
	"github.com/gochenzl/chess/util/services"
	"github.com/golang/protobuf/proto"
)

func HandleLogin(userid uint32, connid uint32, msgBody []byte) {
	var req pb_client.LoginReq
	var resp pb_client.LoginResp

	if err := proto.Unmarshal(msgBody, &req); err != nil {
		log.Warn("unmarshal LoginReq fail:%s", err.Error())
		return
	}

	log.Info("receive LoginReq:%s", req.String())

	var result uint16
	result = common.ResultFail

	defer func() {
		if result != common.ResultSuccess {
			server.LoginFail(connid, userid, MsgidLoginResp, result)
			return
		}

		exitFunc(userid, MsgidLoginResp, result, &resp)
	}()

	key := common.GenLoginInfoKey(userid)
	value, err := redis_cli.Get(key)
	if err != nil {
		if redis_cli.NullError(err) {
			result = ResultFailInvalidToken
		}
		return
	}

	var loginInfo common.LoginInfo
	if err := json.Unmarshal([]byte(value), &loginInfo); err != nil {
		result = ResultFailInvalidToken
		return
	}

	if req.Token != loginInfo.Token {
		result = ResultFailInvalidToken
		return
	}

	ui := user.LoadUserInfo(userid, user.AllUserFlags)
	if ui == nil {
		ui = user.NewUser(userid)
		ui.IncMoney(5000) // initial money
		ui.SetNickName(loginInfo.Nickname)
	}

	if !ui.Save() {
		return
	}

	services.AddConnInfo(common.GetGateid(), connid, userid)
	session.Add(userid, common.GetGateid(), connid)

	result = common.ResultSuccess
	resp.UserInfo = &pb_client.UserInfo{}
	resp.UserInfo.Userid = userid
	resp.UserInfo.Nickname = ui.NickName()
}
