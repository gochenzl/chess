package user

import (
	"strconv"

	"github.com/gochenzl/chess/game_ddz/pb_user"
	"github.com/gochenzl/chess/util/log"
	"github.com/golang/protobuf/proto"
	"gopkg.in/redis.v3"
)

// user flag define
const (
	FlagBasicInfo = 1
	FlagMoney     = 2
	FlagItem      = 88
)

const existFieldName = "exist"

type UserInfo struct {
	userid uint32
	money  *int64
	infos  map[int]proto.Message
	items  map[int]int

	// for update
	modifiedInfos map[int]proto.Message
	moneyInc      int64
	itemIncs      map[int]int
	newGuy        bool
}

var userFieldNames map[int]string
var AllUserFlags []int

func init() {
	userFieldNames = make(map[int]string)
	userFieldNames[FlagBasicInfo] = "basic"
	userFieldNames[FlagMoney] = "money"

	AllUserFlags = append(AllUserFlags, FlagBasicInfo)
	AllUserFlags = append(AllUserFlags, FlagMoney)
	AllUserFlags = append(AllUserFlags, FlagItem)
}

func getUserFieldName(flag int) string {
	if name, ok := userFieldNames[flag]; ok {
		return name
	}

	return ""
}

func decodeUserField(flag int, value []byte) proto.Message {
	var info proto.Message

	switch flag {
	case FlagBasicInfo:
		info = &pb_user.BasicInfo{}

	default:
		log.Warn("unknown user flag %d", flag)
		return nil
	}

	if len(value) == 0 {
		return info
	}

	if err := proto.Unmarshal(value, info); err != nil {
		log.Warn("Unmarshal %s fail:%s", proto.MessageName(info), err.Error())
		return nil
	}

	return info
}

func genUserKey(userid uint32) string {
	return "user_" + strconv.FormatUint(uint64(userid), 10)
}

func genItemKey(userid uint32) string {
	return genUserKey(userid) + "_item"
}

func newUserInfo(userid uint32) *UserInfo {
	ui := &UserInfo{userid: userid}
	ui.infos = make(map[int]proto.Message)
	ui.modifiedInfos = make(map[int]proto.Message)
	return ui
}

func NewUser(userid uint32) *UserInfo {
	ui := newUserInfo(userid)
	ui.newGuy = true

	var money int64
	ui.money = &money

	for _, flag := range AllUserFlags {
		if flag == FlagMoney || flag == FlagItem {
			continue
		}

		ui.infos[flag] = decodeUserField(flag, nil)
	}

	return ui
}

func LoadUserInfo(userid uint32, flags []int) *UserInfo {
	userKey := genUserKey(userid)

	pipeline := dbClient.Pipeline()
	pipeline.HGet(userKey, existFieldName)

	for i := 0; i < len(flags); i++ {
		if flags[i] == FlagItem {
			pipeline.HGetAll(genItemKey(userid))
			continue
		}

		field := getUserFieldName(flags[i])
		if len(field) == 0 {
			log.Warn("unknown user flag:%d", flags[i])
			continue
		}

		pipeline.HGet(userKey, field)
	}

	cmds, err := pipeline.Exec()
	if err != nil && err != redis.Nil {
		log.Error("load user info fail:%s", err.Error())
		return nil
	}

	if len(cmds) != len(flags)+1 {
		log.Error("load user info fail")
		return nil
	}

	var userExist bool
	if value, _ := cmds[0].(*redis.StringCmd).Bytes(); value != nil {
		userExist = true
	}

	if !userExist {
		return nil
	}

	ui := newUserInfo(userid)
	for i := 0; i < len(flags); i++ {
		if flags[i] == FlagMoney {
			if !ui.loadMoney(cmds[i+1]) {
				return nil
			}
		} else if flags[i] == FlagItem {
			if !ui.loadItem(cmds[i+1]) {
				return nil
			}
		} else {
			if !ui.loadOtherInfo(flags[i], cmds[i+1]) {
				return nil
			}
		}

	}

	return ui
}

func (ui *UserInfo) loadMoney(cmd redis.Cmder) bool {
	value := cmd.(*redis.StringCmd).Val()
	if len(value) == 0 {
		var money int64
		ui.money = &money
		return true
	}

	money, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Warn("decode money fail:%s", err.Error())
		return false
	}
	ui.money = &money

	return true
}

func (ui *UserInfo) loadItem(cmd redis.Cmder) bool {
	values := cmd.(*redis.StringSliceCmd).Val()
	ui.items = make(map[int]int)

	for i := 0; i < len(values); i += 2 {
		itemid, err := strconv.Atoi(values[i])
		if err != nil {
			log.Warn("decode itemid fail:%s", err.Error())
			return false
		}

		number, err := strconv.Atoi(values[i+1])
		if err != nil {
			log.Warn("decode item number:%s", err.Error())
			return false
		}

		ui.items[itemid] = number
	}

	return true
}

func (ui *UserInfo) loadOtherInfo(flag int, cmd redis.Cmder) bool {
	value, err := cmd.(*redis.StringCmd).Bytes()
	if err != nil && err != redis.Nil {
		log.Error("load user info fail:%s", err.Error())
		return false
	}

	info := decodeUserField(flag, value)
	if info == nil {
		return false
	}
	ui.infos[flag] = info

	return true
}

func (ui *UserInfo) isModified() bool {
	return len(ui.modifiedInfos) != 0 || ui.moneyInc != 0 ||
		len(ui.itemIncs) != 0 || ui.newGuy == true
}

func (ui *UserInfo) Save() bool {
	if !ui.isModified() {
		return true
	}

	userKey := genUserKey(ui.userid)
	pipeline := dbClient.Pipeline()

	if ui.newGuy {
		pipeline.HSet(userKey, existFieldName, "a")

	}

	if ui.moneyInc != 0 {
		pipeline.HIncrBy(userKey, getUserFieldName(FlagMoney), ui.moneyInc)
	}

	if ui.itemIncs != nil {
		itemKey := genItemKey(ui.userid)
		for itemid, inc := range ui.itemIncs {
			pipeline.HIncrBy(itemKey, strconv.Itoa(itemid), int64(inc))
		}
	}

	for flag, info := range ui.modifiedInfos {
		fieldName := getUserFieldName(flag)
		if len(fieldName) == 0 {
			log.Warn("get user field name fail")
			return false
		}

		value, _ := proto.Marshal(info)
		pipeline.HSet(userKey, fieldName, string(value))
	}

	_, err := pipeline.Exec()
	if err != nil {
		log.Error("save user info fail:%s, userid = %d", err.Error(), ui.userid)
		return false
	}

	ui.moneyInc = 0
	ui.newGuy = false
	ui.itemIncs = nil
	ui.modifiedInfos = make(map[int]proto.Message)

	return true
}
