package core

import (
	"net"
	"newstars/Protocol/plr"
	"newstars/Server/gate/peer"
	"newstars/framework/core/component"
	"newstars/framework/core/gate"
	"newstars/framework/core/session"
	"newstars/framework/glog"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

// GateCore component
type GateCore struct {
	component.Base
	timer *gate.Timer
	uids  sync.Map // uid<->linkSvrID
}

const (
	kickOutType = iota
	mutipleType
)

// ServerIDS list
type ServerIDS []int

func (ss ServerIDS) contains(id int) bool {
	for _, v := range ss {
		if v == id {
			return true
		}
	}
	return false
}

// NewGateCore returns a new GateCore
func NewGateCore() *GateCore {
	g := &GateCore{}
	return g
}

// Init init core
func (g *GateCore) Init() {
	peer.GLinkServers.Go()
	peer.GGateCore = g
}

// AfterInit component lifetime callback
func (g *GateCore) AfterInit() {
	gate.OnSessionClosed(func(s *session.Session) {
		lks := peer.GLinkServers.FindLinksByUID(s.UID())
		for _, v := range lks {
			v.Leave(s)
		}
	})
	g.timer = gate.NewTimer(15*time.Second, func() {
		//var hl *peer.LinkServer
		//push := &explr.P1010003{}
		//var i int32
		//for i = 0; i < peer.GGameNumber; i++ {
		//	counts := &explr.P1010003_GameCounts{}
		//	counts.GameID = i
		//	for _, v := range peer.GLinkServers {
		//		if v.GetType() == peer.TypeHall {
		//			hl = v
		//			counts.Name = v.GetType()
		//			counts.Counts = v.GetCounts(i)
		//		} else {
		//			counts.Rooms = append(counts.Rooms, v.PeopleCounting(i)...)
		//		}
		//	}
		//	push.Counts = append(push.Counts, counts)
		//}
		//
		//if hl != nil {
		//	hl.PushCounts(push)
		//}
	})

	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
		glog.SErrorf("N0000010 request ignore")
		return
	}
	//msg := &plr.N0000010{}
	//ls.Notify("N0000010", msg)

}

func (g *GateCore) StoreUIDAndLinkSvrID(uid string, linkSvrID int) {
	g.uids.Store(uid, linkSvrID)
}

// 玩家是否在玩人人类游戏
func inPlayerGame(s *session.Session) bool {
	status := s.GameStatus()
	if status == peer.StatusFish {
		return true
	}
	return false
}

// C0000001 登录请求
//func (g *GateCore) C0000001(s *session.Session, msg *plr.C0000001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	if msg.GetIPAddr() == "" {
//		ip, err := net.ResolveTCPAddr("tcp", s.RemoteAddr().String())
//		if err == nil {
//			msg.IPAddr = ip.IP.String()
//		}
//	}
//
//	ls.Request("C0000001", msg, func(data interface{}) {
//		m := &plr.S0000001{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			uid := m.GetUserID()
//			os, err := ls.GetSession(uid)
//			if err == nil {
//				push := &plr.P1000008{}
//				push.UserID = uid
//				push.Type = mutipleType
//				os.Push("P1000008", push)
//				for _, v := range peer.GLinkServers.FindLinksByUID(uid) {
//					if v.GetID() != 1 {
//						v.Leave(os)
//					}
//				}
//				os.Clear()
//				ls.Leave(os)
//			}
//			s.Bind(uid)
//			s.Set(peer.KeyGameID, msg.GetGameID())
//			s.Set(peer.KeyUserID, uid)
//			s.SetGameStatus(peer.StatusHall)
//			ls.Add(s)
//		}
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("Response error:%v", err)
//			ls.Leave(s)
//		}
//	})
//	return nil
//}

