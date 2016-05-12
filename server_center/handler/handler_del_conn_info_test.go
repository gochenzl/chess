package handler

import (
	"bytes"
	"testing"

	"github.com/gochenzl/chess/pb/center"
	"github.com/gochenzl/chess/server_center/conn_info"
	"github.com/gochenzl/chess/util/rpc"
	"github.com/golang/protobuf/proto"
)

func TestHandleDelConnInfo(t *testing.T) {
	connInfo := center.ConnInfo{10000, 1, 1}
	conn_info.InitTest()
	conn_info.Add(connInfo)

	var clients []*bytes.Buffer
	clients = append(clients, &bytes.Buffer{})
	clients = append(clients, &bytes.Buffer{})

	for i := 0; i < len(clients); i++ {
		addClient(clients[i])
	}

	req := &center.DelConnInfoReq{1, 1}
	client := &bytes.Buffer{}
	addClient(client)
	HandleDelConnInfo(client, req)

	pb, err := rpc.DecodePb(client)
	if err != nil {
		t.Errorf("decode resp:%s", err.Error())
		return
	}
	if proto.MessageName(pb) != "center.DelConnInfoResp" {
		t.Errorf("invalid response:%s", proto.MessageName(pb))
	}

	for i := 0; i < len(clients); i++ {
		pb, err := rpc.DecodePb(clients[i])
		if err != nil {
			t.Errorf("decode resp:%s", err.Error())
			return
		}

		if proto.MessageName(pb) != "center.DelConnInfoNotify" {
			t.Errorf("invalid response:%s", proto.MessageName(pb))
		}

		resp := pb.(*center.DelConnInfoNotify)
		if *(resp.Info) != connInfo {
			t.Errorf("resp conn info")
		}
	}

	if conn_info.Exist(connInfo) {
		t.Errorf("del conn info fail")
	}

	req = &center.DelConnInfoReq{1, 1}
	HandleDelConnInfo(client, req)

	pb, err = rpc.DecodePb(client)
	if err != nil {
		t.Errorf("decode resp:%s", err.Error())
		return
	}
	if proto.MessageName(pb) != "center.DelConnInfoResp" {
		t.Errorf("invalid response:%s", proto.MessageName(pb))
	}

	for i := 0; i < len(clients); i++ {
		if clients[i].Len() != 0 {
			t.Errorf("duplicate del")
		}
	}
}
