package codec

import (
	"encoding/binary"
	"io"
)

// gate->backend
// 4字节下面所有字段长度
// 2字节消息id
// 4字节连接id
// 消息体
type GateBackend struct {
	Msgid  uint16
	Connid uint32
	MsgBuf []byte
}

// gate<-backend
// 4字节下面所有字段长度
// 2字节连接id数量
// 4字节连接id
// 4字节连接id
//..............
//消息体
type BackendGate struct {
	Connid  uint32
	Connids []uint32
	MsgBuf  []byte
}

func (gb GateBackend) Encode(w io.Writer) error {
	totalSize := uint32(6 + len(gb.MsgBuf))

	var headBuf [10]byte
	offset := 0
	binary.LittleEndian.PutUint32(headBuf[:], totalSize)
	offset += 4
	binary.LittleEndian.PutUint16(headBuf[offset:], gb.Msgid)
	offset += 2
	binary.LittleEndian.PutUint32(headBuf[offset:], gb.Connid)

	if _, err := w.Write(headBuf[:]); err != nil {
		return err
	}

	if len(gb.MsgBuf) != 0 {
		if _, err := w.Write(gb.MsgBuf); err != nil {
			return err
		}
	}

	return nil
}

func (gb *GateBackend) Decode(r io.Reader) error {
	totalSize, err := readTotalSize(r)
	if err != nil {
		return err
	}

	if totalSize < 6 {
		return ErrInvalid
	}

	buf := make([]byte, totalSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	var offset int

	gb.Msgid = binary.LittleEndian.Uint16(buf[offset:])
	offset += 2

	gb.Connid = binary.LittleEndian.Uint32(buf[offset:])
	offset += 4

	gb.MsgBuf = buf[offset:]

	return nil
}

func (bg BackendGate) Encode(w io.Writer) error {
	var totalSize int
	if len(bg.Connids) == 0 {
		totalSize = 6 + len(bg.MsgBuf)
	} else {
		totalSize = 2 + 4*len(bg.Connids) + len(bg.MsgBuf)
	}

	buf := make([]byte, 4+totalSize-len(bg.MsgBuf))
	var offset int

	binary.LittleEndian.PutUint32(buf[offset:], uint32(totalSize))
	offset += 4

	if len(bg.Connids) == 0 {
		binary.LittleEndian.PutUint16(buf[offset:], 1)
		offset += 2
		binary.LittleEndian.PutUint32(buf[offset:], bg.Connid)
		offset += 4
	} else {
		binary.LittleEndian.PutUint16(buf[offset:], uint16(len(bg.Connids)))
		offset += 2
		for i := 0; i < len(bg.Connids); i++ {
			binary.LittleEndian.PutUint32(buf[offset:], bg.Connids[i])
			offset += 4
		}
	}

	if _, err := w.Write(buf); err != nil {
		return err
	}

	if _, err := w.Write(bg.MsgBuf); err != nil {
		return err
	}

	return nil
}

func (bg *BackendGate) Decode(r io.Reader) error {
	totalSize, err := readTotalSize(r)
	if err != nil {
		return err
	}

	if totalSize < 6 {
		return ErrInvalid
	}

	buf := make([]byte, totalSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	var offset int

	nconn := int(binary.LittleEndian.Uint16(buf[offset:]))
	offset += 2

	if totalSize < 2+4*nconn {
		return ErrInvalid
	}

	if nconn == 1 {
		bg.Connid = binary.LittleEndian.Uint32(buf[offset:])
		offset += 4
	} else {
		bg.Connids = make([]uint32, nconn)
		for i := 0; i < nconn; i++ {
			bg.Connids[i] = binary.LittleEndian.Uint32(buf[offset:])
			offset += 4
		}
	}

	bg.MsgBuf = buf[offset:]

	return nil
}
