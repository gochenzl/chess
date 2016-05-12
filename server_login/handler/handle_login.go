package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/pb/login"
	"github.com/gochenzl/chess/server_login/account"
	"github.com/gochenzl/chess/server_login/config"
	"github.com/gochenzl/chess/util/log"
	"github.com/gochenzl/chess/util/redis_cli"
	"github.com/golang/protobuf/proto"
	"github.com/satori/uuid"
)

func HandleLogin(w http.ResponseWriter, req *http.Request) {
	bodyBuf, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		log.Error("read request fail: %s", err.Error())
		return
	}

	data := codec.DecryptWithLen(bodyBuf)
	if data == nil {
		log.Error("invalid request: %v", bodyBuf)
		return
	}

	var loginReq login.LoginReq
	var loginResp login.LoginResp

	if err := proto.Unmarshal(data, &loginReq); err != nil {
		return
	}

	log.Info("receive login request:%s", loginReq.String())

	var info *account.AccountInfo
	if len(loginReq.Username) == 0 {
		info = account.New()
	} else {
		info = account.Query(loginReq.Username)
	}

	if info == nil {
		loginResp.Result = 1
	} else {
		loginResp.Userid = info.Userid
		loginResp.Username = info.Username
		loginResp.GameAddr = config.FindGameServer(int(loginReq.Version))
		loginResp.Token = genToken(info.Username)
	}

	setLoginInfo(&loginResp)

	respData, _ := proto.Marshal(&loginResp)
	w.Write(codec.EncryptWithLen(respData))
}

func genToken(username string) string {
	u := uuid.NewV4()
	data := u.Bytes()
	data = append(data, username...)

	var buf [md5.Size]byte
	buf = md5.Sum(data)
	return hex.EncodeToString(buf[:])
}

func setLoginInfo(loginResp *login.LoginResp) {
	var info common.LoginInfo
	info.Token = loginResp.Token

	data, _ := json.Marshal(&info)
	key := common.GenLoginInfoKey(loginResp.Userid)
	redis_cli.Set(key, string(data), time.Hour*48)
}
