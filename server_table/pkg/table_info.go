package pkg

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/pb/table"
	"github.com/gochenzl/chess/util/redis_cli"
)

type tableInfo struct {
	id       int64
	roomid   int32
	version  int32
	userids  []uint32
	gameInfo []byte
	mu       sync.Mutex

	timer *time.Timer
}

var tableid int64

func newTableInfo(roomid int32, userids []uint32) *tableInfo {
	ti := &tableInfo{}
	ti.id = atomic.AddInt64(&tableid, 1)
	ti.roomid = roomid
	ti.version = 1
	ti.userids = userids

	return ti
}

func (ti *tableInfo) savepb() (pbInfo *table.TableInfo) {
	ti.mu.Lock()

	pbInfo = &table.TableInfo{}
	pbInfo.Id = ti.id
	pbInfo.Roomid = ti.roomid
	pbInfo.Version = ti.version
	pbInfo.IsSetTimer = (ti.timer != nil)
	pbInfo.GameInfo = ti.gameInfo
	pbInfo.Userids = ti.userids

	ti.mu.Unlock()
	return
}

func (ti *tableInfo) update(req *table.UpdateTableInfoReq) bool {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	if ti.version != req.Version {
		return false
	}

	ti.version++
	ti.gameInfo = req.GameInfo

	if req.TimerInfo != nil {
		seconds := time.Duration(req.TimerInfo.Duration)
		ti.setTimer(seconds*time.Second, ti.timeout)
	}

	if req.GameOver {
		ti.stopTimer()
	}

	return true
}

func (ti *tableInfo) timeout() {
	s := strconv.FormatInt(ti.id, 10)
	redis_cli.RPush(common.TableTimeoutList, s)
}

func (ti *tableInfo) setTimer(d time.Duration, f func()) {
	if ti.timer != nil {
		ti.timer.Stop()
	}

	ti.timer = time.AfterFunc(d, f)
}

func (ti *tableInfo) stopTimer() {
	if ti.timer != nil {
		ti.timer.Stop()
	}
}
