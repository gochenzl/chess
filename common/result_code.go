package common

const (
	ResultSuccess = 0
	ResultFail    = 1
)

// for table
const (
	ResultFailInGame        = 10 // 游戏中
	ResultFailInWaiting     = 11 // 配桌中
	ResultFailTableConflict = 12 // 更新桌子信息时发生冲突
	ResultFailTableNotExist = 13 // 桌子不存在
	ResultFailRoomNotExist  = 14 // 房间不存在
)
