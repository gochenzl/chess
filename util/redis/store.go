package redis

import "errors"

type Store interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	HGet(key, field []byte) ([]byte, error)
	HGetAll(key []byte) ([][]byte, error)
	HSet(key, field, value []byte) (bool, error)
	HIncrBy(key, field []byte, incr int64) (int64, error)

	RPush(key, value []byte) (int64, error)
	LPop(key []byte) ([]byte, error)
}

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
var ErrKeyNotExist = errors.New("key not exist")
