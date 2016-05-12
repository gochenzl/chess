package user

import (
	"time"

	"gopkg.in/redis.v3"
)

// 这里使用支持redis协议的数据库，当然也可以改用其他的数据库

var dbClient *redis.Client

func Init(hostAndPort string) bool {
	dbClient = redis.NewClient(&redis.Options{
		Addr:        hostAndPort,
		MaxRetries:  3,
		ReadTimeout: time.Millisecond * 1000,
		PoolSize:    1000,
		PoolTimeout: time.Millisecond * 100,
	})

	return true
}
