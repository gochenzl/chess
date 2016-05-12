package session

import (
	"bufio"
	"net"
	"sync"
	"time"

	"github.com/gochenzl/chess/pb/center"
	"github.com/gochenzl/chess/util/log"
	"github.com/gochenzl/chess/util/rpc"
	"github.com/golang/protobuf/proto"
)

type Session struct {
	Gateid uint32
	Connid uint32
}

var sessionMu sync.RWMutex
var sessions map[uint32]Session = make(map[uint32]Session)
var startChan chan bool = make(chan bool)

func Init(centerHostAndPort string) {
	go centerClient(centerHostAndPort)
}

func CheckStart() {
	<-startChan
}

func centerClient(hostAndPort string) {
	var conn net.Conn
	var err error
	var closeChan bool

CREATE_CONN:
	printLog := true
	for {
		conn, err = net.Dial("tcp", hostAndPort)
		if err != nil {
			if printLog {
				log.Error("connect to center server fail: %s", err.Error())
				printLog = false
			}

			time.Sleep(time.Second * 1)
			continue
		}

		if err = sendGetAllConnInfoReq(conn); err != nil {
			log.Error("sendGetAllConnInfoReq fail: %s", err.Error())
			conn.Close()
			time.Sleep(time.Second * 1)
			continue
		}

		break
	}

	log.Info("connect to center server success")

	br := bufio.NewReader(conn)
	for {
		pbMsg, err := rpc.DecodePb(br)
		if err != nil {
			log.Error("recieve notify fail: %s", err.Error())
			conn.Close()
			goto CREATE_CONN
		}

		name := proto.MessageName(pbMsg)

		if name != "center.GetAllConnInfoResp" {
			log.Info("receive %s: %s", name, pbMsg.String())
		} else {
			log.Info("receive %s", name)
		}

		switch name {
		case "center.GetAllConnInfoResp":
			processGetAllConnInfoResp(pbMsg.(*center.GetAllConnInfoResp))
			if !closeChan {
				close(startChan)
				closeChan = true
			}

		case "center.NewConnInfoNotify":
			processNewConnInfoNotify(pbMsg.(*center.NewConnInfoNotify))
		case "center.DelConnInfoNotify":
			processDelConnInfoNotify(pbMsg.(*center.DelConnInfoNotify))
		case "center.DelConnInfoByGateidNotify":
			processDelConnInfoByGateidNotify(pbMsg.(*center.DelConnInfoByGateidNotify))
		}
	}
}

func sendGetAllConnInfoReq(conn net.Conn) error {
	var req center.GetAllConnInfoReq
	return rpc.EncodePb(conn, &req)
}

func processGetAllConnInfoResp(resp *center.GetAllConnInfoResp) {
	sessionMu.Lock()
	sessions = make(map[uint32]Session)
	for i := 0; i < len(resp.Infos); i++ {
		info := resp.Infos[i]
		sessions[info.Userid] = Session{Gateid: info.Gateid, Connid: info.Connid}
	}
	sessionMu.Unlock()
}

func processNewConnInfoNotify(notify *center.NewConnInfoNotify) {
	sessionMu.Lock()
	sessions[notify.Info.Userid] = Session{Gateid: notify.Info.Gateid, Connid: notify.Info.Connid}
	sessionMu.Unlock()
}

func processDelConnInfoNotify(notify *center.DelConnInfoNotify) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	info, present := sessions[notify.Info.Userid]
	if !present {
		return
	}

	if info.Gateid == notify.Info.Gateid && info.Connid == notify.Info.Connid {
		delete(sessions, notify.Info.Userid)
	}
}

func processDelConnInfoByGateidNotify(notify *center.DelConnInfoByGateidNotify) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	for userid, sess := range sessions {
		if sess.Gateid == notify.Gateid {
			delete(sessions, userid)
		}
	}
}

func Exist(userid uint32) bool {
	sessionMu.RLock()
	_, present := sessions[userid]
	sessionMu.RUnlock()

	return present
}

func Get(userid uint32) (Session, bool) {
	sessionMu.RLock()
	sess, present := sessions[userid]
	sessionMu.RUnlock()

	return sess, present
}

func Add(userid uint32, gateid uint32, connid uint32) {
	sessionMu.Lock()
	sessions[userid] = Session{Gateid: gateid, Connid: connid}
	sessionMu.Unlock()
}
