package testutil

import (
	"bytes"
	"net"
	"time"
)

type Conn struct {
	bytes.Buffer
}

type Addr struct {
}

func (addr Addr) String() string {
	return "test_conn"
}

func (addr Addr) Network() string {
	return "test_conn"
}

func (conn *Conn) Close() error {
	return nil
}

func (conn *Conn) LocalAddr() net.Addr {
	return Addr{}
}

func (conn *Conn) RemoteAddr() net.Addr {
	return Addr{}
}

func (conn *Conn) SetDeadline(t time.Time) error {
	return nil
}

func (conn *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (conn *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}
