package services

import (
	"github.com/gochenzl/chess/pb/table"
	"github.com/gochenzl/chess/util/rpc"
)

func EnterRoom(roomid int32, userid uint32) *table.EnterRoomResp {
	var req table.EnterRoomReq
	var resp table.EnterRoomResp

	req.Roomid = roomid
	req.Userid = userid

	if !rpc.Invoke(Table, &req, &resp) {
		return nil
	}

	return &resp
}

func QueryTableInfo(tableid int64) *table.QueryTableInfoResp {
	var req table.QueryTableInfoReq
	var resp table.QueryTableInfoResp

	req.Id = tableid

	if !rpc.Invoke(Table, &req, &resp) {
		return nil
	}

	return &resp
}

func UpdateTableInfo(req *table.UpdateTableInfoReq) *table.UpdateTableInfoResp {
	var resp table.UpdateTableInfoResp
	if !rpc.Invoke(Table, req, &resp) {
		return nil
	}

	return &resp
}
