package codec

import (
	"bytes"
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()

	key := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
		0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f}
	iv := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f}
	Init(key, iv)
	os.Exit(m.Run())
}

func TestClientGame(t *testing.T) {
	var cg ClientGame
	cg.Userid = 18888
	cg.Msgid = 28
	cg.MsgBody = []byte("hello world hello world hello world")

	buf := bytes.NewBuffer(nil)
	if err := cg.Encode(buf); err != nil {
		t.Errorf("Encode ClientGame:%s", err.Error())
	}

	var cg2 ClientGame
	if err := cg2.Decode(buf.Bytes()[4:]); err != nil {
		t.Errorf("Decode ClientGame:%s", err.Error())
		return
	}

	if cg2.Userid != cg.Userid {
		t.Errorf("Userid:%d", cg2.Userid)
	}

	if cg2.Msgid != cg.Msgid {
		t.Errorf("Msgid:%d", cg2.Msgid)
	}

	if string(cg2.MsgBody) != string(cg.MsgBody) {
		t.Errorf("MsgBody:%s", string(cg2.MsgBody))
	}

	cg.MsgBody = nil
	buf = bytes.NewBuffer(nil)
	if err := cg.Encode(buf); err != nil {
		t.Errorf("Encode ClientGame:%s", err.Error())
	}

	if err := cg2.Decode(buf.Bytes()[4:]); err != nil {
		t.Errorf("Decode ClientGame:%s", err.Error())
	}

	if cg2.MsgBody != nil {
		t.Errorf("MsgBody")
	}
}

func TestGameClient(t *testing.T) {
	var gc GameClient
	gc.Msgid = 38
	gc.Result = 19
	gc.MsgBody = []byte("hello world hello world hello world")

	buf := bytes.NewBuffer(nil)
	if err := gc.Encode(buf); err != nil {
		t.Errorf("Encode GameClient:%s", err.Error())
	}

	var gc2 GameClient
	if err := gc2.Decode(buf.Bytes()[4:]); err != nil {
		t.Errorf("Decode GameClient:%s", err.Error())
	}

	if gc2.Msgid != gc.Msgid {
		t.Errorf("Msgid:%d", gc2.Msgid)
	}

	if gc2.Result != gc.Result {
		t.Errorf("Result:%d", gc2.Result)
	}

	if string(gc2.MsgBody) != string(gc.MsgBody) {
		t.Errorf("MsgBody:%s", string(gc2.MsgBody))
	}

	gc.MsgBody = nil

	buf = bytes.NewBuffer(nil)
	if err := gc.Encode(buf); err != nil {
		t.Errorf("Encode GameClient:%s", err.Error())
	}

	if err := gc2.Decode(buf.Bytes()[4:]); err != nil {
		t.Errorf("Decode GameClient:%s", err.Error())
	}

	if gc2.MsgBody != nil {
		t.Errorf("MsgBody")
	}
}

func TestEncrypt(t *testing.T) {
	data := []byte{202, 255, 115, 220, 17, 107, 40, 194, 133, 242, 141}
	newData := EncryptWithLen(data)
	data2 := DecryptWithLen(newData)

	if bytes.Compare(data, data2) != 0 {
		t.Errorf("EncryptWithLen:%v", data2)
	}

	data = []byte{202, 255, 115, 220, 17, 107, 40, 194, 133, 242, 141, 32, 242, 193, 243, 99}
	newData = EncryptWithLen(data)
	data2 = DecryptWithLen(newData)

	if bytes.Compare(data, data2) != 0 {
		t.Errorf("EncryptWithLen:%v", data2)
	}

	data = []byte{202, 255, 115, 220, 17, 107, 40, 194, 133, 242, 141, 32, 242, 193, 243, 99, 82, 125, 170, 166}
	newData = EncryptWithLen(data)
	data2 = DecryptWithLen(newData)

	if bytes.Compare(data, data2) != 0 {
		t.Errorf("EncryptWithLen:%v", data2)
	}

	data = []byte{202, 255, 115, 220, 17, 107, 40, 194, 133, 242, 141, 32, 242, 193, 243, 99, 82, 125, 170, 166, 180, 218, 69, 106, 177, 34, 13, 12, 98, 232, 0, 33}
	newData = EncryptWithLen(data)
	data2 = DecryptWithLen(newData)

	if bytes.Compare(data, data2) != 0 {
		t.Errorf("EncryptWithLen:%v", data2)
	}
}
