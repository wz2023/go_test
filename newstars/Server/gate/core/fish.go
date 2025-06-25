package core

import (
	"newstars/Protocol/plr"
	"newstars/Server/gate/peer"
	"newstars/framework/core/session"
	"newstars/framework/glog"

	"github.com/golang/protobuf/proto"
)

// C3080001 请求捕鱼游戏房间
func (g *GateCore) C3080001(s *session.Session, msg *plr.C3080001, mid uint) error {
	ls := peer.GLinkServers.FindLinkByType(peer.TypeFish)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeFish)
		m := &plr.S3080001{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls.Request("C3080001", msg, func(data interface{}) {
		m := &plr.S3080001{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
	return nil
}

// C3080002  进入房间
func (g *GateCore) C3080002(s *session.Session, msg *plr.C3080002, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		m := &plr.S3080002{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls := peer.GLinkServers.FindLinkByType(peer.TypeFish)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeFish)
		m := &plr.S3080002{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls.Request("C3080002", msg, func(data interface{}) {
		m := &plr.S3080002{}
		proto.Unmarshal(data.([]byte), m)

		if m.GetRetCode() == 0 {
			s.Set(peer.KeyRoomID, m.GetRoomID())
			s.Set(peer.KeyTableID, m.GetTableID())
			s.Set(peer.KeySeatID, m.GetSeatNo())
			s.SetGameStatus(peer.StatusFish)
			g.StoreUIDAndLinkSvrID(s.UID(), ls.GetID())
			ls.Add(s)

			//通知大厅
			hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
			n1 := &plr.N0000001{}
			n1.RoomID = m.GetRoomID()
			n1.SeatNo = m.GetSeatNo()
			n1.TableID = m.GetTableID()
			n1.UserID = s.String(peer.KeyUserID)
			if hall != nil {
				hall.Notify("N0000001", n1)
			}

			//进入房间即认为游戏已经开始
			n6 := &plr.N0000006{}
			n6.UserID = msg.GetUserID()
			n6.RoomID = m.GetRoomID()
			n6.TableID = m.GetTableID()
			n6.SeatNo = m.GetSeatNo()
			if hall != nil {
				hall.Notify("N0000006", n6)
			}
		}

		err := s.Response(m, mid)
		if err != nil {
			glog.SErrorf("Response error:%v", err)
			ls.Leave(s)
			return
		}
	})
	return nil
}

// C3080003 炮台种类
func (g *GateCore) C3080003(s *session.Session, msg *plr.C3080003, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}
	ls := peer.GLinkServers.FindLinkByType(peer.TypeFish)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeFish)
		m := &plr.S3080003{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls.Request("C3080003", msg, func(data interface{}) {
		m := &plr.S3080003{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
	return nil
}

// C3080004 图鉴
func (g *GateCore) C3080004(s *session.Session, msg *plr.C3080004, mid uint) error {
	ls := peer.GLinkServers.FindLinkByType(peer.TypeFish)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeFish)
		m := &plr.S3080004{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls.Request("C3080004", msg, func(data interface{}) {
		m := &plr.S3080004{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
	return nil
}

// C3080005 使用技能
func (g *GateCore) C3080005(s *session.Session, msg *plr.C3080005, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}

	ls := peer.GLinkServers.FindLinkByType(peer.TypeFish)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeFish)
		m := &plr.S3080005{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls.Request("C3080005", msg, func(data interface{}) {
		m := &plr.S3080005{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
	return nil
}

// C3080006 购买炮台
func (g *GateCore) C3080006(s *session.Session, msg *plr.C3080006, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}

	ls := peer.GLinkServers.FindLinkByType(peer.TypeFish)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeFish)
		m := &plr.S3080006{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls.Request("C3080006", msg, func(data interface{}) {
		m := &plr.S3080006{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
	return nil
}

// C3080007 装载炮台
func (g *GateCore) C3080007(s *session.Session, msg *plr.C3080007, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}
	ls := peer.GLinkServers.FindLinkByType(peer.TypeFish)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeFish)
		m := &plr.S3080007{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls.Request("C3080007", msg, func(data interface{}) {
		m := &plr.S3080007{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
	return nil
}

// C3080008 用户重入
func (g *GateCore) C3080008(s *session.Session, msg *plr.C3080008, mid uint) error {
	v, ok := g.uids.Load(s.UID())
	if !ok {
		m := &plr.S3080008{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}
	linkID := v.(int)
	ls := peer.GLinkServers.FindLinkByID(linkID) //待处理
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeFish)
		m := &plr.S3080008{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls.Request("C3080008", msg, func(data interface{}) {
		m := &plr.S3080008{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
		if m.GetRetCode() == 0 {
			s.Set(peer.KeyUserID, msg.GetUserID())
			s.Set(peer.KeySeatID, msg.GetSeatNo())
			s.Set(peer.KeyTableID, msg.GetTableID())
			s.Set(peer.KeyRoomID, msg.GetRoomID())
			s.SetGameStatus(peer.StatusFish)
			ls.Add(s)
		}
	})
	return nil
}

// N3080001 用户离座
func (g *GateCore) N3080001(s *session.Session, msg *plr.N3080001, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}
	ls := peer.GLinkServers.FindLinkByUID(peer.TypeFish, s.UID())
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
		return nil
	}
	s.Set(peer.KeyLeaveType, int32(1))
	ls.Leave(s)
	return nil
}

// N3080002 发射子弹
func (g *GateCore) N3080002(s *session.Session, msg *plr.N3080002, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}

	ls := peer.GLinkServers.FindLinkByUID(peer.TypeFish, s.UID())
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
		return nil
	}

	err := ls.Notify("N3080002", msg)
	if err != nil {
		glog.SErrorf("Notify err:%v", err)
		return nil
	}
	return nil
}

// N3080003 子弹打中鱼
func (g *GateCore) N3080003(s *session.Session, msg *plr.N3080003, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}

	ls := peer.GLinkServers.FindLinkByUID(peer.TypeFish, s.UID())
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("Invalid UserID.Not Register.%s", s.UID())
		return nil
	}

	err := ls.Notify("N3080003", msg)
	if err != nil {
		glog.SErrorf("Notify err:%v", err)
		return nil
	}
	return nil
}

// N3080004 切换炮台倍率
func (g *GateCore) N3080004(s *session.Session, msg *plr.N3080004, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}

	ls := peer.GLinkServers.FindLinkByUID(peer.TypeFish, s.UID())
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
		return nil
	}

	err := ls.Notify("N3080004", msg)
	if err != nil {
		glog.SErrorf("Notify err:%v", err)
		return nil
	}
	return nil
}
