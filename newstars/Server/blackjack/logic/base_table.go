package logic

import (
	"newstars/Server/blackjack/inter"
)

// 基础桌子实现
type BaseTable struct {
	id      int32
	gameID  int32
	roomID  int32
	players []inter.IPlayer
	status  int // 0:空闲 1:游戏中
}

func (t *BaseTable) GetPlayers() []inter.IPlayer {

	return t.players
}

func (t *BaseTable) GetPlayerByID(userID string) inter.IPlayer {
	for _, v := range t.players {
		if v.GetUserID() == userID {
			return v
		}
	}
	return nil
}

func (t *BaseTable) GetID() int32 {
	return t.id
}

func (t *BaseTable) GetGameID() int32 {
	return t.gameID
}

func (t *BaseTable) StartGame() {
	t.status = 1
}

func (t *BaseTable) PlayerAction(userID string, action interface{}) error {
	// 基础实现不处理具体操作
	return nil
}

func (t *BaseTable) EndGame() {
	t.status = 0
}
