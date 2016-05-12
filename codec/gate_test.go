package codec

import (
	"bytes"
	"testing"
)

func TestGateBackend(t *testing.T) {
	var gb GateBackend
	gb.Connid = 10
	gb.Msgid = 1
	gb.MsgBuf = []byte("test me")

	buf := bytes.NewBuffer(nil)
	if err := gb.Encode(buf); err != nil {
		t.Errorf("Encode GateBackend:%s", err.Error())
	}

	var gb2 GateBackend
	if err := gb2.Decode(buf); err != nil {
		t.Errorf("Decode GateBackend:%s", err.Error())
	}

	if gb.Connid != gb2.Connid || gb.Msgid != gb2.Msgid || bytes.Compare(gb.MsgBuf, gb2.MsgBuf) != 0 {
		t.Errorf("Equal")
	}
}

func TestBackendGate(t *testing.T) {
	var bg BackendGate
	bg.Connid = 10
	bg.MsgBuf = []byte("test me")

	buf := bytes.NewBuffer(nil)
	if err := bg.Encode(buf); err != nil {
		t.Errorf("Encode BackendGate:%s", err.Error())
	}

	var bg2 BackendGate
	if err := bg2.Decode(buf); err != nil {
		t.Errorf("Decode BackendGate:%s", err.Error())
	}

	if bg2.Connid != bg.Connid || bytes.Compare(bg2.MsgBuf, bg.MsgBuf) != 0 {
		t.Errorf("Equal")
	}
}

func TestBackendGate2(t *testing.T) {
	var bg BackendGate
	bg.Connids = append(bg.Connids, 10)
	bg.Connids = append(bg.Connids, 11)
	bg.Connids = append(bg.Connids, 12)
	bg.Connids = append(bg.Connids, 13)
	bg.MsgBuf = []byte("test me")

	buf := bytes.NewBuffer(nil)
	if err := bg.Encode(buf); err != nil {
		t.Errorf("Encode BackendGate:%s", err.Error())
	}

	var bg2 BackendGate
	if err := bg2.Decode(buf); err != nil {
		t.Errorf("Decode BackendGate:%s", err.Error())
	}

	if bg2.Connid != bg.Connid ||
		bytes.Compare(bg2.MsgBuf, bg.MsgBuf) != 0 ||
		!cmpConnids(bg2.Connids, bg.Connids) {
		t.Errorf("Equal")
	}
}

func cmpConnids(connids []uint32, connids2 []uint32) bool {
	if len(connids) != len(connids2) {
		return false
	}

	for i := 0; i < len(connids); i++ {
		if connids[i] != connids2[i] {
			return false
		}
	}

	return true
}
