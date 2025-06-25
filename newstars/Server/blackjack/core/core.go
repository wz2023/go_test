package core

import (
	"newstars/Protocol/plr"
	"newstars/Server/blackjack/inter"
	"newstars/Server/blackjack/logic"
	"newstars/framework/core/component"
	"newstars/framework/core/session"
	"newstars/framework/glog"
	"sync"
	"time"
)

type BlackJackServer struct {
	component.Base
	mgr      map[int32]inter.ITableMgr // key: roomID; value: *tableMgr
	lock     sync.RWMutex
	matchMgr inter.IMatchMgr // 匹配管理器
	s        *session.Session
}

func NewBlackJackServer() *BlackJackServer {
	return &BlackJackServer{
		mgr: make(map[int32]inter.ITableMgr),
	}
}

func (b *BlackJackServer) AfterInit() {
	// 创建匹配管理器
	config := &logic.MatchConfig{
		MatchSize: 1, // 1人一组匹配
		Timeout:   5 * time.Second,
	}
	b.matchMgr = logic.NewMatchMgr(b.MatchCallBack, config)

	b.matchMgr.Run()

	for i := 0; i < 4; i++ {
		roomID := int32(10000 + i)
		b.mgr[roomID] = logic.NewTableMgr(0)
	}
}

// 匹配回调
func (b *BlackJackServer) MatchCallBack(roomID int32, users []string) {
	glog.SInfof("匹配回调: roomID:%v user:%v", roomID, users)

	tableMgr, ok := b.mgr[roomID]
	if !ok {
		glog.SErrorf("not find roomID %v ", roomID)
		return
	}
	table := tableMgr.NewTable(roomID, users)

	players := make([]*plr.PlayerInfo, len(table.GetPlayers()))
	for k, v := range table.GetPlayers() {
		players[k] = &plr.PlayerInfo{
			UserID:  v.GetUserID(),
			Name:    v.GetName(),
			SeatID:  v.GetSeatID(),
			Balance: v.GetBalance(),
		}
	}

	msg := &plr.P3010001{
		TableInfo: &plr.TableInfo{
			RoomID:         roomID,
			TableID:        table.GetID(),
			GameID:         table.GetGameID(),
			DealerID:       table.GetPlayers()[0].GetUserID(),
			Players:        players,
			CurrTableState: plr.TableState_Init,
			CurrOpPlayerID: "",
			CurrOpEndTime:  0,
		},
	}

	b.s.Push("P3010001", msg)
}

// 请求房间列表
func (b *BlackJackServer) C3000001(s *session.Session, msg *plr.C3000001, mid uint) error {
	glog.SInfof("请求房间列表: %v", msg.String())

	rsp := &plr.S3000001{}

	rsp.Rooms = append(rsp.Rooms, &plr.S3000001_RoomInfo{
		RoomID:         10000,
		BaseAmount:     10,
		MinEnterAmount: 1,
		MaxEnterAmount: 1000,
		RoomName:       "初级场",
		MinRatio:       1,
		MaxRatio:       3,
	})

	return s.Response(rsp, mid)
}

// 请求匹配
func (b *BlackJackServer) C3000002(s *session.Session, msg *plr.C3000002, mid uint) error {
	glog.SInfof("请求匹配: %v", msg.String())

	b.s = s

	rsp := &plr.S3000002{}

	rsp.RetCode = b.matchMgr.AddMatch(msg.RoomID, msg.GetUserID())

	return s.Response(rsp, mid)
}

// 玩家操作请求
func (b *BlackJackServer) C3010002(s *session.Session, msg *plr.C3010002, mid uint) error {
	glog.SInfof("玩家操作请求: %v", msg.String())

	rsp := &plr.S3010002{}

	return s.Response(rsp, mid)
}

// C3010003 玩家重进
func (b *BlackJackServer) C3010003(s *session.Session, msg *plr.C3010003, mid uint) error {
	glog.SInfof("玩家重进: %v", msg.String())

	rsp := &plr.S3010003{
		RetCode:   0,
		TableInfo: nil,
	}

	return s.Response(rsp, mid)
}
