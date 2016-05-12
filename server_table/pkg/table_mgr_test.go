package pkg

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/pb/table"
	"github.com/gochenzl/chess/util/redis"
	"github.com/gochenzl/chess/util/redis_cli"
	"github.com/golang/protobuf/proto"
)

func TestMain(m *testing.M) {
	flag.Parse()

	ri := &roomInfo{roomid: 1, numOfPlayers: 3}
	roomInfos = append(roomInfos, ri)
	go ri.match()

	ri = &roomInfo{roomid: 2, numOfPlayers: 4}
	roomInfos = append(roomInfos, ri)
	go ri.match()

	server := redis.NewServer("127.0.0.1:12345", redis.NewMemoryStore())
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run server fail:%s", err.Error())
		return
	}

	redis_cli.Init("127.0.0.1:12345", 10)

	ret := m.Run()
	server.Close()
	os.Exit(ret)
}

func TestEnterRoom(t *testing.T) {
	var resp *table.EnterRoomResp
	req := &table.EnterRoomReq{}

	req.Userid = 1
	req.Roomid = 100

	resp = enterRoom(req)
	if resp.Result != common.ResultFailRoomNotExist {
		t.Errorf("enterRoom:%d", resp.Result)
		return
	}

	req.Userid = 1
	req.Roomid = 1

	resp = enterRoom(req)
	if resp.Result != common.ResultSuccess {
		t.Errorf("enterRoom:%d", resp.Result)
		return
	}

	if queryWaiting(1) != 1 {
		t.Errorf("enterRoom")
	}

	resp = enterRoom(req)
	if resp.Result != common.ResultFailInWaiting {
		t.Errorf("enterRoom:%d", resp.Result)
		return
	}

	req.Userid = 2
	resp = enterRoom(req)
	if resp.Result != common.ResultSuccess {
		t.Errorf("enterRoom:%d", resp.Result)
		return
	}

	req.Userid = 3
	resp = enterRoom(req)
	if resp.Result != common.ResultSuccess {
		t.Errorf("enterRoom:%d", resp.Result)
		return
	}

	time.Sleep(time.Millisecond * 800)
	str, err := redis_cli.LPop(common.TableNewList)
	if err != nil {
		t.Errorf("enterRoom:%s", err.Error())
		return
	}

	var pbTable table.TableInfo
	if err := proto.Unmarshal([]byte(str), &pbTable); err != nil {
		t.Errorf("enterRoom:%s", err.Error())
		return
	}

	resp = enterRoom(req)
	if resp.Result != common.ResultSuccess {
		t.Errorf("enterRoom:%d", resp.Result)
		return
	}
	if resp.TableInfo == nil {
		t.Errorf("enterRoom")
		return
	}

	req.Roomid = 2
	resp = enterRoom(req)
	if resp.Result != common.ResultFailInGame || resp.Tableid != pbTable.Id {
		t.Errorf("enterRoom:%d, %d", resp.Result, resp.Tableid)
		return
	}
}

func TestExitRoom(t *testing.T) {
	enterRoomReq := &table.EnterRoomReq{}
	enterRoomReq.Roomid = 1
	enterRoomReq.Userid = 10
	enterRoom(enterRoomReq)

	var exitRoomResp *table.ExitRoomResp
	exitRoomReq := &table.ExitRoomReq{}
	exitRoomReq.Userid = 10
	exitRoomResp = exitRoom(exitRoomReq)
	if exitRoomResp.Result != common.ResultSuccess {
		t.Errorf("exitRoom:%d", exitRoomResp.Result)
	}

	exitRoomResp = exitRoom(exitRoomReq)
	if exitRoomResp.Result != common.ResultSuccess {
		t.Errorf("exitRoom:%d", exitRoomResp.Result)
	}

	enterRoomReq.Userid = 10
	enterRoom(enterRoomReq)
	enterRoomReq.Userid = 11
	enterRoom(enterRoomReq)
	enterRoomReq.Userid = 12
	enterRoom(enterRoomReq)

	time.Sleep(time.Millisecond * 80)

	redis_cli.LPop(common.TableNewList)

	exitRoomResp = exitRoom(exitRoomReq)
	if exitRoomResp.Result != common.ResultFailInGame {
		t.Errorf("exitRoom:%d", exitRoomResp.Result)
	}
}

func TestUpdateTableInfo(t *testing.T) {
	enterRoomReq := &table.EnterRoomReq{}
	enterRoomReq.Roomid = 1

	enterRoomReq.Userid = 20
	enterRoom(enterRoomReq)
	enterRoomReq.Userid = 21
	enterRoom(enterRoomReq)
	enterRoomReq.Userid = 22
	enterRoom(enterRoomReq)

	time.Sleep(time.Millisecond * 800)
	str, err := redis_cli.LPop(common.TableNewList)
	if err != nil {
		t.Errorf("updateTableInfo:%s", err.Error())
		return
	}

	var pbTable table.TableInfo
	if err := proto.Unmarshal([]byte(str), &pbTable); err != nil {
		t.Errorf("updateTableInfo:%s", err.Error())
		return
	}

	var resp *table.UpdateTableInfoResp
	req := &table.UpdateTableInfoReq{}
	req.Id = pbTable.Id
	req.Version = pbTable.Version
	req.GameInfo = []byte("hello")
	resp = updateTableInfo(req)
	if resp.Result != common.ResultSuccess {
		t.Errorf("updateTableInfo:%d", resp.Result)
	}

	ti := queryById(pbTable.Id)
	if ti.version != req.Version+1 {
		t.Errorf("updateTableInfo, version:%d", ti.version)
	}
	if string(ti.gameInfo) != "hello" {
		t.Errorf("updateTableInfo, gameInfo:%s", string(ti.gameInfo))
	}

	req.TimerInfo = &table.TimerInfo{Duration: 1}
	req.Version++
	updateTableInfo(req)
	if ti.timer == nil {
		t.Errorf("updateTableInfo, set timer")
	}

	time.Sleep(time.Millisecond * 1050)
	str, err = redis_cli.LPop(common.TableTimeoutList)
	if err != nil {
		t.Errorf("updateTableInfo:%s", err.Error())
		return
	}
	if str != strconv.Itoa(int(ti.id)) {
		t.Errorf("updateTableInfo, timeout")
		return
	}

	saveVersion := req.Version
	req.Version--
	resp = updateTableInfo(req)
	if resp.Result != common.ResultFailTableConflict {
		t.Errorf("updateTableInfo:%d", resp.Result)
	}
	if resp.TableInfo == nil {
		t.Errorf("updateTableInfo, conflict")
	}

	req.Version = saveVersion + 1
	req.GameOver = true
	updateTableInfo(req)
	if queryById(pbTable.Id) != nil {
		t.Errorf("updateTableInfo, gameover")
	}
	if queryByUserid(20) != nil ||
		queryByUserid(21) != nil ||
		queryByUserid(22) != nil {
		t.Errorf("updateTableInfo, gameover")
	}

	req.Id = 10000
	resp = updateTableInfo(req)
	if resp.Result != common.ResultFailTableNotExist {
		t.Errorf("updateTableInfo:%d", resp.Result)
	}

}
