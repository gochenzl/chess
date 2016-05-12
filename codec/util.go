package codec

import (
	"crypto/cipher"
	"encoding/binary"
	"hash"
	"hash/adler32"
	"io"
)

type encoder struct {
	w        io.Writer
	checksum hash.Hash32
}

func (enc *encoder) putUint32(val uint32) error {
	var buf [4]byte

	binary.LittleEndian.PutUint32(buf[:], val)
	if _, err := enc.w.Write(buf[:]); err != nil {
		return err
	}
	enc.checksum.Write(buf[:])
	return nil
}

func (enc *encoder) putUint16(val uint16) error {
	var buf [4]byte

	binary.LittleEndian.PutUint16(buf[:], val)
	if _, err := enc.w.Write(buf[:2]); err != nil {
		return err
	}
	enc.checksum.Write(buf[:2])
	return nil
}

func (enc *encoder) putBytes(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}

	if _, err := enc.w.Write(buf); err != nil {
		return err
	}
	enc.checksum.Write(buf)
	return nil
}

func (enc *encoder) finish() error {
	var buf [4]byte
	sum32 := enc.checksum.Sum32()

	binary.LittleEndian.PutUint32(buf[:], sum32)
	if _, err := enc.w.Write(buf[:]); err != nil {
		return err
	}
	return nil
}

func readTotalSize(r io.Reader) (int, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}

	return int(binary.LittleEndian.Uint32(buf[:])), nil
}

func putUint32(w io.Writer, val uint32) error {
	var buf [4]byte

	binary.LittleEndian.PutUint32(buf[:], val)
	if _, err := w.Write(buf[:]); err != nil {
		return err
	}

	return nil
}

func checkGameMsg(buf []byte, minSize int) error {
	if len(buf) < minSize {
		return ErrInvalid
	}

	if binary.LittleEndian.Uint32(buf[len(buf)-4:]) != adler32.Checksum(buf[:len(buf)-4]) {
		return ErrChecksum
	}

	return nil
}

func EncryptWithLen(data []byte) []byte {
	length := len(data) + 4
	if length%aesBlock.BlockSize() != 0 {
		length = (length/aesBlock.BlockSize() + 1) * aesBlock.BlockSize()
	}

	newData := make([]byte, length)
	binary.LittleEndian.PutUint32(newData, uint32(len(data)))
	copy(newData[4:], data)

	encrypter := cipher.NewCBCEncrypter(aesBlock, iv)
	encrypter.CryptBlocks(newData, newData)

	return newData
}

func DecryptWithLen(data []byte) []byte {
	if len(data)%aesBlock.BlockSize() != 0 {
		return nil
	}

	newData := make([]byte, len(data))
	decrypter := cipher.NewCBCDecrypter(aesBlock, iv)
	decrypter.CryptBlocks(newData, data)

	length := int(binary.LittleEndian.Uint32(newData))
	if length > len(newData)-4 {
		return nil
	}

	return newData[4 : length+4]
}
