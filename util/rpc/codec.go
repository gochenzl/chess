package rpc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"reflect"

	"github.com/golang/protobuf/proto"
)

// 4字节长度，包括本字段
// 2字节消息名长度
// 消息名（末尾为\0）
// 消息体
// 4字节checksum，adler32 of above

const minMsgSize = 4 + 2 + 2 + 4

var errInvalidMsg = errors.New("invalid rpc msg")
var errChecksum = errors.New("checksum error")

func EncodePb(w io.Writer, pb proto.Message) error {
	name := proto.MessageName(pb)

	msgBody, err := proto.Marshal(pb)
	if err != nil {
		return err
	}

	return Encode(w, name, msgBody)
}

func Encode(w io.Writer, name string, msgBody []byte) error {
	totalSize := 4 + 2 + len(name) + 1 + len(msgBody) + 4
	checksum := adler32.New()

	var buf [4]byte

	binary.LittleEndian.PutUint32(buf[:], uint32(totalSize))
	if _, err := w.Write(buf[:]); err != nil {
		return err
	}
	checksum.Write(buf[:])

	binary.LittleEndian.PutUint16(buf[:], uint16(len(name)+1))
	if _, err := w.Write(buf[:2]); err != nil {
		return err
	}
	checksum.Write(buf[:2])

	nameBytes := []byte(name)
	nameBytes = append(nameBytes, 0)
	if _, err := w.Write(nameBytes); err != nil {
		return err
	}
	checksum.Write(nameBytes)

	if _, err := w.Write(msgBody); err != nil {
		return err
	}
	checksum.Write(msgBody)

	sum32 := checksum.Sum32()
	binary.LittleEndian.PutUint32(buf[:], sum32)
	if _, err := w.Write(buf[:]); err != nil {
		return err
	}

	return nil
}

func Decode(r io.Reader) (name string, msgBody []byte, err error) {
	var lenBuf [4]byte
	if _, err = io.ReadFull(r, lenBuf[:]); err != nil {
		return
	}

	totalSize := int(binary.LittleEndian.Uint32(lenBuf[:]))

	buf := make([]byte, totalSize)
	if _, err = io.ReadFull(r, buf[4:]); err != nil {
		return
	}

	copy(buf[0:], lenBuf[:])

	if binary.LittleEndian.Uint32(buf[totalSize-4:]) != adler32.Checksum(buf[:totalSize-4]) {
		err = errChecksum
		return
	}

	return decodeHelper(buf[4:])
}

func DecodePb(r io.Reader) (pb proto.Message, err error) {
	var msgBody []byte
	var name string

	name, msgBody, err = Decode(r)
	if err != nil {
		return
	}

	rt := proto.MessageType(name)
	if rt == nil {
		err = errors.New(fmt.Sprintf("can't find proto message type for %s", name))
		return
	}

	pb = reflect.New(rt.Elem()).Interface().(proto.Message)
	err = proto.Unmarshal(msgBody, pb)

	return
}

func decodeHelper(buf []byte) (name string, msgBody []byte, err error) {
	bufSize := len(buf)
	if bufSize-4 < minMsgSize {
		err = errInvalidMsg
		return
	}

	offset := 0
	nameLen := int(binary.LittleEndian.Uint16(buf[offset:]))
	offset += 2

	if nameLen < 1 || nameLen > bufSize-offset-4 {
		err = errInvalidMsg
		return
	}

	name = string(buf[offset : offset+nameLen-1])
	offset += nameLen
	if offset < bufSize-4 {
		msgBody = buf[offset : bufSize-4]
	}

	return
}
