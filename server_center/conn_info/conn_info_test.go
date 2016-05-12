package conn_info

import (
	"testing"

	"github.com/gochenzl/chess/pb/center"
)

func TestAddConnInfo(t *testing.T) {
	InitTest()

	oldUserid, isNew := Add(center.ConnInfo{10000, 1, 1})
	if oldUserid != 0 {
		t.Errorf("add oldUserid=%d", oldUserid)
	}
	if !isNew {
		t.Errorf("add isNew")
	}

	if info, ok := connInfoMap[10000]; !ok || info != (connInfo{1, 1}) {
		t.Errorf("connInfoMap")
	}

	if userid, ok := useridMap[connInfo{1, 1}]; !ok || userid != 10000 {
		t.Errorf("useridMap")
	}

	// 重复添加
	oldUserid, isNew = Add(center.ConnInfo{10000, 1, 1})
	if oldUserid != 10000 {
		t.Errorf("add oldUserid=%d", oldUserid)
	}
	if isNew {
		t.Errorf("add isNew")
	}

	// 相同连接，不同玩家
	oldUserid, isNew = Add(center.ConnInfo{10001, 1, 1})
	if oldUserid != 10000 {
		t.Errorf("add oldUserid=%d", oldUserid)
	}
	if !isNew {
		t.Errorf("add isNew")
	}

	if _, ok := connInfoMap[10000]; ok {
		t.Errorf("connInfoMap")
	}

	if info, ok := connInfoMap[10001]; !ok || info != (connInfo{1, 1}) {
		t.Errorf("connInfoMap")
	}

	if userid, ok := useridMap[connInfo{1, 1}]; !ok || userid != 10001 {
		t.Errorf("useridMap")
	}

	// 同一玩家，不同连接
	oldUserid, isNew = Add(center.ConnInfo{10001, 1, 2})
	if oldUserid != 0 {
		t.Errorf("add oldUserid=%d", oldUserid)
	}
	if !isNew {
		t.Errorf("add isNew")
	}

	if info, ok := connInfoMap[10001]; !ok || info != (connInfo{1, 2}) {
		t.Errorf("connInfoMap")
	}

	if _, ok := useridMap[connInfo{1, 1}]; ok {
		t.Errorf("useridMap")
	}

	if userid, ok := useridMap[connInfo{1, 2}]; !ok || userid != 10001 {
		t.Errorf("useridMap")
	}
}

func TestDelConnInfo(t *testing.T) {
	InitTest()

	if _, ok := Del(1, 1); ok {
		t.Errorf("del not exist")
	}

	Add(center.ConnInfo{10000, 1, 1})
	if _, ok := Del(1, 1); !ok {
		t.Errorf("del exist")
	}

	if _, ok := connInfoMap[10000]; ok {
		t.Errorf("connInfoMap")
	}

	if _, ok := useridMap[connInfo{1, 1}]; ok {
		t.Errorf("useridMap")
	}
}

func TestDelByGateidConnInfo(t *testing.T) {
	InitTest()

	var i uint32

	for i = 1; i <= 10; i++ {
		Add(center.ConnInfo{10000 + i, 1, 1 + i})
	}

	for i = 1; i <= 10; i++ {
		Add(center.ConnInfo{20000 + i, 2, 1 + i})
	}

	DelByGateid(1)

	for i = 1; i <= 10; i++ {
		if _, ok := useridMap[connInfo{1, 1 + i}]; ok {
			t.Errorf("useridMap")
		}

		if _, ok := connInfoMap[10000+i]; ok {
			t.Errorf("connInfoMap")
		}

		if info, ok := connInfoMap[20000+i]; !ok || info != (connInfo{2, 1 + i}) {
			t.Errorf("connInfoMap")
		}

		if userid, ok := useridMap[connInfo{2, 1 + i}]; !ok || userid != 20000+i {
			t.Errorf("connInfoMap")
		}
	}
}
