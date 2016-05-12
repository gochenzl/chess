package rpc

import (
	"bytes"
	"testing"
)

func TestCodec(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	Encode(buf, "chenzl", []byte("hello world"))
	name, msgBody, err := Decode(buf)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if name != "chenzl" {
		t.Errorf("name = %s", name)
	}

	if string(msgBody) != "hello world" {
		t.Errorf("msgBody = %s", string(msgBody))
	}
}
