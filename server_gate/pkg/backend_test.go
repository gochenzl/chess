package pkg

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
	"testing"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
)

func backendServer(port int) net.Conn {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		return conn
	}
}

func TestBackend(t *testing.T) {
	common.SetGateid(100)
	go DoBackend("127.0.0.1:9876")
	conn := backendServer(9876)

	var gb codec.GateBackend
	if err := gb.Decode(conn); err != nil {
		t.Errorf("decode GateBackend:%s", err.Error())
	}

	if gb.Msgid != common.MsgGateid {
		t.Errorf("expected MsgGateid")
	}

	recvGateid := int(binary.LittleEndian.Uint16(gb.MsgBuf))
	if recvGateid != 100 {
		t.Errorf("gateid:%d", recvGateid)
	}

	gb.Msgid = common.MsgRoute
	gb.MsgBuf = []byte("hello world")
	sendToBackend(gb)

	if err := gb.Decode(conn); err != nil {
		t.Errorf("decode GateBackend:%s", err.Error())
	}

	if gb.Msgid != common.MsgRoute {
		t.Errorf("expected MsgGateid")
	}

	if bytes.Compare([]byte("hello world"), gb.MsgBuf) != 0 {
		t.Errorf("recv")
	}
}
