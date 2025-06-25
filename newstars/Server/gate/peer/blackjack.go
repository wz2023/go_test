package peer

import (
	"github.com/golang/protobuf/proto"
	"newstars/Protocol/plr"
	"newstars/framework/core/session"
	"newstars/framework/glog"
)

// 二十一点.进入房间通知
func (p *LinkServer) onP3010001(data interface{}) {
	glog.SInfof("[onP3010001] 进入房间通知")
	m := &plr.P3010001{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		for _, v := range m.TableInfo.GetPlayers() {
			glog.SInfof("[P3010001] 进入房间通知 msg:%v", m.String())

			s, err := session.GMgr.GetSessionByUserID(v.GetUserID())
			if err != nil {
				glog.SErrorf("err:%v", err)
				continue
			}
			s.Push("P3010001", m)

			// 绑定session
			s.Set(KeyRoomID, m.GetTableInfo().GetRoomID())
			s.Set(KeyTableID, m.GetTableInfo().GetTableID())
			s.Set(KeySeatID, v.GetSeatID())
			s.SetGameStatus(StatusBlackjack)

			ls := GLinkServers.FindLinkByType(TypeBlackjack)
			GGateCore.StoreUIDAndLinkSvrID(v.GetUserID(), ls.GetID())
			ls.Add(s)

			ls = GLinkServers.FindLinkByType(TypeHall)
			n1 := &plr.N0000001{}
			n1.RoomID = m.GetTableInfo().GetRoomID()
			n1.SeatNo = v.GetSeatID()
			n1.TableID = m.GetTableInfo().GetTableID()
			n1.UserID = v.GetUserID()
			ls.Notify("N0000001", n1)

			n3 := &plr.N0000003{}
			n3.RoundName = ""
			n3.TableID = m.GetTableInfo().GetTableID()
			ls.Notify("N0000003", n3)
		}
	}
}

// 二十一点.发牌通知
func (p *LinkServer) onP3010002(data interface{}) {
	m := &plr.P3010002{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3010001", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

// 二十一点.玩家操作通知
func (p *LinkServer) onP3010003(data interface{}) {
	m := &plr.P3010003{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3010003", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

// 二十一点.当前桌子状态通知
func (p *LinkServer) onP3010004(data interface{}) {
	m := &plr.P3010004{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3010004", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}
