package rpc

import (
	"net"
	"time"

	"github.com/gochenzl/chess/util/buf_pool"
	"github.com/gochenzl/chess/util/log"
	"github.com/golang/protobuf/proto"
)

var remoteServers map[string]*connPoolInfo = make(map[string]*connPoolInfo)

func Add(serverName string, hostAndPort string, maxConn int) {
	pool := newConnPool(maxConn, hostAndPort, serverName)
	remoteServers[serverName] = pool
}

func Invoke(serverName string, req proto.Message, resp proto.Message) bool {
	pool, ok := remoteServers[serverName]
	if !ok {
		log.Error("%s not exist", serverName)
		return false
	}

	writer := buf_pool.Get()
	defer buf_pool.Put(writer)

	err := EncodePb(writer, req)
	if err != nil {
		log.Warn("encode protobuf fail:%s", err.Error())
		return false
	}
	buf := writer.Bytes()

	retry := 0

RETRY:
	if retry >= 3 {
		return false
	}

	conn := send(buf, pool, retry > 0)
	if conn == nil {
		time.Sleep(time.Millisecond * 100)
		retry++
		goto RETRY
	}

	name, msgBody, err := Decode(conn)
	if err != nil {
		log.Warn("decode rpc msg fail:%s", err.Error())
		conn.Close()
		if retry == 0 {
			pool.decrConnNum(conn)
		}

		time.Sleep(time.Millisecond * 100)
		retry++
		goto RETRY
	}

	if name != proto.MessageName(resp) {
		log.Warn("recv %s, expect %s", name, proto.MessageName(resp))
		conn.Close()
		if retry == 0 {
			pool.decrConnNum(conn)
		}

		return false
	}

	if err := proto.Unmarshal(msgBody, resp); err != nil {
		log.Warn("decode protobuf fail:%s", err.Error())
		conn.Close()
		if retry == 0 {
			pool.decrConnNum(conn)
		}

		return false
	}

	pool.release(conn)
	return true
}

func Notify(serverName string, req proto.Message) bool {
	pool, ok := remoteServers[serverName]
	if !ok {
		log.Error("%s not exist", serverName)
		return false
	}

	writer := buf_pool.Get()
	defer buf_pool.Put(writer)

	err := EncodePb(writer, req)
	if err != nil {
		log.Warn("encode protobuf fail:%s", err.Error())
		return false
	}
	buf := writer.Bytes()

	retry := 0

RETRY:
	if retry >= 3 {
		return false
	}

	conn := send(buf, pool, retry > 0)
	if conn == nil {
		time.Sleep(time.Millisecond * 100)
		retry++
		goto RETRY
	}

	pool.release(conn)
	return true
}

func send(buf []byte, pool *connPoolInfo, retry bool) net.Conn {
	var conn net.Conn
	if !retry {
		conn = pool.get()
		if conn == nil {
			return nil
		}
	} else {
		conn = pool.createConn()
		if conn == nil {
			return nil
		}
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write(buf); err != nil {
		log.Error("%s", err.Error())
		conn.Close()
		if !retry {
			pool.decrConnNum(conn)
		}

		return nil
	}

	return conn
}
