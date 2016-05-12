package user

import (
	"github.com/gochenzl/chess/game_ddz/pb_user"
	"github.com/golang/protobuf/proto"
)

func (ui *UserInfo) getInfo(flag int) proto.Message {
	info, present := ui.infos[flag]
	if !present {
		panic(getUserFieldName(flag) + " does not loaded")
	}

	return info
}

func (ui *UserInfo) Userid() uint32 {
	return ui.userid
}

func (ui *UserInfo) Money() int64 {
	if ui.money == nil {
		panic("money does not loaded")
	}

	return *(ui.money)
}

func (ui *UserInfo) IncMoney(inc int64) {
	if inc == 0 {
		return
	}
	*(ui.money) += inc
	ui.moneyInc += inc
}

func (ui *UserInfo) NickName() string {
	info := ui.getInfo(FlagBasicInfo)
	return info.(*pb_user.BasicInfo).Nickname
}

func (ui *UserInfo) SetNickName(nick string) {
	info := ui.getInfo(FlagBasicInfo)
	info.(*pb_user.BasicInfo).Nickname = nick
	ui.modifiedInfos[FlagBasicInfo] = info
}

func (ui *UserInfo) Sex() string {
	info := ui.getInfo(FlagBasicInfo)
	return info.(*pb_user.BasicInfo).Sex
}

func (ui *UserInfo) SetSex(sex string) {
	info := ui.getInfo(FlagBasicInfo)
	info.(*pb_user.BasicInfo).Sex = sex
	ui.modifiedInfos[FlagBasicInfo] = info
}
