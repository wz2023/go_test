package peer

import (
	"github.com/golang/protobuf/proto"
	"newstars/Protocol/plr"
	"newstars/framework/core/session"
)

func (p *LinkServer) onP3080001(data interface{}) {
	m := &plr.P3080001{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080001", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080002(data interface{}) {
	m := &plr.P3080002{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080002", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				if s.Int32(KeySeatID) != m.GetSeatNo() {
					return true
				}
				return false
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080003(data interface{}) {
	m := &plr.P3080003{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080003", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
		hall := GLinkServers.FindLinkByType(TypeHall)
		if hall != nil {
			n07 := &plr.N0000007{}
			n07.UserID = m.GetUserID()
			hall.Notify("N0000007", n07)
		}

	}
}

func (p *LinkServer) onP3080004(data interface{}) {
	m := &plr.P3080004{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080004", m, func(s *session.Session) bool {

			if s.Int32(KeyTableID) == m.GetTableID() {
				if s.Int32(KeySeatID) == m.GetSeatNo() {
					return false
				}
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080005(data interface{}) {
	m := &plr.P3080005{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080005", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080006(data interface{}) {
	m := &plr.P3080006{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080006", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080007(data interface{}) {
	m := &plr.P3080007{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080007", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				if s.Int32(KeySeatID) == m.GetSeatNo() {
					return false
				}
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080008(data interface{}) {
	m := &plr.P3080008{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080008", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				if s.Int32(KeySeatID) == m.GetSeatNo() {
					return false
				}
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080009(data interface{}) {
	m := &plr.P3080009{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080009", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080010(data interface{}) {
	m := &plr.P3080010{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080010", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080011(data interface{}) {
	m := &plr.P3080011{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080011", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080013(data interface{}) {
	m := &plr.P3080013{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080013", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				if s.Int32(KeySeatID) == m.GetSeatNo() {
					return false
				}
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080014(data interface{}) {
	m := &plr.P3080014{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080014", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				if s.Int32(KeySeatID) == m.GetSeatNo() {
					return false
				}
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080015(data interface{}) {
	m := &plr.P3080015{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080015", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080016(data interface{}) {
	m := &plr.P3080016{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080016", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}

func (p *LinkServer) onP3080017(data interface{}) {
	m := &plr.P3080017{}
	err := proto.Unmarshal(data.([]byte), m)
	if err == nil {
		p.group.Multicast("P3080017", m, func(s *session.Session) bool {
			if s.Int32(KeyTableID) == m.GetTableID() {
				return true
			}
			return false
		})
	}
}
