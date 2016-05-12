package conn_info

import "testing"

func TestDurPut(t *testing.T) {
	InitTest()

	var i uint32

	for i = 1; i < 10; i++ {
		durPut(10000+i, connInfo{1, i + 1})
	}

	load()

	for i = 1; i < 10; i++ {
		if userid, ok := useridMap[connInfo{1, 1 + i}]; !ok || userid != 10000+i {
			t.Errorf("useridMap")
		}

		if info, ok := connInfoMap[10000+i]; !ok || info != (connInfo{1, i + 1}) {
			t.Errorf("connInfoMap")
		}
	}
}

func TestDurDel(t *testing.T) {
	InitTest()

	durPut(10001, connInfo{1, 1})
	durPut(10002, connInfo{1, 2})
	durPut(10003, connInfo{1, 3})
	durPut(10004, connInfo{1, 4})
	durPut(10005, connInfo{1, 5})

	durDel(10002)
	durDel(10005)

	load()

	if info, ok := connInfoMap[10001]; !ok || info != (connInfo{1, 1}) {
		t.Errorf("connInfoMap")
	}

	if _, ok := connInfoMap[10002]; ok {
		t.Errorf("connInfoMap")
	}

	if info, ok := connInfoMap[10003]; !ok || info != (connInfo{1, 3}) {
		t.Errorf("connInfoMap")
	}

	if info, ok := connInfoMap[10004]; !ok || info != (connInfo{1, 4}) {
		t.Errorf("connInfoMap")
	}

	if _, ok := connInfoMap[10005]; ok {
		t.Errorf("connInfoMap")
	}
}

func TestDurDelBatch(t *testing.T) {
	InitTest()

	durPut(10001, connInfo{1, 1})
	durPut(10002, connInfo{1, 2})
	durPut(10003, connInfo{1, 3})
	durPut(10004, connInfo{1, 4})
	durPut(10005, connInfo{1, 5})

	var userids [2]uint32 = [2]uint32{10002, 10005}
	durDelBatch(userids[:])

	load()

	if info, ok := connInfoMap[10001]; !ok || info != (connInfo{1, 1}) {
		t.Errorf("connInfoMap")
	}

	if _, ok := connInfoMap[10002]; ok {
		t.Errorf("connInfoMap")
	}

	if info, ok := connInfoMap[10003]; !ok || info != (connInfo{1, 3}) {
		t.Errorf("connInfoMap")
	}

	if info, ok := connInfoMap[10004]; !ok || info != (connInfo{1, 4}) {
		t.Errorf("connInfoMap")
	}

	if _, ok := connInfoMap[10005]; ok {
		t.Errorf("connInfoMap")
	}
}
