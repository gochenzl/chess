package server

import (
	"encoding/binary"
	"time"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/game/session"
	"github.com/gochenzl/chess/util/log"
	"github.com/gochenzl/chess/util/services"
)

type handleFunc func(userid uint32, connid uint32, msgBody []byte)

var requestQ chan codec.GateBackend = make(chan codec.GateBackend, 10000)
var handlers map[uint16]handleFunc = make(map[uint16]handleFunc)
var workerNum int
var loginReqMsgid uint16

func pushRequest(gb codec.GateBackend) {
	requestQ <- gb
}

func workLoop() {
	for {
		gb := <-requestQ

		if gb.Msgid == common.MsgRoute {
			var cg codec.ClientGame
			if err := cg.Decode(gb.MsgBuf); err != nil {
				log.Warn("decode client game msg fail:%s", err.Error())
				continue
			}

			f, ok := handlers[cg.Msgid]
			if !ok {
				log.Warn("find %d handler fail", cg.Msgid)
				continue
			}

			if cg.Msgid != loginReqMsgid && !session.Exist(cg.Userid) {
				log.Info("user %d has not yet logined", cg.Userid)
				continue
			}

			f(cg.Userid, gb.Connid, cg.MsgBody)
		} else if gb.Msgid == common.MsgGateid {
			if len(gb.MsgBuf) == 4 {
				id := binary.LittleEndian.Uint32(gb.MsgBuf)
				common.SetGateid(id)

				log.Info("recv gateid:%d", id)
			}

		} else if gb.Msgid == common.MsgDisconnect {
			services.DelConnInfo(common.GetGateid(), gb.Connid)
			log.Info("recv disconnect:%d", gb.Connid)
		}
	}
}

func monitorWorker() {
	t := time.Second

	for {
		time.Sleep(t)

		qlen := len(requestQ)

		if qlen > 10 {
			go workLoop()
			workerNum++
			log.Warn("add work routine, workerNum = %d, queueLen = %d", workerNum, qlen)
			t = time.Millisecond * 10
		} else {
			t = time.Second
		}

		if workerNum > 10000 {
			log.Warn("monitorWorker exit")
			return
		}
	}
}

func RegisterHandler(msgid uint16, f handleFunc) {
	handlers[msgid] = f
}

func SetLoginReqMsgid(msgid uint16) {
	loginReqMsgid = msgid
}
