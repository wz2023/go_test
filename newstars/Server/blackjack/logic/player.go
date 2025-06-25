package logic

import (
	"newstars/Server/blackjack/inter"
	"newstars/framework/model"
)

var _ inter.IPlayer = &Player{}

// 玩家信息
type Player struct {
	table    inter.ITable
	uid      string
	sid      int32
	faceID   int32
	name     string
	nick     string
	currency string
	balance  int32
}

func (p *Player) GetUserID() string {
	return p.uid
}

func (p *Player) GetName() string {
	return p.name
}

func (p *Player) GetSeatID() int32 {
	return p.sid
}

func (p *Player) GetBalance() int32 {
	return p.balance
}

func NewPlayer(t inter.ITable, info *model.UserInfo, seatID int32) *Player {
	return &Player{
		table:    t,
		uid:      info.UserID,
		sid:      seatID,
		faceID:   info.FaceFrameID,
		name:     info.DisPlayName,
		nick:     info.NickName,
		currency: info.Currency,
		balance:  int32(info.Wealth),
	}
}
