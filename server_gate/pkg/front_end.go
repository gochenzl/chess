package pkg

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/server_gate/connid"
	"github.com/gochenzl/chess/util/log"
)

func checkTimeoutErr(err error) bool {
	if timeoutErr, ok := err.(net.Error); ok {
		if timeoutErr.Timeout() {
			return true
		}
	}

	return false
}

func sendBackendMsg(connid uint32, msgid uint16, msgBuf []byte) {
	var gb codec.GateBackend
	gb.Connid = connid
	gb.Msgid = msgid
	gb.MsgBuf = msgBuf
	sendToBackend(gb)
}

var connTimeout time.Duration = time.Minute * 5

func doFrontEnd(conn net.Conn) {
	log.Info("connection from %s", conn.RemoteAddr().String())
	defer conn.Close()

	id := connid.Get()
	if id == connid.InvalidId {
		log.Warn("connid exhaust")
		return
	}

	putConn(id, conn)

	defer connid.Release(id)
	defer delConn(id)
	defer sendBackendMsg(id, common.MsgDisconnect, nil)

	br := bufio.NewReader(conn)

	var idleSeconds time.Duration
	lenBuf := make([]byte, 4)
	for {
		if idleSeconds >= connTimeout {
			log.Info("connection %s timeout", conn.RemoteAddr().String())
			return
		}

		conn.SetDeadline(time.Now().Add(5 * time.Second))

		var err error
		if _, err = br.Read(lenBuf); err != nil {
			if checkTimeoutErr(err) {
				idleSeconds += 5 * time.Second
				continue
			}

			if err == io.EOF {
				log.Info("connection %s %s", conn.RemoteAddr().String(), err.Error())
			} else {
				log.Error("connection %s %s", conn.RemoteAddr().String(), err.Error())
			}

			return
		}

		totalSize := binary.LittleEndian.Uint32(lenBuf)

		if totalSize == 0 {
			idleSeconds = 0
			continue
		}

		if totalSize > 10*1024 {
			log.Warn("recv size %d", totalSize)
			return
		}

		conn.SetDeadline(time.Now().Add(10 * time.Second))
		msgBuf := make([]byte, totalSize)

		if _, err := io.ReadFull(br, msgBuf); err != nil {
			log.Error("ReadFull fail:%s, size=%d", err.Error(), totalSize)
			return
		}

		incRecvMsgCounter()
		sendBackendMsg(id, common.MsgRoute, msgBuf)
		idleSeconds = 0
	}
}
