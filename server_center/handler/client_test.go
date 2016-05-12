package handler

import (
	"bytes"
	"testing"
)

func TestClient(t *testing.T) {
	var bufs []*bytes.Buffer
	for i := 0; i < 10; i++ {
		bufs = append(bufs, &bytes.Buffer{})
	}

	for i := 0; i < len(bufs); i++ {
		addClient(bufs[i])
	}

	for i := 0; i < len(bufs); i++ {
		if !exist(bufs[i]) {
			t.Errorf("addClient")
		}
	}

	RemoveClient(bufs[0])
	RemoveClient(bufs[3])
	RemoveClient(bufs[9])

	if exist(bufs[0]) {
		t.Errorf("removeClient")
	}

	if exist(bufs[3]) {
		t.Errorf("removeClient")
	}

	if exist(bufs[9]) {
		t.Errorf("removeClient")
	}
}
