package redis

import (
	"testing"
	"time"

	"gopkg.in/redis.v3"
)

func TestServer(t *testing.T) {
	hostAndPort := "127.0.0.1:8999"
	server := NewServer(hostAndPort, NewMemoryStore())
	if err := server.Run(); err != nil {
		t.Errorf("run server fail:%s", err.Error())
	}

	dbClient := redis.NewClient(&redis.Options{
		Addr:        hostAndPort,
		MaxRetries:  3,
		ReadTimeout: time.Millisecond * 300,
		PoolSize:    1000,
		PoolTimeout: time.Millisecond * 100,
	})

	boolCmd := dbClient.HSet("testkey", "testfield", "testvalue")
	newSet, _ := boolCmd.Result()
	if !newSet {
		t.Error("hset error")
	}

	boolCmd = dbClient.HSet("testkey", "testfield", "testvalue")
	newSet, _ = boolCmd.Result()
	if newSet {
		t.Error("hset error")
	}

	stringCmd := dbClient.HGet("testkey", "testfield")
	if stringCmd.Val() != "testvalue" {
		t.Errorf("hget fail, ret = %s", stringCmd.Val())
	}

	intCmd := dbClient.RPush("test_list", "abcde")
	if intCmd.Val() != 1 {
		t.Errorf("rpush fail")
	}

	intCmd = dbClient.RPush("test_list", "abcdef")
	if intCmd.Val() != 2 {
		t.Errorf("rpush fail")
	}

	stringCmd = dbClient.LPop("test_list")
	if stringCmd.Val() != "abcde" {
		t.Errorf("lpop fail, %s", stringCmd.String())
	}

	stringCmd = dbClient.LPop("test_list")
	if stringCmd.Val() != "abcdef" {
		t.Errorf("lpop fail, %s", stringCmd.String())
	}

	stringCmd = dbClient.LPop("test_list")
	if stringCmd.Err() != redis.Nil {
		t.Errorf("lpop fail")
	}
}
