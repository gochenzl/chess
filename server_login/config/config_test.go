package config

import (
	"sync"
	"testing"
)

func TestFindGameServer(t *testing.T) {
	gameServerGroup.mu = &sync.RWMutex{}
	if addr := FindGameServer(1); len(addr) != 0 {
		t.Error("empty")
	}

	var addrs []string
	addrs = append(addrs, "127.0.0.1:8000")
	addrs = append(addrs, "127.0.0.1:8001")
	addrs = append(addrs, "127.0.0.1:8002")
	gameServerGroup.Servers = append(gameServerGroup.Servers, gameServer{ClientVersion: 1, Addrs: addrs})

	var addrs2 []string
	addrs2 = append(addrs2, "127.0.0.1:9000")
	addrs2 = append(addrs2, "127.0.0.1:9001")
	addrs2 = append(addrs2, "127.0.0.1:9002")
	gameServerGroup.Servers = append(gameServerGroup.Servers, gameServer{ClientVersion: 2, Addrs: addrs2})

	for i := 0; i < len(addrs); i++ {
		if addr := FindGameServer(1); addr != addrs[i] {
			t.Error("exactly")
		}
	}

	for i := 0; i < len(addrs); i++ {
		if addr := FindGameServer(1); addr != addrs[i] {
			t.Error("exactly")
		}
	}

	for i := 0; i < len(addrs2); i++ {
		if addr := FindGameServer(2); addr != addrs2[i] {
			t.Error("exactly")
		}
	}

	for i := 0; i < len(addrs2); i++ {
		if addr := FindGameServer(2); addr != addrs2[i] {
			t.Error("exactly")
		}
	}

	for i := 0; i < len(addrs); i++ {
		if addr := FindGameServer(3); addr != addrs[i] {
			t.Error("default")
		}
	}
}
