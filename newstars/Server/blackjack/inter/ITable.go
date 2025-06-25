package inter

type ITable interface {
	GetID() int32 // 获取桌子ID
	GetGameID() int32
	StartGame()                                           // 开始游戏
	PlayerAction(userID string, action interface{}) error // 玩家操作
	EndGame()                                             // 结束游戏
	GetPlayers() []IPlayer
	GetPlayerByID(userID string) IPlayer
}
