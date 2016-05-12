package pkg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/gochenzl/chess/codec"
	"github.com/gochenzl/chess/common"
	"github.com/gochenzl/chess/util/log"
)

var backendChan chan codec.GateBackend
var connChan chan net.Conn

func init() {
	backendChan = make(chan codec.GateBackend, 10000)
	connChan = make(chan net.Conn, 10)
	go writeBackend()
	go monitorBackendChan()
}

func sendToBackend(gb codec.GateBackend) {
	backendChan <- gb
}

func DoBackend(hostAndPort string) {
	var conn net.Conn
	var err error

BEGIN:
	errLog := true
	for {

		conn, err = net.Dial("tcp", hostAndPort)
		if err != nil {
			if errLog {
				log.Error("%s", err.Error())
				errLog = false
			}

			time.Sleep(time.Millisecond * 500)
			continue
		}

		break
	}

	log.Info("connect to %s success", hostAndPort)

	connChan <- conn

	br := bufio.NewReaderSize(conn, 8*1024*1024)
	for {
		var bg codec.BackendGate
		if err = bg.Decode(br); err != nil {
			log.Error("%s", err.Error())
			conn.Close()
			goto BEGIN
		}

		proccessBg(bg)
	}
}

func proccessBg(bg codec.BackendGate) {
	if len(bg.Connids) > 0 {
		for i := 0; i < len(bg.Connids); i++ {
			writer := getConn(bg.Connids[i])
			if writer == nil {
				continue
			}

			writer.Write(bg.MsgBuf)
		}
	} else {
		writer := getConn(bg.Connid)
		if writer != nil {
			writer.Write(bg.MsgBuf)
		}
	}
}

type backendWriter struct {
	buf  bytes.Buffer
	conn net.Conn
}

func (bw *backendWriter) write(p []byte) (int, error) {
	return bw.buf.Write(p)
}

func (bw *backendWriter) flush() error {
	if bw.buf.Len() == 0 {
		return nil
	}

	if _, err := bw.conn.Write(bw.buf.Bytes()); err != nil {
		return err
	}

	bw.buf.Reset()
	return nil
}

// 转发消息到backend
func writeBackend() {
	pendingBytes := make([]byte, 0, 10240)

	var index int
	bws := make([]*backendWriter, 0, 100)

	for {
	BEGIN:
		if len(bws) == 0 {
			for {
				conn := <-connChan
				bw := &backendWriter{}
				bw.conn = conn
				processGb(buildBackendGateid(), bw, true)
				bws = append(bws, bw)
				break
			}
		}

		if len(pendingBytes) != 0 {
			log.Warn("process pending bytes %d", len(pendingBytes))
			bws[index].write(pendingBytes)

			var err error
			if err = bws[index].flush(); err != nil {
				log.Error("flush pendingBytes fail:%s", err.Error())
				remove(&bws, index)
			} else {
				pendingBytes = pendingBytes[:0]
			}

			index++
			if len(bws) != 0 {
				index = index % len(bws)
			} else {
				index = 0
			}

			if err != nil {
				continue
			}
		}

		for {
			count := 0
			select {
			case conn := <-connChan:
				bw := &backendWriter{}
				bw.conn = conn
				processGb(buildBackendGateid(), bw, true)
				bws = append(bws, bw)

			case gb := <-backendChan:
				processGb(gb, bws[index], false)
				count++
				index++
				index = index % len(bws)
			}

			success := true
			if len(backendChan) == 0 || count > 100 {
				for i := 0; i < len(bws); {
					if err := bws[i].flush(); err != nil {
						log.Error("flush fail:%s", err.Error())
						pendingBytes = append(pendingBytes, bws[i].buf.Bytes()...)
						remove(&bws, i)
						success = false
					} else {
						i++
					}
				}
				count = 0
			}

			if !success {
				index = 0
				goto BEGIN
			}
		}

	}
}

func remove(bws *[]*backendWriter, i int) {
	size := len(*bws)
	(*bws)[i] = (*bws)[size-1]
	*bws = (*bws)[:size-1]
}

func buildBackendGateid() codec.GateBackend {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], common.GetGateid())

	var gb codec.GateBackend
	gb.Msgid = common.MsgGateid
	gb.MsgBuf = buf[:]

	return gb
}

func processGb(gb codec.GateBackend, bw *backendWriter, flush bool) bool {
	gb.Encode(&(bw.buf))

	if flush {
		if err := bw.flush(); err != nil {
			log.Error("flush fail:%s", err.Error())
			return false
		}
	}

	return true
}

func monitorBackendChan() {
	for {
		size := len(backendChan)
		if size > 10 {
			log.Info("backend channel len:%d", size)
		}

		time.Sleep(time.Second * 5)
	}
}
