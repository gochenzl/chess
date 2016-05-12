package pkg

import (
	"sync"
	"time"

	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/pb/table"
	"github.com/gochenzl/chess/util/redis_cli"
	"github.com/golang/protobuf/proto"
)

type roomInfo struct {
	roomid         int32
	numOfPlayers   int
	waitingUserids []uint32
	mu             sync.Mutex
}

var (
	tableInfos     map[int64]*tableInfo  // 桌子集合
	userid2Table   map[uint32]*tableInfo // 玩家所在桌子
	waitingUserids map[uint32]int32      // 等待配桌的所有玩家
	tableMu        sync.Mutex            // 保护上面所有变量

	roomInfos []*roomInfo
	roomMu    sync.RWMutex
)

func init() {
	tableInfos = make(map[int64]*tableInfo)
	userid2Table = make(map[uint32]*tableInfo)
	waitingUserids = make(map[uint32]int32)
	roomInfos = make([]*roomInfo, 0, 100)
}

func addRoomInfo(roomid int32, numOfPlayers int) bool {
	roomMu.Lock()
	defer roomMu.Unlock()

	for i := 0; i < len(roomInfos); i++ {
		if roomInfos[i].roomid == roomid {
			return false
		}
	}

	ri := &roomInfo{roomid: roomid, numOfPlayers: numOfPlayers}
	roomInfos = append(roomInfos, ri)
	go ri.match()

	return true
}

func (ri *roomInfo) add(userid uint32) bool {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	for i := 0; i < len(ri.waitingUserids); i++ {
		if ri.waitingUserids[i] == userid {
			return false
		}
	}

	ri.waitingUserids = append(ri.waitingUserids, userid)
	return true
}

func (ri *roomInfo) inWaiting(userid uint32) bool {
	ri.mu.Lock()
	defer ri.mu.Unlock()

	for i := 0; i < len(ri.waitingUserids); i++ {
		if ri.waitingUserids[i] == userid {
			return true
		}
	}

	return false
}

func (ri *roomInfo) match() {

	tis := make([]*tableInfo, 0, 100)

	for {
		ri.mu.Lock()
		size := len(ri.waitingUserids)
		if size < ri.numOfPlayers {
			ri.mu.Unlock()
			time.Sleep(time.Millisecond * 200)
			continue
		}

		index := 0
		for size >= ri.numOfPlayers {
			userids := make([]uint32, ri.numOfPlayers)
			copy(userids, ri.waitingUserids[index:index+ri.numOfPlayers])

			ti := newTableInfo(ri.roomid, userids)
			tis = append(tis, ti)

			index += ri.numOfPlayers
			size -= ri.numOfPlayers
		}
		copy(ri.waitingUserids, ri.waitingUserids[index:])
		ri.waitingUserids = ri.waitingUserids[:size]
		ri.mu.Unlock()

		addTables(tis)

		for i := 0; i < len(tis); i++ {
			pbBuf, _ := proto.Marshal(tis[i].savepb())
			redis_cli.RPush(common.TableNewList, string(pbBuf))
		}

		tis = tis[:0]

		time.Sleep(time.Millisecond * 200)

	}
}

func enterRoom(req *table.EnterRoomReq) (resp *table.EnterRoomResp) {
	resp = &table.EnterRoomResp{}

	ri := getRoomInfo(req.Roomid)
	if ri == nil {
		resp.Result = common.ResultFailRoomNotExist
		return
	}

	tableMu.Lock()

	// 正在游戏中
	if ti, present := userid2Table[req.Userid]; present {
		tableMu.Unlock()

		if ti.roomid != req.Roomid {
			resp.Result = common.ResultFailInGame
			resp.Roomid = ti.roomid
			resp.Tableid = ti.id
			return
		}

		resp.TableInfo = ti.savepb()
		return
	}

	// 已经在房间等待
	if waitingRoomid, present := waitingUserids[req.Userid]; present {
		tableMu.Unlock()

		resp.Result = common.ResultFailInWaiting
		resp.Roomid = waitingRoomid
		return
	}
	waitingUserids[req.Userid] = req.Roomid

	tableMu.Unlock()

	// 已经在本房间等待
	if !ri.add(req.Userid) {
		resp.Result = common.ResultFailInWaiting
		resp.Roomid = req.Roomid
		return
	}

	return
}

func exitRoom(req *table.ExitRoomReq) (resp *table.ExitRoomResp) {
	resp = &table.ExitRoomResp{}

	tableMu.Lock()

	// 正在游戏中，不能退出房间
	if ti, present := userid2Table[req.Userid]; present {
		tableMu.Unlock()

		resp.Result = common.ResultFailInGame
		resp.Tableid = ti.id
		return
	}

	var waitingRoomid int32
	var present bool
	if waitingRoomid, present = waitingUserids[req.Userid]; !present {
		tableMu.Unlock()
		return
	}

	delete(waitingUserids, req.Userid)
	tableMu.Unlock()

	ri := getRoomInfo(waitingRoomid)
	if ri == nil {
		return
	}

	ri.mu.Lock()
	for i := 0; i < len(ri.waitingUserids); i++ {
		if ri.waitingUserids[i] == req.Userid {
			size := len(ri.waitingUserids)
			ri.waitingUserids[i] = ri.waitingUserids[size-1]
			ri.waitingUserids = ri.waitingUserids[:size-1]
		}
	}

	ri.mu.Unlock()
	return
}

func updateTableInfo(req *table.UpdateTableInfoReq) (resp *table.UpdateTableInfoResp) {
	resp = &table.UpdateTableInfoResp{}

	ti := queryById(req.Id)
	if ti == nil {
		resp.Result = common.ResultFailTableNotExist
		return
	}

	if !ti.update(req) {
		resp.Result = common.ResultFailTableConflict
		resp.TableInfo = ti.savepb()
		return
	}

	if req.GameOver {
		delTable(ti.id, ti.userids)
	}

	return
}

func getRoomInfo(roomid int32) *roomInfo {
	roomMu.RLock()
	defer roomMu.RUnlock()

	for i := 0; i < len(roomInfos); i++ {
		if roomInfos[i].roomid == roomid {
			return roomInfos[i]
		}
	}

	return nil
}

func queryWaiting(userid uint32) (waitingRoomid int32) {
	var present bool
	tableMu.Lock()
	defer tableMu.Unlock()

	if waitingRoomid, present = waitingUserids[userid]; !present {
		return
	}

	ri := getRoomInfo(waitingRoomid)
	if ri == nil {
		waitingRoomid = 0
		return
	}

	if ri.inWaiting(userid) {
		return
	}

	waitingRoomid = 0
	return
}

func queryById(id int64) *tableInfo {
	tableMu.Lock()
	ti, _ := tableInfos[id]
	tableMu.Unlock()
	return ti
}

func queryByUserid(userid uint32) *tableInfo {
	tableMu.Lock()
	defer tableMu.Unlock()
	if ti, present := userid2Table[userid]; present {
		return ti
	}

	return nil
}

func addTables(tis []*tableInfo) {
	tableMu.Lock()

	for i := 0; i < len(tis); i++ {
		ti := tis[i]

		tableInfos[ti.id] = ti

		for j := 0; j < len(ti.userids); j++ {
			userid2Table[ti.userids[j]] = ti
			delete(waitingUserids, ti.userids[j])
		}
	}

	tableMu.Unlock()
}

func delTable(id int64, userids []uint32) {
	tableMu.Lock()
	defer tableMu.Unlock()

	delete(tableInfos, id)
	for i := 0; i < len(userids); i++ {
		delete(userid2Table, userids[i])
	}
}
