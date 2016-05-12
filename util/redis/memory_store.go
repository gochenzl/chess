package redis

import (
	"container/list"
	"strconv"
	"sync"
)

type kvStore struct {
	store map[string]string
}

func newKvStore() kvStore {
	kv := kvStore{}
	kv.store = make(map[string]string)
	return kv
}

func (kv *kvStore) get(key string) (string, bool) {
	value, exist := kv.store[key]
	return value, exist
}

func (kv *kvStore) set(key, value string) {
	kv.store[key] = value
}

type memoryStore struct {
	hash  map[string]kvStore
	kv    kvStore
	lists map[string]*list.List

	mu sync.RWMutex
}

func NewMemoryStore() *memoryStore {
	var s memoryStore
	s.hash = make(map[string]kvStore)
	s.kv = newKvStore()
	s.lists = make(map[string]*list.List)
	return &s
}

func (s *memoryStore) Get(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if value, exist := s.kv.get(string(key)); exist {
		return []byte(value), nil
	}

	return nil, nil
}

func (s *memoryStore) Set(key, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.kv.set(string(key), string(value))
	return nil
}

func (s *memoryStore) HGet(key, field []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kv, exist := s.hash[string(key)]
	if !exist {
		return nil, nil
	}

	if value, exist := kv.get(string(field)); exist {
		return []byte(value), nil
	}

	return nil, nil
}

func (s *memoryStore) HGetAll(key []byte) ([][]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kv, exist := s.hash[string(key)]
	if !exist {
		return nil, nil
	}

	values := make([][]byte, len(kv.store)*2)
	var i int
	for k, v := range kv.store {
		values[i] = []byte(k)
		values[i+1] = []byte(v)
		i += 2
	}

	return values, nil
}

func (s *memoryStore) HSet(key, field, value []byte) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stringKey := string(key)
	stringField := string(field)

	kv, exist := s.hash[stringKey]
	if !exist {
		kv = newKvStore()
		s.hash[stringKey] = kv
	}

	_, exist = kv.get(stringField)
	kv.set(stringField, string(value))
	return !exist, nil
}

func (s *memoryStore) HIncrBy(key, field []byte, incr int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stringKey := string(key)
	stringField := string(field)

	kv, exist := s.hash[stringKey]
	if !exist {
		kv = newKvStore()
		s.hash[stringKey] = kv
	}

	var int64Value int64
	if value, exist := kv.get(stringField); exist {
		var err error
		int64Value, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, ErrWrongType
		}
	}

	int64Value += incr
	kv.set(stringField, strconv.FormatInt(int64Value, 10))
	return int64Value, nil
}

func (s *memoryStore) RPush(key, value []byte) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stringKey := string(key)

	l, exist := s.lists[stringKey]
	if !exist {
		l = list.New()
		s.lists[stringKey] = l
	}

	l.PushBack(value)

	return int64(l.Len()), nil
}

func (s *memoryStore) LPop(key []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stringKey := string(key)

	l, exist := s.lists[stringKey]
	if !exist || l.Len() == 0 {
		return nil, nil
	}

	return l.Remove(l.Front()).([]byte), nil
}
