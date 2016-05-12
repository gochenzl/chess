package codec

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"hash/adler32"
	"io"

	"github.com/gochenzl/chess/util/buf_pool"
)

var ErrInvalid = errors.New("invalid")
var ErrChecksum = errors.New("checksum error")
var ErrCipher = errors.New("cipher error")

var aesBlock cipher.Block
var iv []byte

// 客户端发给服务器的消息格式
//
// 下面所有字段加密后的长度(4字节)
//
//
// 下面所有字段加密前的长度(4字节)
// userid             (4字节)
// 消息id              (2字节)
// 保留，清零           (4字节)
// 消息体
// checksum，使用adler32计算以上所有内容(4字节)
type ClientGame struct {
	Userid  uint32
	Msgid   uint16
	MsgBody []byte
}

// 服务器发给客户端的消息格式
//
// 下面所有字段加密后的长度(4字节)
//
//
// 4字节下面所有字段加密前的长度(4字节)
// 消息id               (2字节)
// 结果码                (2字节)
// 保留，清零             (4字节)
// 消息体
// checksum，使用adler32计算以上所有内容(4字节)

type GameClient struct {
	Msgid   uint16
	Result  uint16
	MsgBody []byte
}

func Init(key []byte, inputIV []byte) {
	if len(key) != 32 {
		panic("aes key should be 32 bytes")
	}

	if len(inputIV) != aes.BlockSize {
		panic("aes iv should be 16 bytes")
	}

	aesBlock, _ = aes.NewCipher(key)
	iv = inputIV
}

func (cg *ClientGame) Decode(buf []byte) error {
	if len(buf) == 0 || len(buf)%aesBlock.BlockSize() != 0 {
		return ErrInvalid
	}

	decrypter := cipher.NewCBCDecrypter(aesBlock, iv)
	decrypter.CryptBlocks(buf, buf)

	msgLen := binary.LittleEndian.Uint32(buf)
	if uint32(len(buf[4:])) < msgLen {
		return ErrInvalid
	}

	// msgLen不包括长度字段，所以要加4
	if err := checkGameMsg(buf[:msgLen+4], 18); err != nil {
		return err
	}

	decodeClientGame(buf[4:msgLen+4], cg)
	return nil
}

func (cg *ClientGame) Encode(w io.Writer) error {
	buf := buf_pool.Get()
	defer buf_pool.Put(buf)
	enc := encoder{w: buf, checksum: adler32.New()}

	if err := enc.putUint32(14 + uint32(len(cg.MsgBody))); err != nil {
		return err
	}

	if err := enc.putUint32(cg.Userid); err != nil {
		return err
	}

	if err := enc.putUint16(cg.Msgid); err != nil {
		return err
	}

	if err := enc.putUint32(0); err != nil {
		return err
	}

	if err := enc.putBytes(cg.MsgBody); err != nil {
		return err
	}

	if err := enc.finish(); err != nil {
		return err
	}

	data := buf.Bytes()
	for len(data)%aesBlock.BlockSize() != 0 {
		data = append(data, 0)
	}

	encrypter := cipher.NewCBCEncrypter(aesBlock, iv)
	encrypter.CryptBlocks(data, data)

	if err := putUint32(w, uint32(len(data))); err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil

}

func (gc *GameClient) Decode(buf []byte) error {
	if len(buf) == 0 || len(buf)%aesBlock.BlockSize() != 0 {
		return ErrInvalid
	}

	decrypter := cipher.NewCBCDecrypter(aesBlock, iv)
	decrypter.CryptBlocks(buf, buf)

	msgLen := binary.LittleEndian.Uint32(buf)
	if uint32(len(buf[4:])) < msgLen {
		return ErrInvalid
	}

	if err := checkGameMsg(buf[:msgLen+4], 16); err != nil {
		return err
	}

	decodeGameClient(buf[4:msgLen+4], gc)
	return nil
}

func (gc *GameClient) DecodeFromReader(reader io.Reader) error {
	totalSize, err := readTotalSize(reader)
	if err != nil {
		return err
	}

	buf := make([]byte, totalSize)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return err
	}

	return gc.Decode(buf)
}

func (gc *GameClient) Encode(w io.Writer) error {
	buf := buf_pool.Get()
	defer buf_pool.Put(buf)
	enc := encoder{w: buf, checksum: adler32.New()}

	if err := enc.putUint32(12 + uint32(len(gc.MsgBody))); err != nil {
		return err
	}

	if err := enc.putUint16(gc.Msgid); err != nil {
		return err
	}

	if err := enc.putUint16(gc.Result); err != nil {
		return err
	}

	if err := enc.putUint32(0); err != nil {
		return err
	}

	if err := enc.putBytes(gc.MsgBody); err != nil {
		return err
	}

	if err := enc.finish(); err != nil {
		return err
	}

	data := buf.Bytes()
	for len(data)%aesBlock.BlockSize() != 0 {
		data = append(data, 0)
	}

	encrypter := cipher.NewCBCEncrypter(aesBlock, iv)
	encrypter.CryptBlocks(data, data)

	if err := putUint32(w, uint32(len(data))); err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func decodeClientGame(buf []byte, cg *ClientGame) {
	offset := 0

	cg.Userid = binary.LittleEndian.Uint32(buf[offset:])
	offset += 4

	cg.Msgid = binary.LittleEndian.Uint16(buf[offset:])
	offset += 2

	offset += 4

	cg.MsgBody = buf[offset : len(buf)-4]
	if len(cg.MsgBody) == 0 {
		cg.MsgBody = nil
	}
}

func decodeGameClient(buf []byte, gc *GameClient) {
	offset := 0

	gc.Msgid = binary.LittleEndian.Uint16(buf[offset:])
	offset += 2

	gc.Result = binary.LittleEndian.Uint16(buf[offset:])
	offset += 2

	offset += 4

	gc.MsgBody = buf[offset : len(buf)-4]
	if len(gc.MsgBody) == 0 {
		gc.MsgBody = nil
	}
}
