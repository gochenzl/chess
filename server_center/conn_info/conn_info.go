package conn_info

import (
	"sync"

	"github.com/gochenzl/chess/pb/center"
)

type connInfo struct {
	gateid uint32
	connid uint32
}

var mu sync.Mutex
var connInfoMap map[uint32]connInfo // userid-->connInfo
var useridMap map[connInfo]uint32   // conInfo-->userid

func Init(dataPath string) bool {
	connInfoMap = make(map[uint32]connInfo)
	useridMap = make(map[connInfo]uint32)

	return initDur(dataPath)
}

func InitTest() {
	connInfoMap = make(map[uint32]connInfo)
	useridMap = make(map[connInfo]uint32)

	initDurTest()
}

func Add(info center.ConnInfo) (oldUserid uint32, isNew bool) {
	oldUserid, isNew = addHelper(info)
	if isNew {
		durPut(info.Userid, connInfo{info.Gateid, info.Connid})
		if oldUserid != 0 {
			durDel(oldUserid)
		}
	}

	return
}

func Del(gateid uint32, connid uint32) (userid uint32, ok bool) {
	if userid, ok = delHelper(connInfo{gateid, connid}); ok {
		durDel(userid)
	}

	return
}

func DelByGateid(gateid uint32) {
	userids := make([]uint32, 0, 1024)
	mu.Lock()
	for userid, connInfo := range connInfoMap {
		if connInfo.gateid != gateid {
			continue
		}

		delete(connInfoMap, userid)
		delete(useridMap, connInfo)
		userids = append(userids, userid)
	}
	mu.Unlock()

	durDelBatch(userids)
}

func GetAll() (infos []*center.ConnInfo) {
	mu.Lock()
	infos = make([]*center.ConnInfo, 0, len(connInfoMap))
	for userid, connInfo := range connInfoMap {
		var info center.ConnInfo
		info.Userid = userid
		info.Gateid = connInfo.gateid
		info.Connid = connInfo.connid
		infos = append(infos, &info)
	}
	mu.Unlock()

	return
}

func Exist(info center.ConnInfo) bool {
	mu.Lock()
	defer mu.Unlock()

	if _, present := connInfoMap[info.Userid]; !present {
		return false
	}

	if _, present := useridMap[connInfo{info.Gateid, info.Connid}]; !present {
		return false
	}

	return true
}

func addHelper(info center.ConnInfo) (oldUserid uint32, isNew bool) {
	newConnInfo := connInfo{info.Gateid, info.Connid}
	var present bool

	mu.Lock()
	defer mu.Unlock()

	// 连接已经存在
	if oldUserid, present = useridMap[newConnInfo]; present {
		if oldUserid != info.Userid { // 少见的情况，有不同的用户使用同一个连接
			delete(connInfoMap, oldUserid)
		} else { // 重复添加
			return
		}
	}

	var oldConnInfo connInfo
	if oldConnInfo, present = connInfoMap[info.Userid]; present { // 此玩家已经有一条连接
		delete(useridMap, oldConnInfo)
	}

	connInfoMap[info.Userid] = newConnInfo
	useridMap[newConnInfo] = info.Userid
	isNew = true
	return
}

func delHelper(info connInfo) (uint32, bool) {
	mu.Lock()
	defer mu.Unlock()

	var userid uint32
	var present bool
	if userid, present = useridMap[info]; !present {
		return 0, false
	}

	delete(connInfoMap, userid)
	delete(useridMap, info)
	return userid, true
}
