package peer

import (
	"github.com/golang/protobuf/proto"
	"newstars/Protocol/plr"
	"newstars/framework/core/session"
)

//func (p *LinkServer) onP1000001(data interface{}) {
//	m := &plr.P1000001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Broadcast("P1000001", m)
//	}
//}
//
//func (p *LinkServer) onP1000002(data interface{}) {
//	m := &plr.P1000002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Broadcast("P1000002", m)
//	}
//}
//
//func (p *LinkServer) onP1000004(data interface{}) {
//	m := &plr.P1000004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P1000004", m, func(s *session.Session) bool {
//			if s.Int32(KeyGameID) == m.GetGameID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP1000005(data interface{}) {
//	m := &plr.P1000005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P1000005", m, func(s *session.Session) bool {
//			if s.Int32(KeyGameID) == m.GetGameID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP1000006(data interface{}) {
//	m := &plr.P1000006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P1000006", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP1000007(data interface{}) {
//	m := &plr.P1000007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P1000007", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}

func (p *LinkServer) onP1000008(data interface{}) {
	m := &plr.P1000008{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P1000008", m, func(s *session.Session) bool {
			if s.String(KeyUserID) == m.GetUserID() {
				return true
			}
			return false
		})
	}
}

//func (p *LinkServer) onP0000002(data interface{}) {
//	m := &plr.P0000002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P0000002", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}

func (p *LinkServer) onP1000009(data interface{}) {
	m := &plr.P1000009{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		if m.GetStatus() == 1 {
			p.group.Multicast("P1000009", m, func(s *session.Session) bool {
				if len(s.String(KeyUserID)) > 0 {
					return true
				}
				return false
			})
		}
	}
}

func (p *LinkServer) onP1000010(data interface{}) {
	m := &plr.P1000010{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P1000010", m, func(s *session.Session) bool {
			if len(s.String(KeyUserID)) > 0 {
				return true
			}
			return false
		})
	}
}

//func (p *LinkServer) onP0000003(data interface{}) {
//	m := &plr.P0000003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P0000003", m, func(s *session.Session) bool {
//			if len(s.String(KeyUserID)) > 0 {
//				return true
//			}
//			return false
//		})
//	}
//}
