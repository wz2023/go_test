package inter

type IMatchMgr interface {
	AddMatch(roomID int32, UserID string) int32 // 根据房间ID,加到对应的房间匹配列表中
	DoMatch()                                   // 没秒钟执行一次匹配检查
	Run()
	Stop()
}
