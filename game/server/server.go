package server

import (
	"bufio"
	"bytes"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/game/session"
	"github.com/gochenzl/chess/util/buf_pool"
	"github.com/gochenzl/chess/util/log"
)

type respInfo struct {
	userid  uint32
	userids []uint32
	gc      codec.GameClient
}

var theConn net.Conn
var theConnMu sync.Mutex
var respQ chan respInfo = make(chan respInfo, 10000)

func Run(port int) error {
	for i := 0; i < 100; i++ {
		go workLoop()
		workerNum++
	}

	session.CheckStart()

	go monitorWorker()
	go writingLoop()
	go pushGateQueue()

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	log.Info("listen on port %d", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("accept fail:%s", err.Error())
			continue
		}

		setTheConn(conn)
		go handleConn(conn)
	}
}

func SendResp(userid uint32, msgid uint16, result uint16, msgBody []byte) {
	var gc codec.GameClient
	gc.Msgid = msgid
	gc.Result = result
	gc.MsgBody = msgBody

	respQ <- respInfo{userid: userid, gc: gc}
}

func LoginFail(connid uint32, userid uint32, msgid uint16, result uint16) {
	var gc codec.GameClient
	gc.Msgid = msgid
	gc.Result = result

	buf := buf_pool.Get()
	defer buf_pool.Put(buf)
	gc.Encode(buf)

	var bg codec.BackendGate
	bg.Connid = connid
	bg.MsgBuf = buf.Bytes()

	conn := getTheConn()
	if conn != nil {
		bg.Encode(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	br := bufio.NewReaderSize(conn, 10*1024*1024)
	for {
		var gb codec.GateBackend
		if err := gb.Decode(br); err != nil {
			log.Error("%s", err.Error())
			return
		}

		pushRequest(gb)

	}
}

func writingLoop() {
	buffer := bytes.NewBuffer(nil)
	pengingInfos := make([]respInfo, 0, 200)
	var bw *bufio.Writer
	var saveConn net.Conn

BEGIN:
	for {
		conn := getTheConn()
		if conn != nil && conn != saveConn {
			saveConn = conn
			bw = bufio.NewWriterSize(conn, 1024*1024)
			break
		}

		time.Sleep(time.Millisecond * 50)
	}

	if len(pengingInfos) > 0 {
		for i := 0; i < len(pengingInfos); i++ {
			bg := pengingInfos[i].encode(buffer)
			if bg.Connid == 0 && len(bg.Connids) == 0 {
				continue
			}

			if err := bg.Encode(bw); err != nil {
				goto BEGIN
			}

			buffer.Reset()
		}

		if err := bw.Flush(); err != nil {
			log.Error("%s", err.Error())
			goto BEGIN
		}
		pengingInfos = pengingInfos[:0]
	}

	for {
		info := <-respQ

		bg := info.encode(buffer)
		if bg.Connid == 0 && len(bg.Connids) == 0 {
			continue
		}

		pengingInfos = append(pengingInfos, info)
		if err := bg.Encode(bw); err != nil {
			goto BEGIN
		}

		buffer.Reset()

		if len(respQ) == 0 || len(pengingInfos) > 100 {
			if err := bw.Flush(); err != nil {
				log.Error("%s", err.Error())
				goto BEGIN
			} else {
				pengingInfos = pengingInfos[:0]
			}
		}
	}
}

func (info respInfo) encode(buffer *bytes.Buffer) codec.BackendGate {
	info.gc.Encode(buffer)

	var bg codec.BackendGate
	bg.MsgBuf = buffer.Bytes()

	if len(info.userids) > 0 {
		bg.Connids = make([]uint32, 0, len(info.userids))
		for i := 0; i < len(info.userids); i++ {
			sess, present := session.Get(info.userids[i])
			if !present {
				log.Warn("user %d has no session", info.userids[i])
				continue
			}
			if sess.Gateid != common.GetGateid() {
				sendToGateQ(sess.Gateid, sess.Connid, bg.MsgBuf)
			} else {
				bg.Connids = append(bg.Connids, sess.Connid)
			}

		}

	} else {
		sess, present := session.Get(info.userid)
		if present {
			if sess.Gateid != common.GetGateid() {
				sendToGateQ(sess.Gateid, sess.Connid, bg.MsgBuf)
			} else {
				bg.Connid = sess.Connid
			}
		} else {
			log.Warn("user %d has no session", info.userid)
		}
	}

	return bg
}

func getTheConn() net.Conn {
	theConnMu.Lock()
	defer theConnMu.Unlock()

	return theConn
}

func setTheConn(conn net.Conn) {
	if theConn != nil {
		theConn.Close()
	}

	theConnMu.Lock()
	theConn = conn
	theConnMu.Unlock()
}
