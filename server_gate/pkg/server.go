package pkg

import (
	"net"
	"strconv"
)

func Serve(listenPort int) error {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(listenPort))
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go doFrontEnd(conn)
	}
}

func Init() {
	go processMsgQueue()
	go printStat()
}