// C0000002 游戏种类请求
//func (g *GateCore) C0000002(s *session.Session, msg *plr.C0000002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &plr.S0000002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000002", msg, func(data interface{}) {
//		m := &plr.S0000002{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000003 获取用户信息
func (g *GateCore) C0000003(s *session.Session, msg *plr.C0000003, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		m := &plr.S0000003{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("Invalid UserID.Not Register.%s", s.UID())
		m := &plr.S0000003{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls.Request("C0000003", msg, func(data interface{}) {
		m := &plr.S0000003{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
	return nil
}

// C0000005 绑定手机
//func (g *GateCore) C0000005(s *session.Session, msg *plr.C0000005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S0000005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000005", msg, func(data interface{}) {
//		m := &plr.S0000005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000017 获取用户财富信息
//func (g *GateCore) C0000017(s *session.Session, msg *plr.C0000017, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000017{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S0000017{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000017", msg, func(data interface{}) {
//		m := &plr.S0000017{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000016 获取验证码
//func (g *GateCore) C0000016(s *session.Session, msg *plr.C0000016, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000016{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000016", msg, func(data interface{}) {
//		m := &plr.S0000016{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000015 手机登陆
//func (g *GateCore) C0000015(s *session.Session, msg *plr.C0000015, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000015{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	if msg.GetIPAddr() == "" {
//		ip, err := net.ResolveTCPAddr("tcp", s.RemoteAddr().String())
//		if err == nil {
//			msg.IPAddr = ip.IP.String()
//		}
//	}
//
//	ls.Request("C0000015", msg, func(data interface{}) {
//		m := &plr.S0000015{}
//		proto.Unmarshal(data.([]byte), m)
//
//		if m.GetRetCode() == 0 {
//			uid := m.GetUserID()
//			os, err := ls.GetSession(uid)
//			if err == nil {
//				push := &plr.P1000008{}
//				push.UserID = uid
//				push.Type = mutipleType
//				os.Push("P1000008", push)
//				for _, v := range peer.GLinkServers.FindLinksByUID((uid) {
//					if v.GetID() != 1 {
//						v.Leave(os)
//					}
//				}
//				os.Clear()
//				ls.Leave(os)
//			}
//			s.Bind(uid)
//			s.Set(peer.KeyGameID, msg.GetGameID())
//			s.Set(peer.KeyUserID, uid)
//			s.SetGameStatus(peer.StatusHall)
//			ls.Add(s)
//
//			// s.Bind(int64(m.GetUserID()))
//			// s.Set(peer.KeyUserID, m.GetUserID())
//			// ls.Add(s)
//		}
//
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}

// C0000014 代充列表
//func (g *GateCore) C0000014(s *session.Session, msg *plr.C0000014, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000014{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S0000014{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000014", msg, func(data interface{}) {
//		m := &plr.S0000014{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000013 兑换
//func (g *GateCore) C0000013(s *session.Session, msg *plr.C0000013, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000013{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S0000013{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	//玩人人游戏时不允许兑换
//	if inPlayerGame(s) {
//		glog.SErrorf("Invalid game status .%d", s.GameStatus())
//		m := &plr.S0000013{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000013", msg, func(data interface{}) {
//		m := &plr.S0000013{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010014 兑换成功
//func (g *GateCore) C1010014(s *session.Session, msg *explr.C1010014, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010014{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypePay)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid Link server")
//		m := &explr.S1010014{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010014", msg, func(data interface{}) {
//		m := &explr.S1010014{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010015 兑换失败
//func (g *GateCore) C1010015(s *session.Session, msg *explr.C1010015, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010015{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypePay)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid link server")
//		m := &explr.S1010015{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010015", msg, func(data interface{}) {
//		m := &explr.S1010015{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010017 冻结状态下设置金额
//func (g *GateCore) C1010017(s *session.Session, msg *explr.C1010017, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010017{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010017{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010017", msg, func(data interface{}) {
//		m := &explr.S1010017{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010019 客户端邮件推送
// func (g *GateCore) C1010019(s *session.Session, msg *explr.C1010019, mid uint) error {
// 	if msg.GetUID() != s.String(peer.KeyUserID) {
// 		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
// 		m := &explr.S1010019{}
// 		m.RetCode = invalidUserid
// 		return s.Response(m, mid)
// 	}

// 	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
// 	if ls == nil || ls.GetConnector() == nil {
// 		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
// 		m := &explr.S1010019{}
// 		m.RetCode = invalidUserid
// 		return s.Response(m, mid)
// 	}

// 	ls.Request("C1010019", msg, func(data interface{}) {
// 		m := &explr.S1010019{}
// 		proto.Unmarshal(data.([]byte), m)
// 		s.Response(m, mid)
// 	})
// 	return nil
// }

// C1010020  充值开关配置
//func (g *GateCore) C1010020(s *session.Session, msg *explr.C1010020, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010020{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypePay)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid Link server")
//		m := &explr.S1010020{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010020", msg, func(data interface{}) {
//		m := &explr.S1010020{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010021 兑换配置
//func (g *GateCore) C1010021(s *session.Session, msg *explr.C1010021, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010021{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010021{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010021", msg, func(data interface{}) {
//		m := &explr.S1010021{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010022 设置保险箱金额
//func (g *GateCore) C1010022(s *session.Session, msg *explr.C1010022, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010022{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010022{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010022", msg, func(data interface{}) {
//		m := &explr.S1010022{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010023 修改玩家平台ID
//func (g *GateCore) C1010023(s *session.Session, msg *explr.C1010023, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010023{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010023{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010023", msg, func(data interface{}) {
//		m := &explr.S1010023{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010024 后台发送邮件
//func (g *GateCore) C1010024(s *session.Session, msg *explr.C1010024, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010024{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010024{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010024", msg, func(data interface{}) {
//		m := &explr.S1010024{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010025 后台发送邮件
//func (g *GateCore) C1010025(s *session.Session, msg *explr.C1010025, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010025{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010025{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010025", msg, func(data interface{}) {
//		m := &explr.S1010025{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1010026 游戏分平台配置更新通知
//func (g *GateCore) C1010026(s *session.Session, msg *explr.C1010026, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010026{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//	kindid := msg.GetKindID()
//	linkType, ok := peer.GameKindMap[kindid]
//	if !ok {
//		glog.SErrorf("Invalid gamekindid :%v", kindid)
//		m := &explr.S1010026{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls := peer.GLinkServers.FindLinkByType(linkType)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid Link server linkType:%v", linkType)
//		m := &explr.S1010026{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010026", msg, func(data interface{}) {
//		m := &explr.S1010026{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// 绑定手机
//func (g *GateCore) C1010027(s *session.Session, msg *explr.C1010027, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010023{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010027{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010027", msg, func(data interface{}) {
//		m := &explr.S1010027{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

//func (g *GateCore) C1010028(s *session.Session, msg *explr.C1010028, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010028{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010028{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010028", msg, func(data interface{}) {
//		m := &explr.S1010028{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000012 变更用户昵称
//func (g *GateCore) C0000012(s *session.Session, msg *plr.C0000012, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000012{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S0000012{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000012", msg, func(data interface{}) {
//		m := &plr.S0000012{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000011 保险箱存取
//func (g *GateCore) C0000011(s *session.Session, msg *plr.C0000011, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000011{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S0000011{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	//玩人人游戏时不允许存取保险箱
//	if inPlayerGame(s) {
//		glog.SErrorf("Invalid game status .%d", s.GameStatus())
//		m := &plr.S0000011{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000011", msg, func(data interface{}) {
//		m := &plr.S0000011{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000018 Token登陆
func (g *GateCore) C0000018(s *session.Session, msg *plr.C0000018, mid uint) error {

	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
		m := &plr.S0000018{}
		m.RetCode = errFindLinkServer
		return s.Response(m, mid)
	}

	if msg.GetIPAddr() == "" {
		ip, err := net.ResolveTCPAddr("tcp", s.RemoteAddr().String())
		if err == nil {
			msg.IPAddr = ip.IP.String()
		}
	}

	ls.Request("C0000018", msg, func(data interface{}) {
		m := &plr.S0000018{}
		proto.Unmarshal(data.([]byte), m)
		if m.GetRetCode() == 0 {
			uid := m.GetUserID()
			os, err := ls.GetSession(uid)
			if err == nil {
				push := &plr.P1000008{}
				push.UserID = uid
				push.Type = mutipleType
				os.Push("P1000008", push)
				for _, v := range peer.GLinkServers.FindLinksByUID(uid) {
					if v.GetID() != 1 {
						v.Leave(os)
					}
				}
				os.Clear()
				ls.Leave(os)
			}
			s.Set(peer.KeyGameID, msg.GetGameID())
			s.Set(peer.KeyUserID, uid)
			s.SetGameStatus(peer.StatusHall)
			s.Set(peer.KeyUserToken, msg.GetToken())
			session.GMgr.Bind(uid, s)
			ls.Add(s)

			// s.Bind(int64(m.GetUserID()))
			// s.Set(peer.KeyUserID, m.GetUserID())
			// ls.Add(s)
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

// C0000019 check valiad
//func (g *GateCore) C0000019(s *session.Session, msg *plr.C0000019, mid uint) error {
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000019{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000019", msg, func(data interface{}) {
//		m := &plr.S0000019{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000020 check valiad
//func (g *GateCore) C0000020(s *session.Session, msg *plr.C0000020, mid uint) error {
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000020{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000020", msg, func(data interface{}) {
//		m := &plr.S0000020{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000021 用户支付信息
//func (g *GateCore) C0000021(s *session.Session, msg *plr.C0000021, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000021{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000021", msg, func(data interface{}) {
//		m := &plr.S0000021{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000022 设置保险箱密码
//func (g *GateCore) C0000022(s *session.Session, msg *plr.C0000022, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000022{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000022", msg, func(data interface{}) {
//		m := &plr.S0000022{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000023 设置保险箱密码
//func (g *GateCore) C0000023(s *session.Session, msg *plr.C0000023, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000023{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000023", msg, func(data interface{}) {
//		m := &plr.S0000023{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C1070007 审核状态
//func (g *GateCore) C1070007(s *session.Session, msg *plr.C1070007, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S1070007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1070007", msg, func(data interface{}) {
//		m := &plr.S1070007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000026 设置邮件已读
//func (g *GateCore) C0000026(s *session.Session, msg *plr.C0000026, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000026{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000026", msg, func(data interface{}) {
//		m := &plr.S0000026{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000025 获取用户邮件列表
//func (g *GateCore) C0000025(s *session.Session, msg *plr.C0000025, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000025{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000025", msg, func(data interface{}) {
//		m := &plr.S0000025{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000027 兑换限制
//func (g *GateCore) C0000027(s *session.Session, msg *plr.C0000027, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000027{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000027{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000027", msg, func(data interface{}) {
//		m := &plr.S0000027{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000028 找回账号
//func (g *GateCore) C0000028(s *session.Session, msg *plr.C0000028, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000028{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C0000028", msg, func(data interface{}) {
//		m := &plr.S0000028{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000029 获取兑换记录
//func (g *GateCore) C0000029(s *session.Session, msg *plr.C0000029, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000029{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C0000029", msg, func(data interface{}) {
//		m := &plr.S0000029{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000030 获取VIP等级
//func (g *GateCore) C0000030(s *session.Session, msg *plr.C0000030, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000030{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C0000030", msg, func(data interface{}) {
//		m := &plr.S0000030{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000031 获取服务器维护状态
func (g *GateCore) C0000031(s *session.Session, msg *plr.C0000031, mid uint) error {
	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
		m := &plr.S0000031{}
		m.RetCode = errFindLinkServer
		return s.Response(m, mid)
	}
	ls.Request("C0000031", msg, func(data interface{}) {
		m := &plr.S0000031{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
	return nil
}

// C0000032 获取白名单状态
//func (g *GateCore) C0000032(s *session.Session, msg *plr.C0000032, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//		m := &plr.S0000032{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C0000032", msg, func(data interface{}) {
//		m := &plr.S0000032{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000033 邮件领取金币
//func (g *GateCore) C0000033(s *session.Session, msg *plr.C0000033, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000033{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S0000033{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000033", msg, func(data interface{}) {
//		m := &plr.S0000033{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000035 盈利榜
//func (g *GateCore) C0000035(s *session.Session, msg *plr.C0000035, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000035{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S0000035{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000035", msg, func(data interface{}) {
//		m := &plr.S0000035{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000036 幸运榜
//func (g *GateCore) C0000036(s *session.Session, msg *plr.C0000036, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000036{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S0000036{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000036", msg, func(data interface{}) {
//		m := &plr.S0000036{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000037 VIP等级配置
//func (g *GateCore) C0000037(s *session.Session, msg *plr.C0000037, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid Servers.Not Register.%d", s.UID())
//		m := &plr.S0000037{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000037", msg, func(data interface{}) {
//		m := &plr.S0000037{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000038 获取投注记录
//func (g *GateCore) C0000038(s *session.Session, msg *plr.C0000038, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid Servers.Not Register.%d", s.UID())
//		m := &plr.S0000038{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000038", msg, func(data interface{}) {
//		m := &plr.S0000038{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

//// C3010001 请求斗地主游戏房间
//func (g *GateCore) C3010001(s *session.Session, msg *plr.C3010001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeLandlord)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeLandlord)
//		m := &plr.S3010001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3010001", msg, func(data interface{}) {
//		m := &plr.S3010001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3010002  进入房间
//func (g *GateCore) C3010002(s *session.Session, msg *plr.C3010002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3010002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeLandlord)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeLandlord)
//		m := &plr.S3010002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	if s.GameStatus() != peer.StatusHall {
//		glog.SErrorf("Game status error.%v", s.GameStatus())
//		m := &plr.S3010002{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//	ls.Request("C3010002", msg, func(data interface{}) {
//		m := &plr.S3010002{}
//		proto.Unmarshal(data.([]byte), m)
//
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, m.GetRoomID())
//			s.Set(peer.KeyTableID, m.GetTableID())
//			s.Set(peer.KeySeatID, m.GetSeatNo())
//			s.SetGameStatus(peer.StatusLandload)
//			g.uids.Store(s.UID(), ls.GetID())
//			ls.Add(s)
//
//			//通知大厅
//			hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			n1 := &plr.N0000001{}
//			n1.RoomID = m.GetRoomID()
//			n1.SeatNo = m.GetSeatNo()
//			n1.TableID = m.GetTableID()
//			n1.UserID = s.String(peer.KeyUserID)
//			if hall != nil {
//				hall.Notify("N0000001", n1)
//			}
//		}
//
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3010003 用户通知准备开始游戏
//func (g *GateCore) C3010003(s *session.Session, msg *plr.C3010003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3010003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeLandlord, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3010003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3010003", msg, func(data interface{}) {
//		m := &plr.S3010003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3010004 用户请求桌台信息
//func (g *GateCore) C3010004(s *session.Session, msg *plr.C3010004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3010004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeLandlord, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3010004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3010004", msg, func(data interface{}) {
//		m := &plr.S3010004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3010005 用户叫分
//func (g *GateCore) C3010005(s *session.Session, msg *plr.C3010005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3010005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeLandlord, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3010005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3010005", msg, func(data interface{}) {
//		m := &plr.S3010005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3010006 用户出牌
//func (g *GateCore) C3010006(s *session.Session, msg *plr.C3010006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3010006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeLandlord, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3010006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3010006", msg, func(data interface{}) {
//		m := &plr.S3010006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3010008 托管
//func (g *GateCore) C3010008(s *session.Session, msg *plr.C3010008, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3010008{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeLandlord, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3010008{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3010008", msg, func(data interface{}) {
//		m := &plr.S3010008{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3010009 农民加倍
//func (g *GateCore) C3010009(s *session.Session, msg *plr.C3010009, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3010009{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeLandlord, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3010009{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3010009", msg, func(data interface{}) {
//		m := &plr.S3010009{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3010010 用户离开
//func (g *GateCore) C3010010(s *session.Session, msg *plr.C3010010, mid uint) error {
//	rsp := &plr.S3010010{}
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return s.Response(rsp, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeLandlord, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return s.Response(rsp, mid)
//	}
//
//	ls.Request("C3010010", msg, func(data interface{}) {
//		m := &plr.S3010010{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//		ls.GroupLeave(s)
//		if m.GetRetCode() == 0 {
//			s.SetGameStatus(peer.StatusHall)
//			hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			if hall == nil {
//				glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//				return
//			}
//			m := &plr.N0000002{}
//			m.UserID = s.String(peer.KeyUserID)
//			hall.Notify("N0000002", m)
//		}
//	})
//
//	return nil
//}
//
//// C3010007 用户重入
//func (g *GateCore) C3010007(s *session.Session, msg *plr.C3010007, mid uint) error {
//	v, ok := g.uids.Load(s.UID())
//	if !ok {
//		m := &plr.S3010007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//	linkID := v.(int)
//	ls := peer.GLinkServers.FindLinkByID(linkID) //待处理
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeLandlord)
//		m := &plr.S3010007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3010007", msg, func(data interface{}) {
//		m := &plr.S3010007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyUserID, msg.GetUserID())
//			s.Set(peer.KeySeatID, msg.GetSeatNo())
//			s.Set(peer.KeyTableID, msg.GetTableID())
//			s.Set(peer.KeyRoundName, msg.GetRoundName())
//			s.Set(peer.KeyRoomID, msg.GetRoomID())
//			s.SetGameStatus(peer.StatusLandload)
//			ls.Add(s)
//		}
//	})
//	return nil
//}
//
//// N3010002 用户离开
//func (g *GateCore) N3010002(s *session.Session, msg *plr.N3010002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeLandlord, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//
//	hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if hall == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	m := &plr.N0000002{}
//	m.UserID = s.String(peer.KeyUserID)
//	hall.Notify("N0000002", m)
//	return nil
//}

// C0000004 游客注册
//func (g *GateCore) C0000004(s *session.Session, msg *plr.C0000004, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeHall)
//		m := &plr.S0000004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	// if msg.GetIPAddr() == "" {
//	// 	ip, err := net.ResolveTCPAddr("tcp", s.RemoteAddr().String())
//	// 	if err == nil {
//	// 		msg.IPAddr = ip.IP.String()
//	// 	}
//	// }
//
//	ls.Request("C0000004", msg, func(data interface{}) {
//		m := &plr.S0000004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000006 注册
//func (g *GateCore) C0000006(s *session.Session, msg *plr.C0000006, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeHall)
//		m := &plr.S0000006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ip, err := net.ResolveTCPAddr("tcp", s.RemoteAddr().String())
//	if err == nil {
//		msg.IPAddr = ip.IP.String()
//	}
//
//	ls.Request("C0000006", msg, func(data interface{}) {
//		m := &plr.S0000006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000007 账户绑定支付宝
//func (g *GateCore) C0000007(s *session.Session, msg *plr.C0000007, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeHall)
//		m := &plr.S0000007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000007", msg, func(data interface{}) {
//		m := &plr.S0000007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C0000008 账户绑定银行卡
//func (g *GateCore) C0000008(s *session.Session, msg *plr.C0000008, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeHall)
//		m := &plr.S0000008{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000008", msg, func(data interface{}) {
//		m := &plr.S0000008{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// N0000005 用户退出
func (g *GateCore) N0000005(s *session.Session, msg *plr.N0000005, mid uint) error {
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		return nil
	}

	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
		return nil
	}
	return ls.Notify("N0000005", msg)
}

// N0000008 获取手机信息
//func (g *GateCore) N0000008(s *session.Session, msg *plr.N0000008, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Link Server is not exsit.")
//		return nil
//	}
//	return ls.Notify("N0000008", msg)
//}

// N0000009 客户端操作日志打印(苹果提炸弹金花审核包时加的日志打印，已经没用了)
//func (g *GateCore) N0000009(s *session.Session, msg *plr.N0000009, mid uint) error {
//	return nil
//}

// C0000009 用户更新头像
//func (g *GateCore) C0000009(s *session.Session, msg *plr.C0000009, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S0000009{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &plr.S0000009{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C0000009", msg, func(data interface{}) {
//		m := &plr.S0000009{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C000000A 用户退出大厅
//func (g *GateCore) C000000A(s *session.Session, msg *plr.C000000A, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S000000A{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &plr.S000000A{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C000000A", msg, func(data interface{}) {
//		m := &plr.S000000A{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//		ls.Leave(s)
//	})
//	return nil
//}

//// C3020001 请求炸金花游戏房间
//func (g *GateCore) C3020001(s *session.Session, msg *plr.C3020001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeThreecard)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeThreecard)
//		m := &plr.S3020001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3020001", msg, func(data interface{}) {
//		m := &plr.S3020001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3020002 请求进入炸金花游戏房间
//func (g *GateCore) C3020002(s *session.Session, msg *plr.C3020002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3020002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeThreecard)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeThreecard)
//		m := &plr.S3020002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	if s.GameStatus() != peer.StatusHall {
//		glog.SErrorf("Game status error.%v", s.GameStatus())
//		m := &plr.S3020002{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//	ls.Request("C3020002", msg, func(data interface{}) {
//		m := &plr.S3020002{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, m.GetRoomID())
//			s.Set(peer.KeyTableID, m.GetTableID())
//			s.Set(peer.KeySeatID, m.GetSeatNo())
//			s.SetGameStatus(peer.StatusThreeCard)
//			g.uids.Store(s.UID(), ls.GetID())
//			ls.Add(s)
//			//通知大厅
//			// hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			// n1 := &plr.N0000001{}
//			// n1.RoomID = m.GetRoomID()
//			// n1.SeatNo = m.GetSeatNo()
//			// n1.TableID = m.GetTableID()
//			// n1.UserID = s.String(peer.KeyUserID)
//			// if hall != nil {
//			// 	hall.Notify("N0000001", n1)
//			// }
//		}
//
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("C3020002 Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3020004 玩家弃牌
//func (g *GateCore) C3020004(s *session.Session, msg *plr.C3020004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3020004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeThreecard)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeThreecard)
//		m := &plr.S3020004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3020004", msg, func(data interface{}) {
//		m := &plr.S3020004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3020005 玩家看牌
//func (g *GateCore) C3020005(s *session.Session, msg *plr.C3020005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3020005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeThreecard)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeThreecard)
//		m := &plr.S3020005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3020005", msg, func(data interface{}) {
//		m := &plr.S3020005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3020006 玩家下注
//func (g *GateCore) C3020006(s *session.Session, msg *plr.C3020006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3020006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeThreecard)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeThreecard)
//		m := &plr.S3020006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3020006", msg, func(data interface{}) {
//		m := &plr.S3020006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3020007 玩家比牌
//func (g *GateCore) C3020007(s *session.Session, msg *plr.C3020007, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3020007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeThreecard)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeThreecard)
//		m := &plr.S3020007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3020007", msg, func(data interface{}) {
//		m := &plr.S3020007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3020009 玩家全压
//func (g *GateCore) C3020009(s *session.Session, msg *plr.C3020009, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3020009{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeThreecard)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeThreecard)
//		m := &plr.S3020009{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3020009", msg, func(data interface{}) {
//		m := &plr.S3020009{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3020010 游戏等待时间
//func (g *GateCore) C3020010(s *session.Session, msg *plr.C3020010, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeThreecard)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeThreecard)
//		m := &plr.S3020010{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3020010", msg, func(data interface{}) {
//		m := &plr.S3020010{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// N3020001 用户离开炸金花
//func (g *GateCore) N3020001(s *session.Session, msg *plr.C3020007, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeThreecard, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//	return nil
//}
//
//// C3030001  请求百人牛牛房间
//func (g *GateCore) C3030001(s *session.Session, msg *plr.C3030001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeOx100)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeOx100)
//		m := &plr.S3030001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3030001", msg, func(data interface{}) {
//		m := &plr.S3030001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3030002  请求进入牛牛房间
//func (g *GateCore) C3030002(s *session.Session, msg *plr.C3030002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3030002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeOx100)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeOx100)
//		m := &plr.S3030002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	if s.GameStatus() != peer.StatusHall {
//		glog.SErrorf("Game status error.%v", s.GameStatus())
//		m := &plr.S3030002{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//	ls.Request("C3030002", msg, func(data interface{}) {
//		m := &plr.S3030002{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, m.GetRoomID())
//			s.Set(peer.KeyTableID, m.GetTableID())
//			s.Set(peer.KeySeatID, m.GetSeatNo())
//			s.Set(peer.KeyRoundName, m.GetRoundName())
//			s.SetGameStatus(peer.StatusOx100)
//			g.uids.Store(s.UID(), ls.GetID())
//			ls.Add(s)
//			//通知大厅
//			// hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			// n1 := &plr.N0000001{}
//			// n1.RoomID = m.GetRoomID()
//			// n1.SeatNo = m.GetSeatNo()
//			// n1.TableID = m.GetTableID()
//			// n1.UserID = s.String(peer.KeyUserID)
//			// if hall != nil {
//			// 	hall.Notify("N0000001", n1)
//			// }
//		}
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("C3030002 Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3030003 牛牛下注
//func (g *GateCore) C3030003(s *session.Session, msg *plr.C3030003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3030003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3030003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3030003", msg, func(data interface{}) {
//		m := &plr.S3030003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3030004 top10
//func (g *GateCore) C3030004(s *session.Session, msg *plr.C3030004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3030004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3030004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3030004", msg, func(data interface{}) {
//		m := &plr.S3030004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3030005 历史记录
//func (g *GateCore) C3030005(s *session.Session, msg *plr.C3030005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3030005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3030005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3030005", msg, func(data interface{}) {
//		m := &plr.S3030005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3030006 申请上庄
//func (g *GateCore) C3030006(s *session.Session, msg *plr.C3030006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3030006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3030006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3030006", msg, func(data interface{}) {
//		m := &plr.S3030006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3030007 获取上庄列表
//func (g *GateCore) C3030007(s *session.Session, msg *plr.C3030007, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3030007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3030007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3030007", msg, func(data interface{}) {
//		m := &plr.S3030007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3030008 获取旁注列表
//func (g *GateCore) C3030008(s *session.Session, msg *plr.C3030008, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3030008{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3030008{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3030008", msg, func(data interface{}) {
//		m := &plr.S3030008{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3030009 申请下庄
//func (g *GateCore) C3030009(s *session.Session, msg *plr.C3030009, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3030009{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3030009{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3030009", msg, func(data interface{}) {
//		m := &plr.S3030009{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3030010 玩家列表
//func (g *GateCore) C3030010(s *session.Session, msg *plr.C3030010, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3030010{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3030010{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3030010", msg, func(data interface{}) {
//		m := &plr.S3030010{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// N3030001 离开牛牛
//func (g *GateCore) N3030001(s *session.Session, msg *plr.N3030001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//	return nil
//}
//
//// C3040001 请求房间
//func (g *GateCore) C3040001(s *session.Session, msg *plr.C3040001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeRedBlack)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeRedBlack)
//		m := &plr.S3040001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3040001", msg, func(data interface{}) {
//		m := &plr.S3040001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3040002 进入房间
//func (g *GateCore) C3040002(s *session.Session, msg *plr.C3040002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3040002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeRedBlack)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeRedBlack)
//		m := &plr.S3040002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	if s.GameStatus() != peer.StatusHall {
//		glog.SErrorf("Game status error.%v", s.GameStatus())
//		m := &plr.S3040002{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//	ls.Request("C3040002", msg, func(data interface{}) {
//		m := &plr.S3040002{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, m.GetRoomID())
//			s.Set(peer.KeyTableID, m.GetTableID())
//			s.Set(peer.KeySeatID, m.GetSeatNo())
//			s.Set(peer.KeyRoundName, m.GetRoundName())
//			g.uids.Store(s.UID(), ls.GetID())
//			s.SetGameStatus(peer.StatusRedBlack)
//			ls.Add(s)
//			//通知大厅
//			// hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			// n1 := &plr.N0000001{}
//			// n1.RoomID = m.GetRoomID()
//			// n1.SeatNo = m.GetSeatNo()
//			// n1.TableID = m.GetTableID()
//			// n1.UserID = s.String(peer.KeyUserID)
//			// if hall != nil {
//			// 	hall.Notify("N0000001", n1)
//			// }
//		}
//
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("C3040002 Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3040003 下注
//func (g *GateCore) C3040003(s *session.Session, msg *plr.C3040003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3040003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeRedBlack, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3040003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3040003", msg, func(data interface{}) {
//		m := &plr.S3040003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3040004 20局记录
//func (g *GateCore) C3040004(s *session.Session, msg *plr.C3040004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3040004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeRedBlack, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3040004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3040004", msg, func(data interface{}) {
//		m := &plr.S3040004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3040005 入座列表
//func (g *GateCore) C3040005(s *session.Session, msg *plr.C3040005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3040005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeRedBlack, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3040005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3040005", msg, func(data interface{}) {
//		m := &plr.S3040005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3040006 旁注列表
//func (g *GateCore) C3040006(s *session.Session, msg *plr.C3040006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3040006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeRedBlack, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3040006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3040006", msg, func(data interface{}) {
//		m := &plr.S3040006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3040007 玩家列表
//func (g *GateCore) C3040007(s *session.Session, msg *plr.C3040007, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3040007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeRedBlack, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3040007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3040007", msg, func(data interface{}) {
//		m := &plr.S3040007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// N3040001 用户离座
//func (g *GateCore) N3040001(s *session.Session, msg *plr.N3040001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeRedBlack, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//	return nil
//}
//
//// C3050001 请求房间
//func (g *GateCore) C3050001(s *session.Session, msg *plr.C3050001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeOx5)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeOx5)
//		m := &plr.S3050001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3050001", msg, func(data interface{}) {
//		m := &plr.S3050001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3050002 进入房间
//func (g *GateCore) C3050002(s *session.Session, msg *plr.C3050002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3050002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeOx5)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeOx5)
//		m := &plr.S3050002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	if s.GameStatus() != peer.StatusHall {
//		glog.SErrorf("Game status error.%v", s.GameStatus())
//		m := &plr.S3050002{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//	ls.Request("C3050002", msg, func(data interface{}) {
//		m := &plr.S3050002{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, m.GetRoomID())
//			s.Set(peer.KeyTableID, m.GetTableID())
//			s.Set(peer.KeySeatID, m.GetSeatNo())
//			s.SetGameStatus(peer.StatusOx5)
//			g.uids.Store(s.UID(), ls.GetID())
//			ls.Add(s)
//			//通知大厅
//			hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			n1 := &plr.N0000006{}
//			n1.RoomID = m.GetRoomID()
//			n1.SeatNo = m.GetSeatNo()
//			n1.TableID = m.GetTableID()
//			n1.UserID = s.String(peer.KeyUserID)
//			if hall != nil {
//				hall.Notify("N0000006", n1)
//			}
//		}
//
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("C3050002 Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3050003 抢庄
//func (g *GateCore) C3050003(s *session.Session, msg *plr.C3050003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3050003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx5, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3050003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3050003", msg, func(data interface{}) {
//		m := &plr.S3050003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3050004 加倍
//func (g *GateCore) C3050004(s *session.Session, msg *plr.C3050004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3050004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx5, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3050004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3050004", msg, func(data interface{}) {
//		m := &plr.S3050004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3050005 重入房间
//func (g *GateCore) C3050005(s *session.Session, msg *plr.C3050005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3050005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeOx5)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid Type.%v", peer.TypeOx5)
//		m := &plr.S3050005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3050005", msg, func(data interface{}) {
//		m := &plr.S3050005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, msg.GetRoomID())
//			s.Set(peer.KeySeatID, msg.GetSeatNo())
//			s.Set(peer.KeyTableID, msg.GetTableID())
//			s.SetGameStatus(peer.StatusOx5)
//			ls.Add(s)
//		}
//	})
//	return nil
//}
//
//// C3050006 凑牌
//func (g *GateCore) C3050006(s *session.Session, msg *plr.C3050006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3050006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx5, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3050006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3050006", msg, func(data interface{}) {
//		m := &plr.S3050006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3050007 游戏等待时间
//func (g *GateCore) C3050007(s *session.Session, msg *plr.C3050007, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeOx5)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3050007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3050007", msg, func(data interface{}) {
//		m := &plr.S3050007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3050008 玩家退出
//func (g *GateCore) C3050008(s *session.Session, msg *plr.C3050008, mid uint) error {
//	rsp := &plr.S3050008{}
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		s.Response(rsp, mid)
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx5, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		s.Response(rsp, mid)
//		return nil
//	}
//
//	ls.Request("C3050008", msg, func(data interface{}) {
//		m := &plr.S3050008{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//		if m.GetRetCode() == 0 {
//			ls.GroupLeave(s)
//			s.SetGameStatus(peer.StatusHall)
//			hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			if hall == nil {
//				glog.SErrorf("C3050008 Invalid UserID.Not Register.%d", s.UID())
//				return
//			}
//			m := &plr.N0000007{}
//			m.UserID = s.String(peer.KeyUserID)
//			hall.Notify("N0000007", m)
//		}
//	})
//	return nil
//}
//
//// N3050001 用户离座
//func (g *GateCore) N3050001(s *session.Session, msg *plr.N3050001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx5, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//
//	hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if hall == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	m := &plr.N0000007{}
//	m.UserID = s.String(peer.KeyUserID)
//	hall.Notify("N0000007", m)
//	return nil
//}
//
//// C3060001 请求房间
//func (g *GateCore) C3060001(s *session.Session, msg *plr.C3060001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeDT)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeDT)
//		m := &plr.S3060001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3060001", msg, func(data interface{}) {
//		m := &plr.S3060001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3060002 进入房间
//func (g *GateCore) C3060002(s *session.Session, msg *plr.C3060002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3060002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeDT)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeDT)
//		m := &plr.S3060002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	if s.GameStatus() != peer.StatusHall {
//		glog.SErrorf("Game status error.%v", s.GameStatus())
//		m := &plr.S3060002{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//	ls.Request("C3060002", msg, func(data interface{}) {
//		m := &plr.S3060002{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, m.GetRoomID())
//			s.Set(peer.KeyTableID, m.GetTableID())
//			s.Set(peer.KeySeatID, m.GetSeatNo())
//			s.SetGameStatus(peer.StatusDT)
//			g.uids.Store(s.UID(), ls.GetID())
//			ls.Add(s)
//		}
//
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("C3060002 Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3060003 下注
//func (g *GateCore) C3060003(s *session.Session, msg *plr.C3060003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3060003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDT, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3060003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3060003", msg, func(data interface{}) {
//		m := &plr.S3060003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3060004 历史记录
//func (g *GateCore) C3060004(s *session.Session, msg *plr.C3060004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3060004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDT, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3060004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3060004", msg, func(data interface{}) {
//		m := &plr.S3060004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3060005 获取入座列表
//func (g *GateCore) C3060005(s *session.Session, msg *plr.C3060005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3060005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDT, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3060005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3060005", msg, func(data interface{}) {
//		m := &plr.S3060005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3060006 获取入座列表
//func (g *GateCore) C3060006(s *session.Session, msg *plr.C3060006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3060006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDT, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3060006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3060006", msg, func(data interface{}) {
//		m := &plr.S3060006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3060007 玩家列表
//func (g *GateCore) C3060007(s *session.Session, msg *plr.C3060007, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3060007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDT, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3060007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3060007", msg, func(data interface{}) {
//		m := &plr.S3060007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// N3060001 用户离座
//func (g *GateCore) N3060001(s *session.Session, msg *plr.N3060001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDT, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//	return nil
//}
//
//// C3070001  请求奔驰宝马房间
//func (g *GateCore) C3070001(s *session.Session, msg *plr.C3070001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeBenz)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBenz)
//		m := &plr.S3070001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3070001", msg, func(data interface{}) {
//		m := &plr.S3070001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3070002  请求进入奔驰宝马房间
//func (g *GateCore) C3070002(s *session.Session, msg *plr.C3070002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3070002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeBenz)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBenz)
//		m := &plr.S3070002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	if s.GameStatus() != peer.StatusHall {
//		glog.SErrorf("Game status error.%v", s.GameStatus())
//		m := &plr.S3070002{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//	ls.Request("C3070002", msg, func(data interface{}) {
//		m := &plr.S3070002{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, m.GetRoomID())
//			s.Set(peer.KeyTableID, m.GetTableID())
//			s.Set(peer.KeySeatID, m.GetSeatNo())
//			s.Set(peer.KeyRoundName, m.GetRoundName())
//			s.SetGameStatus(peer.StatusBenz)
//			g.uids.Store(s.UID(), ls.GetID())
//			ls.Add(s)
//		}
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("C3070002 Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3070003 奔驰宝马下注
//func (g *GateCore) C3070003(s *session.Session, msg *plr.C3070003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3070003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3070003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3070003", msg, func(data interface{}) {
//		m := &plr.S3070003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3070004 历史记录
//func (g *GateCore) C3070004(s *session.Session, msg *plr.C3070004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3070004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3070004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3070004", msg, func(data interface{}) {
//		m := &plr.S3070004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3070005 获取旁注列表
//func (g *GateCore) C3070005(s *session.Session, msg *plr.C3070005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3070005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3070005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3070005", msg, func(data interface{}) {
//		m := &plr.S3070005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3070006 申请上庄
//func (g *GateCore) C3070006(s *session.Session, msg *plr.C3070006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3070006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3070006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3070006", msg, func(data interface{}) {
//		m := &plr.S3070006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3070007 获取上庄列表
//func (g *GateCore) C3070007(s *session.Session, msg *plr.C3070007, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3070007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3070007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3070007", msg, func(data interface{}) {
//		m := &plr.S3070007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3070008 申请下庄
//func (g *GateCore) C3070008(s *session.Session, msg *plr.C3070008, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3070008{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3070008{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3070008", msg, func(data interface{}) {
//		m := &plr.S3070008{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3070009 获取玩家列表
//func (g *GateCore) C3070009(s *session.Session, msg *plr.C3070009, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3070009{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3070009{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3070009", msg, func(data interface{}) {
//		m := &plr.S3070009{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// N3070001 离开奔驰宝马
//func (g *GateCore) N3070001(s *session.Session, msg *plr.N3070001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//	return nil
//}
//
//// C3090001 请求房间
//func (g *GateCore) C3090001(s *session.Session, msg *plr.C3090001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypePoker13)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypePoker13)
//		m := &plr.S3090001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3090001", msg, func(data interface{}) {
//		m := &plr.S3090001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3090002 进入房间
//func (g *GateCore) C3090002(s *session.Session, msg *plr.C3090002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3090002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypePoker13)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypePoker13)
//		m := &plr.S3090002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3090002", msg, func(data interface{}) {
//		m := &plr.S3090002{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, m.GetRoomID())
//			s.Set(peer.KeyTableID, m.GetTableID())
//			s.Set(peer.KeySeatID, m.GetSeatNo())
//			g.uids.Store(s.UID(), ls.GetID())
//			ls.Add(s)
//			//通知大厅
//			hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			n1 := &plr.N0000006{}
//			n1.RoomID = m.GetRoomID()
//			n1.SeatNo = m.GetSeatNo()
//			n1.TableID = m.GetTableID()
//			n1.UserID = s.String(peer.KeyUserID)
//			if hall != nil {
//				hall.Notify("N0000006", n1)
//			}
//		}
//
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("C3090002 Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3090003 准备
//func (g *GateCore) C3090003(s *session.Session, msg *plr.C3090003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3090003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypePoker13, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3090003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3090003", msg, func(data interface{}) {
//		m := &plr.S3090003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3090005 重入房间
//func (g *GateCore) C3090005(s *session.Session, msg *plr.C3090005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3090005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypePoker13)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid Type.%v", peer.TypePoker13)
//		m := &plr.S3090005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3090005", msg, func(data interface{}) {
//		m := &plr.S3090005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, msg.GetRoomID())
//			s.Set(peer.KeySeatID, msg.GetSeatNo())
//			s.Set(peer.KeyTableID, msg.GetTableID())
//			ls.Add(s)
//		}
//	})
//	return nil
//}
//
//// C3090006 出牌
//func (g *GateCore) C3090006(s *session.Session, msg *plr.C3090006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3090006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypePoker13, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3090006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3090006", msg, func(data interface{}) {
//		m := &plr.S3090006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// N3090001 用户离座
//func (g *GateCore) N3090001(s *session.Session, msg *plr.N3090001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypePoker13, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//
//	hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if hall == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	m := &plr.N0000007{}
//	m.UserID = s.String(peer.KeyUserID)
//	hall.Notify("N0000007", m)
//	return nil
//}
//
//// C3100001 请求房间
//func (g *GateCore) C3100001(s *session.Session, msg *plr.C3100001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeBaccarat)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBaccarat)
//		m := &plr.S3100001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3100001", msg, func(data interface{}) {
//		m := &plr.S3100001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3100002 进入房间
//func (g *GateCore) C3100002(s *session.Session, msg *plr.C3100002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3100002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeBaccarat)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBaccarat)
//		m := &plr.S3100002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	if s.GameStatus() != peer.StatusHall {
//		glog.SErrorf("Game status error.%v", s.GameStatus())
//		m := &plr.S3100002{}
//		m.RetCode = errStatus
//		return s.Response(m, mid)
//	}
//	ls.Request("C3100002", msg, func(data interface{}) {
//		m := &plr.S3100002{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, m.GetRoomID())
//			s.Set(peer.KeyTableID, m.GetTableID())
//			s.Set(peer.KeySeatID, m.GetSeatNo())
//			s.Set(peer.KeyRoundName, m.GetRoundName())
//			s.SetGameStatus(peer.StatusBaccarat)
//			g.uids.Store(s.UID(), ls.GetID())
//			ls.Add(s)
//			//通知大厅
//			// hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			// n1 := &plr.N0000001{}
//			// n1.RoomID = m.GetRoomID()
//			// n1.SeatNo = m.GetSeatNo()
//			// n1.TableID = m.GetTableID()
//			// n1.UserID = s.String(peer.KeyUserID)
//			// if hall != nil {
//			// 	hall.Notify("N0000001", n1)
//			// }
//		}
//
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("C3100002 Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3100003 下注
//func (g *GateCore) C3100003(s *session.Session, msg *plr.C3100003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3100003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBaccarat, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3100003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3100003", msg, func(data interface{}) {
//		m := &plr.S3100003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3100004 20局记录
//func (g *GateCore) C3100004(s *session.Session, msg *plr.C3100004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3100004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBaccarat, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3100004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3100004", msg, func(data interface{}) {
//		m := &plr.S3100004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3100005 入座列表
//func (g *GateCore) C3100005(s *session.Session, msg *plr.C3100005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3100005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBaccarat, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3100005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3100005", msg, func(data interface{}) {
//		m := &plr.S3100005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3100006 旁注列表
//func (g *GateCore) C3100006(s *session.Session, msg *plr.C3100006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3100006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBaccarat, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3100006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3100006", msg, func(data interface{}) {
//		m := &plr.S3100006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3100007 玩家列表
//func (g *GateCore) C3100007(s *session.Session, msg *plr.C3100007, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3100007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBaccarat, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3100007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3100007", msg, func(data interface{}) {
//		m := &plr.S3100007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// N3100001 用户离座
//func (g *GateCore) N3100001(s *session.Session, msg *plr.N3100001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBaccarat, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//	return nil
//}
//
//// C3110001 基础配置
//func (g *GateCore) C3110005(s *session.Session, msg *plr.C3110005, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeDuofu)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeDuofu)
//		m := &plr.S3110005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C3110005", msg, func(data interface{}) {
//		m := &plr.S3110005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3110001 进入房间
//func (g *GateCore) C3110001(s *session.Session, msg *plr.C3110001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3110001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeDuofu)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeDuofu)
//		m := &plr.S3110001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	// if s.GameStatus() != peer.StatusHall {
//	// 	glog.SErrorf("Game status error.%v", s.GameStatus())
//	// 	m := &plr.S3110001{}
//	// 	m.RetCode = errStatus
//	// 	return s.Response(m, mid)
//	// }
//	ls.Request("C3110001", msg, func(data interface{}) {
//		m := &plr.S3110001{}
//		proto.Unmarshal(data.([]byte), m)
//		if m.GetRetCode() == 0 {
//			s.Set(peer.KeyRoomID, msg.GetRoomID())
//			s.SetGameStatus(peer.StatusDuofu)
//			g.uids.Store(s.UID(), ls.GetID())
//			ls.Add(s)
//			//通知大厅
//			// hall := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//			// n1 := &plr.N0000001{}
//			// n1.RoomID = m.GetRoomID()
//			// n1.SeatNo = m.GetSeatNo()
//			// n1.TableID = m.GetTableID()
//			// n1.UserID = s.String(peer.KeyUserID)
//			// if hall != nil {
//			// 	hall.Notify("N0000001", n1)
//			// }
//		}
//
//		err := s.Response(m, mid)
//		if err != nil {
//			glog.SErrorf("C3110001 Response error:%v", err)
//			ls.Leave(s)
//			return
//		}
//	})
//	return nil
//}
//
//// C3110002 旋转
//func (g *GateCore) C3110002(s *session.Session, msg *plr.C3110002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3110002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDuofu, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3110002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3110002", msg, func(data interface{}) {
//		m := &plr.S3110002{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3110003 免费游戏旋转
//func (g *GateCore) C3110003(s *session.Session, msg *plr.C3110003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3110003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDuofu, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3110003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3110003", msg, func(data interface{}) {
//		m := &plr.S3110003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C3110004 奖池翻牌
//func (g *GateCore) C3110004(s *session.Session, msg *plr.C3110004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S3110004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDuofu, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S3110004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//	ls.Request("C3110004", msg, func(data interface{}) {
//		m := &plr.S3110004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// N3110001 用户离座
//func (g *GateCore) N3110001(s *session.Session, msg *plr.N3110001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		return nil
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDuofu, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		return nil
//	}
//	ls.Leave(s)
//	return nil
//}
//
//// C1010001 超端登陆
//func (g *GateCore) C1010001(s *session.Session, msg *explr.C1010001, mid uint) error {
//	//ls := peer.GLinkServers.FindLinkByType(peer.TypeHall)
//	//if ls == nil || ls.GetConnector() == nil {
//	//	glog.SErrorf("FindLinkByType %v Failed.", peer.TypeHall)
//	//	m := &explr.S1010001{}
//	//	m.RetCode = errFindLinkServer
//	//	return s.Response(m, mid)
//	//}
//	//
//	//ls.Request("C1010001", msg, func(data interface{}) {
//	//	m := &explr.S1010001{}
//	//	proto.Unmarshal(data.([]byte), m)
//	//	if m.GetRetCode() == 0 {
//	//		s.Bind(m.GetUID())
//	//		s.Set(peer.KeyUserID, m.GetUID())
//	//		ls.Add(s)
//	//	}
//	//
//	//	err := s.Response(m, mid)
//	//	if err != nil {
//	//		glog.SErrorf("Response error:%v", err)
//	//		ls.Leave(s)
//	//		return
//	//	}
//	//})
//	return nil
//}
//
//// C1010010 获取百人游戏房间信息
//func (g *GateCore) C1010010(s *session.Session, msg *explr.C1010010, mid uint) error {
//	//if msg.GetUID() != s.String(peer.KeyUserID) {
//	//	glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//	//	m := &explr.S1010010{}
//	//	m.RetCode = invalidUserid
//	//	return s.Response(m, mid)
//	//}
//	//
//	//ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	//if ls == nil || ls.GetConnector() == nil {
//	//	glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//	//	m := &explr.S1010010{}
//	//	m.RetCode = invalidUserid
//	//	return s.Response(m, mid)
//	//}
//	//
//	//ls.Request("C1010010", msg, func(data interface{}) {
//	//	m := &explr.S1010010{}
//	//	proto.Unmarshal(data.([]byte), m)
//	//	s.Response(m, mid)
//	//})
//	return nil
//}
//
//// C1010011 获取桌台信息
//func (g *GateCore) C1010011(s *session.Session, msg *explr.C1010011, mid uint) error {
//	//if msg.GetUID() != s.String(peer.KeyUserID) {
//	//	glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//	//	m := &explr.S1010011{}
//	//	m.RetCode = invalidUserid
//	//	return s.Response(m, mid)
//	//}
//	//
//	//ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	//if ls == nil || ls.GetConnector() == nil {
//	//	glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//	//	m := &explr.S1010011{}
//	//	m.RetCode = invalidUserid
//	//	return s.Response(m, mid)
//	//}
//	//
//	//ls.Request("C1010011", msg, func(data interface{}) {
//	//	m := &explr.S1010011{}
//	//	proto.Unmarshal(data.([]byte), m)
//	//	s.Response(m, mid)
//	//})
//	return nil
//}
//
//// C1010002 支付开关
//func (g *GateCore) C1010002(s *session.Session, msg *explr.C1010002, mid uint) error {
//	//if msg.GetUID() != s.String(peer.KeyUserID) {
//	//	glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//	//	m := &explr.S1010002{}
//	//	m.RetCode = invalidUserid
//	//	return s.Response(m, mid)
//	//}
//	//
//	//ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	//if ls == nil || ls.GetConnector() == nil {
//	//	glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//	//	m := &explr.S1010002{}
//	//	m.RetCode = invalidUserid
//	//	return s.Response(m, mid)
//	//}
//	//
//	//ls.Request("C1010002", msg, func(data interface{}) {
//	//	m := &explr.S1010002{}
//	//	proto.Unmarshal(data.([]byte), m)
//	//	s.Response(m, mid)
//	//})
//	return nil
//}

// C1010003 兑换开关
// func (g *GateCore) C1010003(s *session.Session, msg *explr.C1010003, mid uint) error {
// 	if msg.GetUID() != s.String(peer.KeyUserID) {
// 		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
// 		m := &explr.S1010003{}
// 		m.RetCode = invalidUserid
// 		return s.Response(m, mid)
// 	}

// 	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
// 	if ls == nil || ls.GetConnector() == nil {
// 		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
// 		m := &explr.S1010003{}
// 		m.RetCode = invalidUserid
// 		return s.Response(m, mid)
// 	}

// 	ls.Request("C1010003", msg, func(data interface{}) {
// 		m := &explr.S1010003{}
// 		proto.Unmarshal(data.([]byte), m)
// 		s.Response(m, mid)
// 	})
// 	return nil
// }

// C1010004 充值渠道
// func (g *GateCore) C1010004(s *session.Session, msg *explr.C1010004, mid uint) error {
// 	if msg.GetUID() != s.String(peer.KeyUserID) {
// 		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
// 		m := &explr.S1010004{}
// 		m.RetCode = invalidUserid
// 		return s.Response(m, mid)
// 	}

// 	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
// 	if ls == nil || ls.GetConnector() == nil {
// 		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
// 		m := &explr.S1010004{}
// 		m.RetCode = invalidUserid
// 		return s.Response(m, mid)
// 	}

// 	ls.Request("C1010004", msg, func(data interface{}) {
// 		m := &explr.S1010004{}
// 		proto.Unmarshal(data.([]byte), m)
// 		s.Response(m, mid)
// 	})
// 	return nil
// }

// C1010005 订单扣除
// func (g *GateCore) C1010005(s *session.Session, msg *explr.C1010005, mid uint) error {
// 	if msg.GetUID() != s.String(peer.KeyUserID) {
// 		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUID())
// 		m := &explr.S1010005{}
// 		m.RetCode = invalidUserid
// 		return s.Response(m, mid)
// 	}

// 	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
// 	if ls == nil || ls.GetConnector() == nil {
// 		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
// 		m := &explr.S1010005{}
// 		m.RetCode = invalidUserid
// 		return s.Response(m, mid)
// 	}

// 	ls.Request("C1010005", msg, func(data interface{}) {
// 		m := &explr.S1010005{}
// 		proto.Unmarshal(data.([]byte), m)
// 		s.Response(m, mid)
// 	})
// 	return nil
// }

//// C1010006 公告管理
//func (g *GateCore) C1010006(s *session.Session, msg *explr.C1010006, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010006", msg, func(data interface{}) {
//		m := &explr.S1010006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C1010007 跑马灯管理
//func (g *GateCore) C1010007(s *session.Session, msg *explr.C1010007, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010007", msg, func(data interface{}) {
//		m := &explr.S1010007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C1010008 玩家充值
//func (g *GateCore) C1010008(s *session.Session, msg *explr.C1010008, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010008{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010008{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010008", msg, func(data interface{}) {
//		m := &explr.S1010008{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C1010009 玩家充值
//func (g *GateCore) C1010009(s *session.Session, msg *explr.C1010009, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1010009{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S1010009{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1010009", msg, func(data interface{}) {
//		m := &explr.S1010009{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C101000C 退单申请
//func (g *GateCore) C101000C(s *session.Session, msg *explr.C101000C, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S101000C{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S101000C{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C101000C", msg, func(data interface{}) {
//		m := &explr.S101000C{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C101000D 充值确认
//func (g *GateCore) C101000D(s *session.Session, msg *explr.C101000D, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S101000D{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S101000D{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C101000D", msg, func(data interface{}) {
//		m := &explr.S101000D{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C101000F 退单确认
//func (g *GateCore) C101000F(s *session.Session, msg *explr.C101000F, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S101000F{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S101000F{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C101000F", msg, func(data interface{}) {
//		m := &explr.S101000F{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C101000A 踢玩家下线
//func (g *GateCore) C101000A(s *session.Session, msg *explr.C101000A, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S101000A{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S101000A{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C101000A", msg, func(data interface{}) {
//		m := &explr.S101000A{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C101000B 冻结玩家
//func (g *GateCore) C101000B(s *session.Session, msg *explr.C101000B, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S101000B{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeHall, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Login.%d", s.UID())
//		m := &explr.S101000B{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C101000B", msg, func(data interface{}) {
//		m := &explr.S101000B{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C1020001 开始订阅百人牛牛通知消息
//func (g *GateCore) C1020001(s *session.Session, msg *explr.C1020001, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1020001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeOx100)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeOx100)
//		m := &explr.S1020001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1020001", msg, func(data interface{}) {
//		m := &explr.S1020001{}
//		proto.Unmarshal(data.([]byte), m)
//		err := s.Response(m, mid)
//		if err != nil || m.GetRetCode() != 0 {
//			glog.SErrorf("subcribe ox failed:%v %v", err, m.GetRetCode())
//			return
//		}
//		ls.Add(s)
//	})
//	return nil
//}
//
//// C1020002 百人牛牛控制
//func (g *GateCore) C1020002(s *session.Session, msg *explr.C1020002, mid uint) error {
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1020002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1020002", msg, func(data interface{}) {
//		m := &explr.S1020002{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1020003 百人牛牛库存修改
//func (g *GateCore) C1020003(s *session.Session, msg *explr.C1020003, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1020003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1020003", msg, func(data interface{}) {
//		m := &explr.S1020003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1020004 百人牛牛控制取消
//func (g *GateCore) C1020004(s *session.Session, msg *explr.C1020004, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1020004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1020004", msg, func(data interface{}) {
//		m := &explr.S1020004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1020005 百人牛牛概率修正
//func (g *GateCore) C1020005(s *session.Session, msg *explr.C1020005, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1020005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1020005", msg, func(data interface{}) {
//		m := &explr.S1020005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1020006 百人牛牛概率查询
//func (g *GateCore) C1020006(s *session.Session, msg *explr.C1020006, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeOx100, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1020006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1020006", msg, func(data interface{}) {
//		m := &explr.S1020006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1030001 开始订阅红黑通知消息
//func (g *GateCore) C1030001(s *session.Session, msg *explr.C1030001, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1030001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeRedBlack)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeRedBlack)
//		m := &explr.S1030001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1030001", msg, func(data interface{}) {
//		m := &explr.S1030001{}
//		proto.Unmarshal(data.([]byte), m)
//		err := s.Response(m, mid)
//		if err != nil || m.GetRetCode() != 0 {
//			glog.SErrorf("subcribe rb failed:%v %v", err, m.GetRetCode())
//			return
//		}
//		ls.Add(s)
//	})
//	return nil
//}
//
//// C1030002 红黑控制
//func (g *GateCore) C1030002(s *session.Session, msg *explr.C1030002, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeRedBlack, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1030002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1030002", msg, func(data interface{}) {
//		m := &explr.S1030002{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1030003 红黑库存修改
//func (g *GateCore) C1030003(s *session.Session, msg *explr.C1030003, mid uint) error {
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeRedBlack, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1030003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1030003", msg, func(data interface{}) {
//		m := &explr.S1030003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1030004 红黑控制取消
//func (g *GateCore) C1030004(s *session.Session, msg *explr.C1030004, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeRedBlack, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1030004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1030004", msg, func(data interface{}) {
//		m := &explr.S1030004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1040001 开始订阅龙虎通知消息
//func (g *GateCore) C1040001(s *session.Session, msg *explr.C1040001, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1040001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeDT)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeDT)
//		m := &explr.S1040001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1040001", msg, func(data interface{}) {
//		m := &explr.S1040001{}
//		proto.Unmarshal(data.([]byte), m)
//		err := s.Response(m, mid)
//		if err != nil || m.GetRetCode() != 0 {
//			glog.SErrorf("subcribe dt failed:%v %v", err, m.GetRetCode())
//			return
//		}
//		ls.Add(s)
//	})
//	return nil
//}
//
//// C1040002 龙虎控制
//func (g *GateCore) C1040002(s *session.Session, msg *explr.C1040002, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDT, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1040002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1040002", msg, func(data interface{}) {
//		m := &explr.S1040002{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1040003 龙虎库存修改
//func (g *GateCore) C1040003(s *session.Session, msg *explr.C1040003, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDT, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1040003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1040003", msg, func(data interface{}) {
//		m := &explr.S1040003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1040004 取消
//func (g *GateCore) C1040004(s *session.Session, msg *explr.C1040004, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDT, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1040004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1040004", msg, func(data interface{}) {
//		m := &explr.S1040004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1060001 开始订阅奔驰宝马通知消息
//func (g *GateCore) C1060001(s *session.Session, msg *explr.C1060001, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1060001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeBenz)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBenz)
//		m := &explr.S1060001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1060001", msg, func(data interface{}) {
//		m := &explr.S1060001{}
//		proto.Unmarshal(data.([]byte), m)
//		err := s.Response(m, mid)
//		if err != nil || m.GetRetCode() != 0 {
//			glog.SErrorf("subcribe benz failed:%v %v", err, m.GetRetCode())
//			return
//		}
//		ls.Add(s)
//	})
//	return nil
//}
//
//// C1060002 奔驰宝马控制
//func (g *GateCore) C1060002(s *session.Session, msg *explr.C1060002, mid uint) error {
//
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1060002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1060002", msg, func(data interface{}) {
//		m := &explr.S1060002{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1060003 奔驰宝马库存修改
//func (g *GateCore) C1060003(s *session.Session, msg *explr.C1060003, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1060003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1060003", msg, func(data interface{}) {
//		m := &explr.S1060003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1060004 奔驰宝马控制取消
//func (g *GateCore) C1060004(s *session.Session, msg *explr.C1060004, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBenz, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1060004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1060004", msg, func(data interface{}) {
//		m := &explr.S1060004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1080001 订阅捕鱼
//func (g *GateCore) C1080001(s *session.Session, msg *explr.C1080001, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1080001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeFish)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeFish)
//		m := &explr.S1080001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1080001", msg, func(data interface{}) {
//		m := &explr.S1080001{}
//		proto.Unmarshal(data.([]byte), m)
//		err := s.Response(m, mid)
//		if err != nil || m.GetRetCode() != 0 {
//			glog.SErrorf("subcribe dt failed:%v %v", err, m.GetRetCode())
//			return
//		}
//		ls.Add(s)
//	})
//	return nil
//}
//
//// C1080002 捕获概率查询
//func (g *GateCore) C1080002(s *session.Session, msg *explr.C1080002, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeFish, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1080002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1080002", msg, func(data interface{}) {
//		m := &explr.S1080002{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C1080003 修改库存
//func (g *GateCore) C1080003(s *session.Session, msg *explr.C1080003, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeFish, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1080003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1080003", msg, func(data interface{}) {
//		m := &explr.S1080003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C1080004 捕获概率修改
//func (g *GateCore) C1080004(s *session.Session, msg *explr.C1080004, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeFish, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1080004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1080004", msg, func(data interface{}) {
//		m := &explr.S1080004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C1110001 开始订阅百家乐通知消息
//func (g *GateCore) C1110001(s *session.Session, msg *explr.C1110001, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1110001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeBaccarat)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBaccarat)
//		m := &explr.S1110001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1110001", msg, func(data interface{}) {
//		m := &explr.S1110001{}
//		proto.Unmarshal(data.([]byte), m)
//		err := s.Response(m, mid)
//		if err != nil || m.GetRetCode() != 0 {
//			glog.SErrorf("subcribe dt failed:%v %v", err, m.GetRetCode())
//			return
//		}
//		ls.Add(s)
//	})
//	return nil
//}
//
//// C1110002 百家乐控制
//func (g *GateCore) C1110002(s *session.Session, msg *explr.C1110002, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBaccarat, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1110002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1110002", msg, func(data interface{}) {
//		m := &explr.S1110002{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1110003 百家乐库存修改
//func (g *GateCore) C1110003(s *session.Session, msg *explr.C1110003, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBaccarat, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1110003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1110003", msg, func(data interface{}) {
//		m := &explr.S1110003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1110004 取消
//func (g *GateCore) C1110004(s *session.Session, msg *explr.C1110004, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeBaccarat, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1110004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1110004", msg, func(data interface{}) {
//		m := &explr.S1110004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C1120001 订阅
//func (g *GateCore) C1120001(s *session.Session, msg *explr.C1120001, mid uint) error {
//	if msg.GetUID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUID())
//		m := &explr.S1120001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeDuofu)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeDuofu)
//		m := &explr.S1120001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1120001", msg, func(data interface{}) {
//		m := &explr.S1120001{}
//		proto.Unmarshal(data.([]byte), m)
//		err := s.Response(m, mid)
//		if err != nil || m.GetRetCode() != 0 {
//			glog.SErrorf("subcribe duofu failed:%v %v", err, m.GetRetCode())
//			return
//		}
//		ls.Add(s)
//	})
//	return nil
//}
//
//// C1120003 库存修改
//func (g *GateCore) C1120003(s *session.Session, msg *explr.C1120003, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByUID(peer.TypeDuofu, s.UID())
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &explr.S1120003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C1120003", msg, func(data interface{}) {
//		m := &explr.S1120003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//
//	return nil
//}
//
//// C5000001 获取活动列表
//func (g *GateCore) C5000001(s *session.Session, msg *plr.C5000001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeActivity)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S5000001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C5000001", msg, func(data interface{}) {
//		m := &plr.S5000001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C5010001 获取连赢列表
//func (g *GateCore) C5010001(s *session.Session, msg *plr.C5010001, mid uint) error {
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeActivity)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S5010001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C5010001", msg, func(data interface{}) {
//		m := &plr.S5010001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
//
//// C5030001 推广链接
//func (g *GateCore) C5030001(s *session.Session, msg *plr.C5030001, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%d msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S5030001{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeWagency)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S5030001{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C5030001", msg, func(data interface{}) {
//		m := &plr.S5030001{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C5030002 我的推广
//func (g *GateCore) C5030002(s *session.Session, msg *plr.C5030002, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S5030002{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeWagency)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S5030002{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C5030002", msg, func(data interface{}) {
//		m := &plr.S5030002{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C5030003 推广明细
//func (g *GateCore) C5030003(s *session.Session, msg *plr.C5030003, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S5030003{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeWagency)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S5030003{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C5030003", msg, func(data interface{}) {
//		m := &plr.S5030003{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C5030004 推广周榜
//func (g *GateCore) C5030004(s *session.Session, msg *plr.C5030004, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S5030004{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeWagency)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S5030004{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C5030004", msg, func(data interface{}) {
//		m := &plr.S5030004{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C5030005 领取返利金币
//func (g *GateCore) C5030005(s *session.Session, msg *plr.C5030005, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S5030005{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeWagency)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S5030005{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C5030005", msg, func(data interface{}) {
//		m := &plr.S5030005{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C5030006 领取返利金币记录
//func (g *GateCore) C5030006(s *session.Session, msg *plr.C5030006, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S5030006{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeWagency)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S5030006{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C5030006", msg, func(data interface{}) {
//		m := &plr.S5030006{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}

// C5030007 全民代理开关
//func (g *GateCore) C5030007(s *session.Session, msg *plr.C5030007, mid uint) error {
//	if msg.GetUserID() != s.String(peer.KeyUserID) {
//		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
//		m := &plr.S5030007{}
//		m.RetCode = invalidUserid
//		return s.Response(m, mid)
//	}
//	ls := peer.GLinkServers.FindLinkByType(peer.TypeWagency)
//	if ls == nil || ls.GetConnector() == nil {
//		glog.SErrorf("Invalid UserID.Not Register.%d", s.UID())
//		m := &plr.S5030007{}
//		m.RetCode = errFindLinkServer
//		return s.Response(m, mid)
//	}
//
//	ls.Request("C5030007", msg, func(data interface{}) {
//		m := &plr.S5030007{}
//		proto.Unmarshal(data.([]byte), m)
//		s.Response(m, mid)
//	})
//	return nil
//}
