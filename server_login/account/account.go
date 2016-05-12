package account

// just for test

import (
	"encoding/json"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb"
)

type AccountInfo struct {
	Username string `json:"username"`
	Userid   uint32 `json:"userid"`
}

var db *leveldb.DB
var maxUserid uint32

func Init() {
	os.Mkdir("data", 0777)

	db, _ = leveldb.OpenFile("data", nil)
	it := db.NewIterator(nil, nil)
	for it.Next() {
		value := it.Value()
		var info AccountInfo
		json.Unmarshal(value, &info)
		if maxUserid < info.Userid {
			maxUserid = info.Userid
		}
	}

	if maxUserid == 0 {
		maxUserid = 100000
	}
}

func New() *AccountInfo {
	var info AccountInfo

	info.Userid = atomic.AddUint32(&maxUserid, 1)
	info.Username = "VK" + strconv.FormatUint(uint64(info.Userid), 10)

	value, _ := json.Marshal(&info)
	db.Put([]byte(info.Username), value, nil)
	return &info
}

func Query(username string) *AccountInfo {
	value, err := db.Get([]byte(username), nil)
	if err != nil {
		return nil
	}

	var info AccountInfo
	json.Unmarshal(value, &info)
	return &info
}
