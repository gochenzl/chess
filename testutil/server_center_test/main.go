package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gochenzl/chess/pb/center"
	"github.com/gochenzl/chess/util/log"
	"github.com/gochenzl/chess/util/rpc"
	"github.com/golang/protobuf/proto"
)

func sendGetAllConnInfoReq(conn net.Conn) error {
	var req center.GetAllConnInfoReq
	return rpc.EncodePb(conn, &req)
}

func centerClient(hostAndPort string) {
	var conn net.Conn
	var err error

CREATE_CONN:
	printLog := true
	for {
		conn, err = net.Dial("tcp", hostAndPort)
		if err != nil {
			if printLog {
				log.Error("connect to center server fail: %s", err.Error())
				printLog = false
			}

			time.Sleep(time.Second * 1)
			continue
		}

		if err = sendGetAllConnInfoReq(conn); err != nil {
			log.Error("sendGetAllConnInfoReq fail: %s", err.Error())
			conn.Close()
			time.Sleep(time.Second * 1)
			continue
		}

		break
	}

	br := bufio.NewReader(conn)
	for {
		pbMsg, err := rpc.DecodePb(br)
		if err != nil {
			log.Error("recieve notify fail: %s", err.Error())
			conn.Close()
			goto CREATE_CONN
		}

		log.Info("receive %s: %s", proto.MessageName(pbMsg), pbMsg.String())
	}
}

func add(userid uint32, gateid uint32, connid uint32) {
	var req center.AddConnInfoReq
	var resp center.AddConnInfoResp

	req.Info = &center.ConnInfo{}
	req.Info.Userid = userid
	req.Info.Gateid = gateid
	req.Info.Connid = connid

	if !rpc.Invoke("center", &req, &resp) {
		log.Error("AddConnInfoReq fail")
	}
}

func del(gateid uint32, connid uint32) {
	var req center.DelConnInfoReq
	var resp center.DelConnInfoResp

	req.Gateid = gateid
	req.Connid = connid

	if !rpc.Invoke("center", &req, &resp) {
		log.Error("DelConnInfoReq fail")
	}
}

func main() {
	log.OpenDefaultLog("log.txt", log.LevelDebug, 100000, true)
	go centerClient("127.0.0.1:9090")

	rpc.Add("center", "127.0.0.1:9090", 10)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(">")
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")

		cmd := strings.TrimSpace(fields[0])
		switch cmd {
		case "add":
			if len(fields) != 4 {
				fmt.Print(">")
				continue
			}

			userid, _ := strconv.Atoi(strings.TrimSpace(fields[1]))
			gateid, _ := strconv.Atoi(strings.TrimSpace(fields[2]))
			connid, _ := strconv.Atoi(strings.TrimSpace(fields[3]))
			add(uint32(userid), uint32(gateid), uint32(connid))

		case "del":
			if len(fields) != 3 {
				fmt.Print(">")
				continue
			}

			gateid, _ := strconv.Atoi(strings.TrimSpace(fields[1]))
			connid, _ := strconv.Atoi(strings.TrimSpace(fields[2]))
			del(uint32(gateid), uint32(connid))

		}

		fmt.Print(">")
	}
}
