package pkg

import (
	"bufio"
	"net"

	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/pb/table"
	"github.com/gochenzl/chess/util/log"
	"github.com/gochenzl/chess/util/rpc"
	"github.com/golang/protobuf/proto"
)

func handleEnterRoom(req proto.Message) proto.Message {
	enterRoomReq, ok := req.(*table.EnterRoomReq)
	if !ok {
		return nil
	}

	resp := enterRoom(enterRoomReq)
	return resp
}

func handleExitRoom(req proto.Message) proto.Message {
	exitRoomReq, ok := req.(*table.ExitRoomReq)
	if !ok {
		return nil
	}

	resp := exitRoom(exitRoomReq)
	return resp
}

func handleUpdateTableInfo(req proto.Message) proto.Message {
	updateTableInfoReq, ok := req.(*table.UpdateTableInfoReq)
	if !ok {
		return nil
	}

	resp := updateTableInfo(updateTableInfoReq)
	return resp
}

func handleQueryTableInfo(req proto.Message) proto.Message {
	queryTableInfoReq, ok := req.(*table.QueryTableInfoReq)
	if !ok {
		return nil
	}

	resp := &table.QueryTableInfoResp{}

	ti := queryById(queryTableInfoReq.Id)
	if ti == nil {
		resp.Result = common.ResultFailTableNotExist
		return resp
	}

	resp.TableInfo = ti.savepb()
	return resp
}

func handleQueryByUserid(req proto.Message) proto.Message {
	queryByUseridReq, ok := req.(*table.QueryByUseridReq)
	if !ok {
		return nil
	}

	resp := &table.QueryByUseridResp{}

	ti := queryByUserid(queryByUseridReq.Userid)
	if ti == nil {
		return resp
	}

	resp.Roomid = ti.roomid
	resp.Tableid = ti.id
	return resp
}

func HandleConn(conn net.Conn) {
	defer conn.Close()

	log.Info("new connection from %s", conn.RemoteAddr().String())
	br := bufio.NewReader(conn)
	bw := bufio.NewWriter(conn)

	for {
		req, err := rpc.DecodePb(br)
		if err != nil {
			log.Error("connection from %s error: %s", conn.RemoteAddr().String(), err.Error())
			return
		}

		name := proto.MessageName(req)
		var resp proto.Message
		switch name {
		case "table.EnterRoomReq":
			resp = handleEnterRoom(req)
		case "table.ExitRoomReq":
			resp = handleExitRoom(req)
		case "table.UpdateTableInfoReq":
			resp = handleUpdateTableInfo(req)
		case "table.QueryTableInfoReq":
			resp = handleQueryTableInfo(req)
		case "table.QueryByUseridReq":
			resp = handleQueryByUserid(req)

		default:
			log.Warn("invalid request:%s", name)
		}

		if resp == nil {
			continue
		}

		if err := rpc.EncodePb(bw, resp); err != nil {
			log.Error("encode resp fail:%s", err.Error())
			return
		}

		if err := bw.Flush(); err != nil {
			log.Error("flush fail:%s", err.Error())
			return
		}
	}
}
