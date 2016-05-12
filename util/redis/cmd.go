package redis

import (
	"strconv"
	"strings"
)

const (
	resultSuccess = iota
	resultWrongNumberArguments
	resultWrongType
	resultInternalError
	resultSyntaxError
	resultListFullError
	resultInvalidInteger
)

type cmdFunc func(*Proto, Store) (*Proto, int)

var cmdFuncs map[string]cmdFunc

func init() {
	cmdFuncs = make(map[string]cmdFunc)
	cmdFuncs["get"] = get
	cmdFuncs["set"] = set
	cmdFuncs["hget"] = hget
	cmdFuncs["hgetall"] = hgetall
	cmdFuncs["hset"] = hset
	cmdFuncs["hincrby"] = hincrby

	cmdFuncs["rpush"] = rpush
	cmdFuncs["lpop"] = lpop
}

func processCmd(req *Proto, store Store) *Proto {
	name := req.GetCommandName()
	name = strings.ToLower(name)

	f, ok := cmdFuncs[name]
	if !ok {
		rsp := &Proto{Type: Error}
		rsp.Str = "-ERR unknown command '" + name + "'\r\n"
		return rsp
	}

	rsp, result := f(req, store)
	if result != resultSuccess {
		rsp = &Proto{}
		rsp.Type = Error
		rsp.Str = strError(req, result)
	}

	return rsp
}

func get(req *Proto, store Store) (*Proto, int) {
	if len(req.Elems) != 2 {
		return nil, resultWrongNumberArguments
	}

	value, err := store.Get(req.Elems[1].Raw)
	if err != nil && err != ErrKeyNotExist {
		return nil, resultInternalError
	}

	rsp := &Proto{}
	rsp.Type = BulkString
	rsp.Raw = value
	return rsp, resultSuccess
}

func set(req *Proto, store Store) (*Proto, int) {
	if len(req.Elems) < 3 {
		return nil, resultWrongNumberArguments
	}

	err := store.Set(req.Elems[1].Raw, req.Elems[2].Raw)
	if err != nil {
		return nil, resultInternalError
	}

	rsp := &Proto{}
	rsp.Type = SimpleString
	rsp.Str = "+OK\r\n"
	return rsp, resultSuccess
}

func hget(req *Proto, store Store) (*Proto, int) {
	if len(req.Elems) != 3 {
		return nil, resultWrongNumberArguments
	}

	// field can not empty
	if len(req.Elems[2].Raw) == 0 {
		return nil, resultSyntaxError
	}

	value, err := store.HGet(req.Elems[1].Raw, req.Elems[2].Raw)
	if err != nil {
		return nil, resultInternalError
	}

	rsp := &Proto{}
	rsp.Type = BulkString
	rsp.Raw = value
	return rsp, resultSuccess
}

func hgetall(req *Proto, store Store) (*Proto, int) {
	if len(req.Elems) != 2 {
		return nil, resultWrongNumberArguments
	}

	values, err := store.HGetAll(req.Elems[1].Raw)
	if err != nil {
		return nil, resultInternalError
	}

	rsp := &Proto{}
	rsp.Type = Array
	if len(values) == 0 {
		return rsp, resultSuccess
	}

	rsp.Elems = make([]*Proto, 0, len(values))
	for i := 0; i < len(values); i++ {
		proto := &Proto{}
		proto.Type = BulkString
		proto.Raw = values[i]
		rsp.Elems = append(rsp.Elems, proto)
	}
	return rsp, resultSuccess
}

func hset(req *Proto, store Store) (*Proto, int) {
	if len(req.Elems) != 4 {
		return nil, resultWrongNumberArguments
	}

	// field can not empty
	if len(req.Elems[2].Raw) == 0 {
		return nil, resultSyntaxError
	}

	newInsert, err := store.HSet(req.Elems[1].Raw, req.Elems[2].Raw, req.Elems[3].Raw)
	if err != nil {
		return nil, resultInternalError
	}

	rsp := &Proto{}
	rsp.Type = Integer
	rsp.Int = 0
	if newInsert {
		rsp.Int = 1
	}
	return rsp, resultSuccess
}

func hincrby(req *Proto, store Store) (*Proto, int) {
	if len(req.Elems) != 4 {
		return nil, resultWrongNumberArguments
	}

	// field can not empty
	if len(req.Elems[2].Raw) == 0 {
		return nil, resultSyntaxError
	}

	increment, err := strconv.ParseInt(string(req.Elems[3].Raw), 10, 64)
	if err != nil {
		return nil, resultInvalidInteger
	}

	newInt64, err := store.HIncrBy(req.Elems[1].Raw, req.Elems[2].Raw, increment)
	if err != nil {
		if err == ErrWrongType {
			return nil, resultWrongType
		}
		return nil, resultInternalError
	}

	rsp := &Proto{}
	rsp.Type = Integer
	rsp.Int = int(newInt64)
	return rsp, resultSuccess
}

func rpush(req *Proto, store Store) (*Proto, int) {
	if len(req.Elems) != 3 {
		return nil, resultWrongNumberArguments
	}

	size, err := store.RPush(req.Elems[1].Raw, req.Elems[2].Raw)
	if err != nil {
		return nil, resultInternalError
	}

	rsp := &Proto{}
	rsp.Type = Integer
	rsp.Int = int(size)
	return rsp, resultSuccess
}

func lpop(req *Proto, store Store) (*Proto, int) {
	if len(req.Elems) != 2 {
		return nil, resultWrongNumberArguments
	}

	value, err := store.LPop(req.Elems[1].Raw)
	if err != nil {
		return nil, resultInternalError
	}

	rsp := &Proto{}
	rsp.Type = BulkString
	rsp.Raw = value
	return rsp, resultSuccess
}

func strError(req *Proto, result int) string {
	if result == resultWrongNumberArguments {
		return "-ERR wrong number of arguments for '" + req.GetCommandName() + "' command\r\n"
	} else if result == resultWrongType {
		return "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	} else if result == resultSyntaxError {
		return "-ERR syntax error\r\n"
	} else if result == resultListFullError {
		return "-ERR list full\r\n"
	} else if result == resultInvalidInteger {
		return "-ERR value is not an integer or out of range\r\n"
	}

	return "-ERR internal error\r\n"
}
