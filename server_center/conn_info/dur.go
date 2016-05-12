package conn_info

import (
	"encoding/binary"

	"github.com/gochenzl/chess/util/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

var durDb *leveldb.DB

func initDurTest() {
	durDb, _ = leveldb.Open(storage.NewMemStorage(), nil)
}

func initDur(dataPath string) bool {
	var err error
	durDb, err = leveldb.OpenFile(dataPath, nil)
	if err != nil {
		log.Error("open db:%s", err.Error())
		if errors.IsCorrupted(err) {
			durDb, err = leveldb.RecoverFile(dataPath, nil)
			if err != nil {
				log.Error("recover db:%s", err.Error())
				return false
			}
		} else {
			return false
		}
	}

	return load()
}

func Close() {
	durDb.Close()
}

func load() bool {
	iter := durDb.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		if len(key) != 4 || len(value) != 8 {
			continue
		}

		var info connInfo
		userid := binary.LittleEndian.Uint32(key)
		info.gateid = binary.LittleEndian.Uint32(value)
		info.connid = binary.LittleEndian.Uint32(value[4:])

		connInfoMap[userid] = info
		useridMap[info] = userid
	}

	if err := iter.Error(); err != nil {
		log.Error("iterator error:%s", err.Error())
	}

	return true
}

func durPut(userid uint32, info connInfo) {
	var key [4]byte
	var value [8]byte

	binary.LittleEndian.PutUint32(key[:], userid)
	binary.LittleEndian.PutUint32(value[:], info.gateid)
	binary.LittleEndian.PutUint32(value[4:], info.connid)

	if err := durDb.Put(key[:], value[:], nil); err != nil {
		log.Error("put error:%s", err.Error())
	}
}

func durDel(userid uint32) {
	var key [4]byte
	binary.LittleEndian.PutUint32(key[:], userid)

	if err := durDb.Delete(key[:], nil); err != nil {
		log.Error("delete error:%s", err.Error())
	}
}

func durDelBatch(userids []uint32) {
	batch := new(leveldb.Batch)
	var key [4]byte

	for i := 0; i < len(userids); i++ {
		binary.LittleEndian.PutUint32(key[:], userids[i])
		batch.Delete(key[:])
	}

	if err := durDb.Write(batch, nil); err != nil {
		log.Error("delete batch error:%s", err.Error())
	}
}
