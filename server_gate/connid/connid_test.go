package connid

import "testing"

func TestConnid(t *testing.T) {
	Init()

	ids := make([]uint32, 0, maxConnNum)

	for i := 0; i < maxConnNum; i++ {
		id := Get()
		if id == InvalidId {
			t.Error("Get")
			return
		}
		ids = append(ids, id)
	}

	var id uint32 = 1
	for i := 0; i < len(ids); i++ {
		if ids[i] != id {
			t.Error("Get")
			return
		}
		id++
	}

	if Get() != InvalidId {
		t.Error("exhaust")
	}

	for i := 0; i < len(ids); i++ {
		if !Release(ids[i]) {
			t.Error("Release")
			return
		}
	}

	for i := 0; i < len(ids); i++ {
		if Release(ids[i]) {
			t.Error("Release")
			return
		}
	}

	for i := 0; i < 10; i++ {
		ids := ids[:0]
		for {
			id := Get()
			if id == InvalidId {
				break
			}
			ids = append(ids, id)
		}

		for i := 0; i < len(ids); i++ {
			if !Release(ids[i]) {
				t.Error("Release")
			}
		}
	}
}
