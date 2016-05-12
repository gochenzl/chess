package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/golang/protobuf/proto"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	ddz_handler "github.com/gochenzl/chess/game_ddz/handler"
	ddz_pb_client "github.com/gochenzl/chess/game_ddz/pb_client"
	"github.com/gochenzl/chess/pb/login"
	"github.com/gochenzl/chess/util/log"
)

const accountLoginUrl = "http://127.0.0.1:9090/login"

func loginAccount() (loginResp login.LoginResp, success bool) {
	var req login.LoginReq
	req.Version = 1

	data, _ := proto.Marshal(&req)

	resp, err := http.Post(accountLoginUrl, "", bytes.NewReader(codec.EncryptWithLen(data)))
	if err != nil {
		log.Error("%s", err.Error())
		return
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("%s", err.Error())
		return
	}

	if err := proto.Unmarshal(codec.DecryptWithLen(buf), &loginResp); err != nil {
		log.Error("%s", err.Error())
		return
	}

	log.Info("account login resp:%s", loginResp.String())

	success = true
	return
}

func loginGame(brw *bufio.ReadWriter, userid uint32, token string) {
	var req ddz_pb_client.LoginReq
	var resp ddz_pb_client.LoginResp

	req.Token = token

	var cg codec.ClientGame
	cg.Userid = userid
	cg.Msgid = ddz_handler.MsgidLoginReq
	cg.MsgBody, _ = proto.Marshal(&req)

	if err := cg.Encode(brw); err != nil {
		log.Error("login game fail:%s", err.Error())
		return
	}

	if err := brw.Flush(); err != nil {
		log.Error("login game fail:%s", err.Error())
		return
	}

	var gc codec.GameClient
	if err := gc.DecodeFromReader(brw); err != nil {
		log.Error("receive login game resp fail:%s", err.Error())
		return
	}

	if gc.Result != common.ResultSuccess {
		log.Error("login game result = %d", gc.Result)
		return
	}

	proto.Unmarshal(gc.MsgBody, &resp)
	log.Info("receive login game resp:%s", resp.String())
}

func echo(brw *bufio.ReadWriter, userid uint32) bool {
	var cg codec.ClientGame
	cg.Userid = userid
	cg.Msgid = ddz_handler.MsgidEchoReq
	cg.MsgBody = []byte("hello world")

	if err := cg.Encode(brw); err != nil {
		log.Error("echo fail:%s", err.Error())
		return false
	}

	if err := brw.Flush(); err != nil {
		log.Error("login game fail:%s", err.Error())
		return false
	}

	var gc codec.GameClient
	if err := gc.DecodeFromReader(brw); err != nil {
		log.Error("receive echo resp fail:%s", err.Error())
		return false
	}

	log.Info("receive echo resp: %s", gc.MsgBody)
	return true
}

func sendMessage(conn net.Conn, userid uint32, receiver uint32) {
	var req ddz_pb_client.SendMessageReq
	var resp ddz_pb_client.SendMessageResp

	req.Receiver = receiver
	req.Content = "hello world"

	var cg codec.ClientGame
	cg.Userid = userid
	cg.Msgid = ddz_handler.MsgidSendMessageReq
	cg.MsgBody, _ = proto.Marshal(&req)

	if err := cg.Encode(conn); err != nil {
		log.Error("send message fail:%s", err.Error())
		return
	}

	var gc codec.GameClient
	if err := gc.DecodeFromReader(conn); err != nil {
		log.Error("receive send message resp fail:%s", err.Error())
		return
	}

	if gc.Result != common.ResultSuccess {
		log.Error("login game result = %s", ddz_handler.ResultName(gc.Result))
		return
	}

	proto.Unmarshal(gc.MsgBody, &resp)
	log.Info("receive send message resp:%s", resp.String())
}

func recv(conn net.Conn) {
	for {
		var gc codec.GameClient
		if err := gc.DecodeFromReader(conn); err != nil {
			log.Error("receive echo resp fail:%s", err.Error())
			return
		}

		switch gc.Msgid {
		case ddz_handler.MsgidMessageNotify:
			var notify ddz_pb_client.MessageNotify
			proto.Unmarshal(gc.MsgBody, &notify)
			log.Info("receive MessageNotify: %s", notify.String())
		}
	}
}

func main() {
	key := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
		0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f}
	iv := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f}
	codec.Init(key, iv)

	loginResp, success := loginAccount()
	if !success {
		return
	}

	conn, err := net.Dial("tcp", loginResp.GameAddr)
	if err != nil {
		log.Error("%s", err.Error())
		return
	}
	defer conn.Close()

	brw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	loginGame(brw, loginResp.Userid, loginResp.Token)
	echo(brw, loginResp.Userid)
}
