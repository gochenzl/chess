package handler

import (
	"bytes"

	"github.com/gochenzl/chess/pb/center"
	"github.com/gochenzl/chess/server_center/conn_info"
	"github.com/gochenzl/chess/util/rpc"

	"testing"
)

func TestHandleGetAllConnInfo(t *testing.T) {
	conn_info.InitTest()
	var connInfos []center.ConnInfo
	connInfos = append(connInfos, center.ConnInfo{10000, 1, 1})
	connInfos = append(connInfos, center.ConnInfo{20000, 1, 2})
	connInfos = append(connInfos, center.ConnInfo{30000, 2, 1})
	connInfos = append(connInfos, center.ConnInfo{40000, 2, 2})

	for i := 0; i < len(connInfos); i++ {
		conn_info.Add(connInfos[i])
	}

	req := &center.GetAllConnInfoReq{}
	client := &bytes.Buffer{}
	HandleGetAllConnInfo(client, req)

	pb, err := rpc.DecodePb(client)
	if err != nil {
		t.Errorf("decode resp:%s", err.Error())
		return
	}
	resp := pb.(*center.GetAllConnInfoResp)
	if resp == nil {
		t.Errorf("invalid resp")
	}

	if len(resp.Infos) != len(connInfos) {
		t.Errorf("resp error %d", len(resp.Infos))
	}

	for i := 0; i < len(resp.Infos); i++ {
		var find bool
		for j := 0; j < len(connInfos); j++ {
			if *(resp.Infos[i]) == connInfos[j] {
				find = true
				break
			}
		}

		if !find {
			t.Errorf("resp error")
		}
	}
}
