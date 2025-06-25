package core

import (
	"github.com/golang/protobuf/proto"
	"newstars/Protocol/plr"
	"newstars/Server/gate/peer"
	"newstars/framework/core/session"
	"newstars/framework/glog"
)

// C3000001 请求房间列表
func (g *GateCore) C3000001(s *session.Session, msg *plr.S3000001, mid uint) error {
	glog.SInfof("[C3000001] 请求房间列表: %v", msg.String())

	ls := peer.GLinkServers.FindLinkByType(peer.TypeBlackjack)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBlackjack)
		m := &plr.S3000001{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	return ls.Request("C3000001", msg, func(data interface{}) {
		m := &plr.S3000001{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
}

// C3000002 请求匹配
func (g *GateCore) C3000002(s *session.Session, msg *plr.S3000002, mid uint) error {
	glog.SInfof("[C3000002] 请求匹配: %v", msg.String())

	ls := peer.GLinkServers.FindLinkByType(peer.TypeBlackjack)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBlackjack)
		m := &plr.S3000002{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	return ls.Request("C3000002", msg, func(data interface{}) {
		m := &plr.S3000002{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
}

// C3010002 玩家操作请求
func (g *GateCore) C3010002(s *session.Session, msg *plr.C3010002, mid uint) error {
	glog.SInfof("[C3010002] 玩家操作请求: %v", msg.String())
	if msg.GetUserID() != s.String(peer.KeyUserID) {
		glog.SErrorf("Invalid UserID.session:%s msg:%v", s.String(peer.KeyUserID), msg.GetUserID())
		m := &plr.S3010002{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	ls := peer.GLinkServers.FindLinkByType(peer.TypeBlackjack)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBlackjack)
		m := &plr.S3010002{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	return ls.Request("C3010002", msg, func(data interface{}) {
		m := &plr.S3010002{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
}

// C3010003 玩家重进
func (g *GateCore) C3010003(s *session.Session, msg *plr.C3010003, mid uint) error {
	glog.SInfof("[C3010003] 玩家重进: %v", msg.String())
	_, ok := g.uids.Load(s.UID())
	if !ok {
		m := &plr.S3010003{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}
	//linkID := v.(int)
	//ls := peer.GLinkServers.FindLinkByID(linkID) //待处理
	ls := peer.GLinkServers.FindLinkByType(peer.TypeBlackjack)
	if ls == nil || ls.GetConnector() == nil {
		glog.SErrorf("FindLinkByType Failed.%v", peer.TypeBlackjack)
		m := &plr.S3010003{}
		m.RetCode = invalidUserid
		return s.Response(m, mid)
	}

	return ls.Request("C3010003", msg, func(data interface{}) {
		m := &plr.S3010003{}
		proto.Unmarshal(data.([]byte), m)
		s.Response(m, mid)
	})
}
