package redis

import (
	"bufio"
	"errors"
	"io"
	"strconv"

	"github.com/gochenzl/chess/util/log"
)

const (
	SimpleString = iota
	Error
	Integer
	BulkString
	Array
)

type Proto struct {
	Type  int
	Str   string   // for SimpleString, Error, first byte is type char +-
	Raw   []byte   // for BulkString
	Int   int      // for Integer
	Elems []*Proto // for Array
}

var EmptyBytes []byte = make([]byte, 0, 1)
var ErrInvalidProto = errors.New("invalid protocol")

func (proto *Proto) GetCommandName() string {
	if proto.Type != Array {
		return ""
	}

	if len(proto.Elems) == 0 {
		return ""
	}

	if proto.Elems[0].Type != BulkString {
		return ""
	}

	return string(proto.Elems[0].Raw)
}

func (proto *Proto) Valid() bool {
	return proto.Type == Array
}

func (proto *Proto) AppendBulkString(value []byte) {
	var elem Proto
	elem.Type = BulkString
	elem.Raw = value
	proto.Elems = append(proto.Elems, &elem)
}

func (proto *Proto) Pack(w *bufio.Writer) bool {
	var err error

	switch proto.Type {
	case SimpleString:
		_, err = w.WriteString(proto.Str)

	case Error:
		_, err = w.WriteString(proto.Str)

	case Integer:
		_, err = w.WriteString(":" + strconv.Itoa(proto.Int) + "\r\n")

	case BulkString:
		if proto.Raw == nil {
			_, err = w.WriteString("$-1\r\n")
			break
		}

		_, err = w.WriteString("$" + strconv.Itoa(len(proto.Raw)) + "\r\n")
		if err != nil {
			break
		}
		_, err = w.Write(proto.Raw)
		if err != nil {
			break
		}
		_, err = w.WriteString("\r\n")

	case Array:
		_, err = w.WriteString("*" + strconv.Itoa(len(proto.Elems)) + "\r\n")
		if err != nil {
			break
		}
		for i := 0; i < len(proto.Elems); i++ {
			if !proto.Elems[i].Pack(w) {
				return false
			}
		}
	}

	if err != nil {
		log.Error("%v", err)
		return false
	}

	return true
}

func readLine(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' { // \r\n
		return nil, ErrInvalidProto
	}

	return line, nil
}

func Btoi(b []byte) (int, bool) {
	n := 0
	sign := 1
	for i := uint8(0); i < uint8(len(b)); i++ {
		if i == 0 && b[i] == '-' {
			if len(b) == 1 {
				return 0, false
			}
			sign = -1
			continue
		}

		if b[i] >= '0' && b[i] <= '9' {
			if i > 0 {
				n *= 10
			}
			n += int(b[i]) - '0'
			continue
		}

		return 0, false
	}

	return sign * n, true
}

func readBulk(r *bufio.Reader, size int, raw *[]byte) error {
	if size < 0 {
		return nil
	}

	size += 2 // \r\n

	*raw = make([]byte, 0, size)

	//avoid copy
	if _, err := io.ReadFull(r, (*raw)[0:size]); err != nil {
		return err
	}
	*raw = (*raw)[0:size:cap(*raw)]

	if (*raw)[size-2] != '\r' || (*raw)[size-1] != '\n' {
		return ErrInvalidProto
	}

	*raw = (*raw)[0 : size-2]

	return nil
}

func Parse(r *bufio.Reader) (*Proto, error) {
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}

	proto := &Proto{}

	switch line[0] {
	case '+':
		proto.Type = SimpleString
		proto.Str = string(line[:len(line)-2])
		return proto, nil

	case '-':
		proto.Type = Error
		proto.Str = string(line[:len(line)-2])
		return proto, nil

	case ':':
		proto.Type = Integer
		intValue, success := Btoi(line[1 : len(line)-2])
		if !success {
			return nil, ErrInvalidProto
		}
		proto.Int = intValue
		return proto, nil

	case '$':
		proto.Type = BulkString
		size, success := Btoi(line[1 : len(line)-2])
		if !success {
			return nil, ErrInvalidProto
		}
		if err := readBulk(r, size, &proto.Raw); err != nil {
			return nil, err
		}
		return proto, nil

	case '*':
		i, success := Btoi(line[1 : len(line)-2]) //strip \r\n
		if !success {
			return nil, ErrInvalidProto
		}
		proto.Type = Array
		if i >= 0 {
			elems := make([]*Proto, i)
			for j := 0; j < i; j++ {
				rp, err := Parse(r)
				if err != nil {
					return nil, err
				}
				elems[j] = rp
			}
			proto.Elems = elems
		}
		return proto, nil
	}

	return nil, ErrInvalidProto
}
