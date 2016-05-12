package handler

import (
	"bytes"
	"io"
	"testing"

	"github.com/gochenzl/chess/pb/center"
	"github.com/gochenzl/chess/server_center/conn_info"
	"github.com/gochenzl/chess/util/rpc"
	"github.com/golang/protobuf/proto"
)

func TestHandleDelConnInfoByGateid(t *testing.T) {
	conn_info.InitTest()
	conn_info.Add(center.ConnInfo{10000, 1, 1})
	conn_info.Add(center.ConnInfo{20000, 1, 2})
	conn_info.Add(center.ConnInfo{30000, 2, 1})
	conn_info.Add(center.ConnInfo{40000, 2, 2})

	var clients []io.ReadWriter
	clients = append(clients, &bytes.Buffer{})
	clients = append(clients, &bytes.Buffer{})

	for i := 0; i < len(clients); i++ {
		addClient(clients[i])
	}

	req := &center.DelConnInfoByGateidReq{1}
	client := &bytes.Buffer{}
	addClient(client)
	HandleDelConnInfoByGateid(client, req)

	pb, err := rpc.DecodePb(client)
	if err != nil {
		t.Errorf("decode resp:%s", err.Error())
		return
	}
	if proto.MessageName(pb) != "center.DelConnInfoByGateidResp" {
		t.Errorf("invalid response:%s", proto.MessageName(pb))
	}

	for i := 0; i < len(clients); i++ {
		pb, err := rpc.DecodePb(clients[i])
		if err != nil {
			t.Errorf("decode resp:%s", err.Error())
			return
		}

		if proto.MessageName(pb) != "center.DelConnInfoByGateidNotify" {
			t.Errorf("invalid response:%s", proto.MessageName(pb))
		}

		notify := pb.(*center.DelConnInfoByGateidNotify)
		if notify.Gateid != 1 {
			t.Errorf("gateid")
		}
	}

	if conn_info.Exist(center.ConnInfo{10000, 1, 1}) ||
		conn_info.Exist(center.ConnInfo{10000, 1, 2}) {
		t.Errorf("del conn info by gateid fail")
	}
}
