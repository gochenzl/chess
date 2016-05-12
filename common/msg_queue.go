package common

import "strconv"

type LoginInfo struct {
	Token    string `json:"token"`
	Nickname string `json:"nickname"`
}

const TableTimeoutList = "table_timeout_list"
const TableNewList = "table_new_list"

func GenLoginInfoKey(userid uint32) string {
	return "login_info_" + strconv.FormatUint(uint64(userid), 10)
}

func GenGateQueueKey(gateid uint32) string {
	return "gate_msg_queue" + strconv.FormatUint(uint64(gateid), 10)
}
