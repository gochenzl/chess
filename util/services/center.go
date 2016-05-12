package services

import (
	"github.com/gochenzl/chess/pb/center"
	"github.com/gochenzl/chess/util/rpc"
)

func AddConnInfo(gateid uint32, connid uint32, userid uint32) bool {
	var req center.AddConnInfoReq
	var resp center.AddConnInfoResp

	req.Info = &center.ConnInfo{}
	req.Info.Connid = connid
	req.Info.Gateid = gateid
	req.Info.Userid = userid
	if !rpc.Invoke(Center, &req, &resp) {
		return false
	}

	return true
}

func DelConnInfo(gateid uint32, connid uint32) bool {
	var req center.DelConnInfoReq
	var resp center.DelConnInfoResp

	req.Gateid = gateid
	req.Connid = connid
	if !rpc.Invoke(Center, &req, &resp) {
		return false
	}

	return true
}

func DelConnInfoByGateid(gateid uint32) bool {
	var req center.DelConnInfoByGateidReq
	var resp center.DelConnInfoByGateidResp

	req.Gateid = gateid
	if !rpc.Invoke(Center, &req, &resp) {
		return false
	}

	return true
}
