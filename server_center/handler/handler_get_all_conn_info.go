package handler

import (
	"io"

	"github.com/gochenzl/chess/pb/center"
	"github.com/gochenzl/chess/server_center/conn_info"
	"github.com/golang/protobuf/proto"
)

func HandleGetAllConnInfo(client io.Writer, req proto.Message) error {
	var getAllConnInfoResp center.GetAllConnInfoResp

	getAllConnInfoResp.Infos = conn_info.GetAll()

	err := sendResp(client, &getAllConnInfoResp)

	// 发送响应之后才加入，防止收到getAllConnInfoResp之前，收到其他的notify
	addClient(client)
	return err
}
