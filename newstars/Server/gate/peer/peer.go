package peer

import (
	"encoding/json"
	"errors"
	"fmt"
	"newstars/Protocol/plr"
	"newstars/framework/core/gate"
	"newstars/framework/core/session"
	"newstars/framework/glog"
	"os"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

func Init() {
	data, err := os.ReadFile("conf/server.json")
	if err != nil {
		glog.SFatalf("%v", err)
	}
	err = json.Unmarshal(data, &Conf)
	if err != nil {
		glog.SFatalf("%v", err)
	}

	GGameNumber = Conf.GameNumber
	for _, v := range Conf.Servers {
		ls := NewLinkServer(v.ID, v.Status, v.Type, v.Addr, v.Reserved)
		GLinkServers = append(GLinkServers, ls)
	}
	if GLinkServers == nil {
		glog.SFatalf("Init LinkServers Failed")
	}
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				data, err := os.ReadFile("conf/server.json")
				if err != nil {
					glog.SFatalf("%v", err)
				}
				err = json.Unmarshal(data, &Conf)
				if err != nil {
					glog.SFatalf("%v", err)
				} else {
					//for _, v := range Conf.Servers {
					//	if v.Type == TypePay {
					//		pl := GLinkServers.FindLinkByID(v.ID)
					//		if pl != nil {
					//			if pl.reserved != v.Reserved {
					//				pl.reserved = v.Reserved
					//				glog.SInfof("set %v reserved:%d", pl.GetID(), v.Reserved)
					//			}
					//		}
					//	}
					//}
				}
			}
		}
	}()
}

// LinkServer 连接服务
type LinkServer struct {
	id       int
	iType    string
	addr     string
	status   int
	reserved int
	con      *gate.Connector
	group    *gate.Group
	wTime    time.Duration
	bConnect bool
	mu       sync.RWMutex
}

// NewLinkServer create instance
func NewLinkServer(id, status int, iType, addr string, reserved int) *LinkServer {
	return &LinkServer{
		id:       id,
		iType:    iType,
		addr:     addr,
		status:   status,
		con:      gate.NewConnector(),
		group:    gate.NewGroup(iType),
		wTime:    time.Second,
		reserved: reserved,
	}
}

// Add add session
func (p *LinkServer) Add(s *session.Session) error {
	return p.group.Add(s)
}

// GetSession by id
func (p *LinkServer) GetSession(id string) (*session.Session, error) {
	return p.group.Member(id)
}

// Leave session
func (p *LinkServer) Leave(s *session.Session) error {
	switch p.iType {
	case TypeHall:
		uid := s.String(KeyUserID)
		if len(uid) <= 0 {
			//m := &explr.N1010001{}
			//m.UID = uid
			//p.Notify("N1010001", m)
		} else {
			m := &plr.N0000005{}
			m.UserID = uid
			p.Notify("N0000005", m)
		}
	case TypeFish:
		uid := s.String(KeyUserID)
		leaveType := s.Int32(KeyLeaveType)
		if len(uid) > 0 {
			m := &plr.N3080001{}
			m.UserID = uid
			m.LeaveType = leaveType
			s.Remove(KeyLeaveType)
			p.Notify("N3080001", m)
		}
	}

	if p.iType != TypeHall {
		s.SetGameStatus(StatusHall)
	}

	return p.group.Leave(s)
}

// GroupLeave GroupLeave
func (p *LinkServer) GroupLeave(s *session.Session) error {
	return p.group.Leave(s)
}

// GetType 获取设备类型
func (p *LinkServer) GetType() string {
	return p.iType
}

// GetCounts 获取在线人数
func (p *LinkServer) GetCounts(gameid int32) int32 {
	return int32(p.group.CountBy(func(s *session.Session) bool {
		if s.Int32(KeyGameID) == gameid && len(s.String(KeyUserID)) > 0 {
			return true
		}
		return false
	}))
}

//// PeopleCounting 统计人数
//func (p *LinkServer) PeopleCounting(gameid int32) []*explr.P1010003_RoomCounts {
//	pCount := make([]*explr.P1010003_RoomCounts, 0)
//
//	mems := p.group.Members()
//	rCounts := make(map[int32]int32)
//	for _, v := range mems {
//		if ses, err := p.group.Member(v); err == nil {
//			if ses.Int32(KeyGameID) == gameid {
//				rid := ses.Int32(KeyRoomID)
//				if rid > 0 {
//					_, ok := rCounts[rid]
//					if ok {
//						rCounts[rid]++
//					} else {
//						rCounts[rid] = 1
//					}
//				}
//			}
//		}
//	}
//
//	for k, v := range rCounts {
//		item := &explr.P1010003_RoomCounts{}
//		item.RoomID = k
//		item.Counts = v
//		pCount = append(pCount, item)
//	}
//	return pCount
//}
//
//// PushCounts push counts
//func (p *LinkServer) PushCounts(push *explr.P1010003) {
//
//	p.group.Multicast("P1010003", push, func(s *session.Session) bool {
//		if len(s.String(KeyUserID)) < 0 {
//			return true
//		}
//		return false
//	})
//}

// Request 请求
func (p *LinkServer) Request(route string, v proto.Message, callback gate.Callback) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.bConnect {
		return p.con.Request(route, v, callback)
	}
	return errors.New("connector not connect")
}

// Notify 通知
func (p *LinkServer) Notify(route string, v proto.Message) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.bConnect {
		return p.con.Notify(route, v)
	}
	return errors.New("connector not connect")
}

// GetConnector 获取连接
func (p *LinkServer) GetConnector() *gate.Connector {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.bConnect {
		return nil
	}
	return p.con
}

// Contains uid check
func (p *LinkServer) Contains(uid string) bool {
	return p.group.Contains(uid)
}

// GetType hall landlord
func (p *LinkServer) String() string {
	return fmt.Sprintf("Type:%v Count:%v", p.iType, p.group.Count())
}

// OnDisConnected disconnected from linkserver
func (p *LinkServer) OnDisConnected() {
	glog.SErrorf("Link Server OnDisConnected %v Failed.", p.addr)
	p.mu.Lock()
	p.bConnect = false
	p.mu.Unlock()
	p.group.LeaveAll()

	gate.NewCountTimer(p.wTime, 1, func() {
		p.con = gate.NewConnector()
		err := p.Start()
		if err != nil {
			p.wTime = p.wTime * 2
			if p.wTime > 32*time.Second {
				p.wTime = 32 * time.Second
			}
		} else {
			p.wTime = time.Second
		}
	})
}

// Start start server
func (p *LinkServer) Start() error {
	glog.SInfof("Start connect link server:%v", p.addr)
	err := p.con.Start(p.addr)
	if err != nil {
		//glog.SErrorf("Link Server %v Failed.", p.addr)
		p.OnDisConnected()
		return errors.New("Link Server Except")
	}

	p.mu.Lock()
	p.bConnect = true
	p.mu.Unlock()

	glog.SInfof("Start Connect OK:%v", p.addr)
	p.con.OnDisConnected(p.OnDisConnected)

	switch p.iType {
	case TypeFish:
		p.con.On("P3080001", p.onP3080001)
		p.con.On("P3080002", p.onP3080002)
		p.con.On("P3080003", p.onP3080003)
		p.con.On("P3080004", p.onP3080004)
		p.con.On("P3080005", p.onP3080005)
		p.con.On("P3080006", p.onP3080006)
		p.con.On("P3080007", p.onP3080007)
		p.con.On("P3080008", p.onP3080008)
		p.con.On("P3080009", p.onP3080009)
		p.con.On("P3080010", p.onP3080010)
		p.con.On("P3080011", p.onP3080011)
		p.con.On("P3080013", p.onP3080013)
		p.con.On("P3080014", p.onP3080014)
		p.con.On("P3080015", p.onP3080015)
		p.con.On("P3080016", p.onP3080016)
		p.con.On("P3080017", p.onP3080017)
		//control
		p.con.On("P1010002", p.onP1010002)
	case TypeBlackjack:
		p.con.On("P3010001", p.onP3010001)
		p.con.On("P3010002", p.onP3010002)
		p.con.On("P3010003", p.onP3010003)
		p.con.On("P3010004", p.onP3010004)
	case TypeHall:
		//p.con.On("P1000001", p.onP1000001)
		//p.con.On("P1000002", p.onP1000002)
		//p.con.On("P1000004", p.onP1000004)
		//p.con.On("P1000005", p.onP1000005)
		//p.con.On("P1000006", p.onP1000006)
		//p.con.On("P1000007", p.onP1000007)
		p.con.On("P1000008", p.onP1000008)
		//p.con.On("P0000002", p.onP0000002)
		p.con.On("P1000009", p.onP1000009)
		p.con.On("P1000010", p.onP1000010)
	}
	//p.con.On("P0000001", p.onP0000001)
	//p.con.On("P0000003", p.onP0000003)
	return nil
}

// GetID id
func (p *LinkServer) GetID() int {
	return p.id
}

// LinkServers 集合
type LinkServers []*LinkServer

// GLinkServers globla
var GLinkServers LinkServers

type IGateCore interface {
	StoreUIDAndLinkSvrID(uid string, linkSvrID int)
}

var GGateCore IGateCore

// FindLinkByType type
func (ls LinkServers) FindLinkByType(iType string) *LinkServer {
	for _, v := range ls {
		if v.group.Count() < maxSessions && v.iType == iType && v.reserved == 0 {
			return v
		}
	}
	return nil
}

// FindLinkByID server id
func (ls LinkServers) FindLinkByID(id int) *LinkServer {
	for _, v := range ls {
		if v.group.Count() < maxSessions && v.id == id {
			return v
		}
	}
	return nil
}

// FindLinkByUID uid
func (ls LinkServers) FindLinkByUID(iType string, uid string) *LinkServer {
	for _, v := range ls {
		if v.iType == iType && v.Contains(uid) {
			return v
		}
	}
	return nil
}

// FindLinksByUID uid
func (ls LinkServers) FindLinksByUID(uid string) LinkServers {
	ret := make(LinkServers, 0)
	for _, v := range ls {
		if v.Contains(uid) {
			ret = append(ret, v)
		}
	}
	return ret
}

// Go run ls
func (ls LinkServers) Go() error {
	var err error
	for _, v := range ls {
		if v.status == 1 {
			err = v.Start()
			if err != nil {
				glog.SErrorf("Start link server %v failed %v.", v.iType, err)
			}
		}
	}
	return err
}

func (p *LinkServer) onP1010001(data interface{}) {
	//m := &explr.P1010001{}
	//err := proto.Unmarshal(data.([]byte), m)
	//if err == nil {
	//	p.group.Multicast("P1010001", m, func(s *session.Session) bool {
	//		if len(s.String(KeyUserID)) < 0 {
	//			return true
	//		}
	//		return false
	//	})
	//}
}

func (p *LinkServer) onP1010002(data interface{}) {
	//m := &explr.P1010002{}
	//err := proto.Unmarshal(data.([]byte), m)
	//if err == nil {
	//	p.group.Multicast("P1010002", m, func(s *session.Session) bool {
	//		if len(s.String(KeyUserID)) < 0 {
	//			return true
	//		}
	//		return false
	//	})
	//}
}

//
//// 推送用户准备成功消息
//func (p *LinkServer) onP3010001(data interface{}) {
//	m := &plr.P3010001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010001", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 服务端推送游戏开始消息
//func (p *LinkServer) onP3010002(data interface{}) {
//	m := &plr.P3010002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010002", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				s.Set(KeyRoundName, m.GetRoundName())
//				return true
//			}
//			return false
//		})
//
//		n3 := &plr.N0000003{}
//		n3.RoundName = m.GetRoundName()
//		n3.TableID = m.GetTableID()
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls != nil {
//			ls.Notify("N0000003", n3)
//		}
//	}
//}
//
//// 服务端发牌
//func (p *LinkServer) onP3010003(data interface{}) {
//	m := &plr.P3010003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010003", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() && s.Int32(KeySeatID) == m.GetSeatNo() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 开始叫分
//func (p *LinkServer) onP3010004(data interface{}) {
//	m := &plr.P3010004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010004", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 推送叫分结果
//func (p *LinkServer) onP3010005(data interface{}) {
//	m := &plr.P3010005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010005", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() && s.Int32(KeySeatID) != m.GetSeatNo() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 牌局庄家
//func (p *LinkServer) onP3010006(data interface{}) {
//	m := &plr.P3010006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010006", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 开始出牌
//func (p *LinkServer) onP3010007(data interface{}) {
//	m := &plr.P3010007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010007", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 牌局结束
//func (p *LinkServer) onP3010008(data interface{}) {
//	m := &plr.P3010008{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010008", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				s.Set(KeyRoomID, int32(-1))
//				s.Set(KeySeatID, int32(-1))
//				s.Set(KeyTableID, int32(-1))
//				s.Set(KeyRoundName, "")
//				return true
//			}
//			return false
//		})
//
//		n4 := &plr.N0000004{}
//		n4.RoundName = m.GetRoundName()
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls != nil {
//			ls.Notify("N0000004", n4)
//		}
//	}
//}
//
//// 推送用户出牌数据
//func (p *LinkServer) onP3010009(data interface{}) {
//	m := &plr.P3010009{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010009", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 通知农民开始加倍
//func (p *LinkServer) onP3010010(data interface{}) {
//	m := &plr.P3010010{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010010", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() &&
//				s.Int32(KeySeatID) == m.GetSeatNo() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 通知加倍结果
//func (p *LinkServer) onP3010011(data interface{}) {
//	m := &plr.P3010011{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3010011", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 推送用户退出消息
//func (p *LinkServer) onP301000A(data interface{}) {
//	m := &plr.P301000A{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P301000A", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 推送托管消息
//func (p *LinkServer) onP301000B(data interface{}) {
//	m := &plr.P301000B{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P301000B", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3020001(data interface{}) {
//	m := &plr.P3020001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3020001", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3020002(data interface{}) {
//	m := &plr.P3020002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3020002", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				s.Set(KeyRoundName, m.GetRoundName())
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3020003(data interface{}) {
//	m := &plr.P3020003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3020003", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3020004(data interface{}) {
//	m := &plr.P3020004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3020004", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3020005(data interface{}) {
//	m := &plr.P3020005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3020005", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3020006(data interface{}) {
//	m := &plr.P3020006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3020006", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3020007(data interface{}) {
//	m := &plr.P3020007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3020007", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3020008(data interface{}) {
//	m := &plr.P3020008{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3020008", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				if s.Int32(KeySeatID) == m.GetSeatNo() {
//					s.Set(KeyRoomID, int32(-1))
//					s.Set(KeySeatID, int32(-1))
//					s.Set(KeyTableID, int32(-1))
//					s.Set(KeyRoundName, "")
//				}
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3020009(data interface{}) {
//	m := &plr.P3020009{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3020009", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP302000B(data interface{}) {
//	m := &plr.P302000B{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P302000B", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP302000C(data interface{}) {
//	m := &plr.P302000C{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P302000C", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3030001(data interface{}) {
//	m := &plr.P3030001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3030001", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3030002(data interface{}) {
//	m := &plr.P3030002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3030002", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				s.Set(KeyRoundName, m.GetRoundName())
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3030003(data interface{}) {
//	m := &plr.P3030003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3030003", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3030004(data interface{}) {
//	m := &plr.P3030004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3030004", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3030005(data interface{}) {
//	m := &plr.P3030005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3030005", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3030006(data interface{}) {
//	m := &plr.P3030006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3030006", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3030007(data interface{}) {
//	m := &plr.P3030007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3030007", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3030008(data interface{}) {
//	m := &plr.P3030008{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3030008", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				if s.String(KeyUserID) == m.GetSeatNo() {
//					s.Set(KeySeatID, int32(-1))
//				}
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3030009(data interface{}) {
//	m := &plr.P3030009{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3030009", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP303000A(data interface{}) {
//	m := &plr.P303000A{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P303000A", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP303000B(data interface{}) {
//	m := &plr.P303000B{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P303000B", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP303000C(data interface{}) {
//	m := &plr.P303000C{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P303000C", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040001(data interface{}) {
//	m := &plr.P3040001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040001", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040002(data interface{}) {
//	m := &plr.P3040002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040002", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040003(data interface{}) {
//	m := &plr.P3040003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040003", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040004(data interface{}) {
//	m := &plr.P3040004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040004", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040005(data interface{}) {
//	m := &plr.P3040005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040005", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040006(data interface{}) {
//	m := &plr.P3040006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040006", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040007(data interface{}) {
//	m := &plr.P3040007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040007", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040008(data interface{}) {
//	m := &plr.P3040008{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040008", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040009(data interface{}) {
//	m := &plr.P3040009{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040009", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3040010(data interface{}) {
//	m := &plr.P3040010{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3040010", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050001(data interface{}) {
//	m := &plr.P3050001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050001", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050002(data interface{}) {
//	m := &plr.P3050002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050002", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				s.Set(KeyRoundName, m.GetRoundName())
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050003(data interface{}) {
//	m := &plr.P3050003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050003", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() && s.Int32(KeySeatID) == m.GetSeatNo() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050004(data interface{}) {
//	m := &plr.P3050004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050004", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050005(data interface{}) {
//	m := &plr.P3050005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050005", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050006(data interface{}) {
//	m := &plr.P3050006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050006", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050007(data interface{}) {
//	m := &plr.P3050007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050007", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//func (p *LinkServer) onP3050008(data interface{}) {
//	m := &plr.P3050008{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050008", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050009(data interface{}) {
//	m := &plr.P3050009{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050009", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() && s.Int32(KeySeatID) == m.GetSeatNo() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050010(data interface{}) {
//	m := &plr.P3050010{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050010", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				if m.GetRoundEndType() == -1 {
//					for _, v := range m.GetUserSettles() {
//						n := &plr.N0000007{}
//						n.UserID = v.GetUserID()
//						ls := GLinkServers.FindLinkByType(TypeHall)
//						if ls != nil {
//							ls.Notify("N0000007", n)
//						}
//					}
//				}
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050011(data interface{}) {
//	m := &plr.P3050011{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050011", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//
//		n := &plr.N0000007{}
//		n.UserID = m.GetUserID()
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls != nil {
//			ls.Notify("N0000007", n)
//		}
//	}
//}
//
//func (p *LinkServer) onP3050012(data interface{}) {
//	m := &plr.P3050012{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050012", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3050013(data interface{}) {
//	m := &plr.P3050013{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3050013", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060001(data interface{}) {
//	m := &plr.P3060001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060001", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060002(data interface{}) {
//	m := &plr.P3060002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060002", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060003(data interface{}) {
//	m := &plr.P3060003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060003", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060004(data interface{}) {
//	m := &plr.P3060004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060004", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060005(data interface{}) {
//	m := &plr.P3060005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060005", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060006(data interface{}) {
//	m := &plr.P3060006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060006", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060007(data interface{}) {
//	m := &plr.P3060007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060007", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060008(data interface{}) {
//	m := &plr.P3060008{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060008", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060009(data interface{}) {
//	m := &plr.P3060009{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060009", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3060010(data interface{}) {
//	m := &plr.P3060010{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3060010", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3070001(data interface{}) {
//	m := &plr.P3070001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3070001", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3070002(data interface{}) {
//	m := &plr.P3070002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3070002", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				s.Set(KeyRoundName, m.GetRoundName())
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3070003(data interface{}) {
//	m := &plr.P3070003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3070003", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3070004(data interface{}) {
//	m := &plr.P3070004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3070004", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3070005(data interface{}) {
//	m := &plr.P3070005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3070005", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3070006(data interface{}) {
//	m := &plr.P3070006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3070006", m, func(s *session.Session) bool {
//			if s.String(KeyRoundName) == m.GetRoundName() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3070007(data interface{}) {
//	m := &plr.P3070007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3070007", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3070008(data interface{}) {
//	m := &plr.P3070008{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3070008", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				if s.String(KeyUserID) == m.GetSeatNo() {
//					s.Set(KeySeatID, int32(-1))
//				}
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3070009(data interface{}) {
//	m := &plr.P3070009{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3070009", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP307000B(data interface{}) {
//	m := &plr.P307000B{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P307000B", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP307000C(data interface{}) {
//	m := &plr.P307000C{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P307000C", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 十三水
//// 推送用户退出消息
//func (p *LinkServer) onP309000A(data interface{}) {
//	m := &plr.P309000A{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P309000A", m, func(s *session.Session) bool {
//			return s.Int32(KeyTableID) == m.GetTableID()
//		})
//	}
//}
//
//// 入座
//func (p *LinkServer) onP3090001(data interface{}) {
//	m := &plr.P3090001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3090001", m, func(s *session.Session) bool {
//			return s.Int32(KeyTableID) == m.GetTableID()
//		})
//	}
//}
//
//// 准备
//func (p *LinkServer) onP3090002(data interface{}) {
//	m := &plr.P3090002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3090002", m, func(s *session.Session) bool {
//			return s.Int32(KeyTableID) == m.GetTableID()
//		})
//	}
//}
//
//// 开局
//func (p *LinkServer) onP3090003(data interface{}) {
//	m := &plr.P3090003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3090003", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				s.Set(KeyRoundName, m.GetRoundName())
//				return true
//			}
//			return false
//		})
//	}
//}
//
//// 发牌
//func (p *LinkServer) onP3090004(data interface{}) {
//	m := &plr.P3090004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3090004", m, func(s *session.Session) bool {
//			return s.Int32(KeyTableID) == m.GetTableID() && s.Int32(KeySeatID) == m.GetSeatNo()
//		})
//	}
//}
//
//// 理牌完成
//func (p *LinkServer) onP3090005(data interface{}) {
//	m := &plr.P3090005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3090005", m, func(s *session.Session) bool {
//			return s.Int32(KeyTableID) == m.GetTableID()
//		})
//	}
//}
//
//// 结算
//func (p *LinkServer) onP3090006(data interface{}) {
//	m := &plr.P3090006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3090006", m, func(s *session.Session) bool {
//			return s.Int32(KeyTableID) == m.GetTableID()
//		})
//	}
//}
//
//// 结束
//func (p *LinkServer) onP3090007(data interface{}) {
//	m := &plr.P3090007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3090007", m, func(s *session.Session) bool {
//			return s.Int32(KeyTableID) == m.GetTableID()
//		})
//	}
//}
//

//func (p *LinkServer) onP0000001(data interface{}) {
//	m := &plr.P0000001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P0000001", m, func(s *session.Session) bool {
//			if len(s.String(KeyUserID)) > 0 {
//				gids := m.GetGameIDs()
//				if gids == nil || len(gids) == 0 {
//					return true
//				}
//				for _, v := range gids {
//					if s.Int32(KeyGameID) == v {
//						return true
//					}
//				}
//				return false
//			}
//			return false
//		})
//	}
//}

//
//func (p *LinkServer) onP2010001(data interface{}) {
//	m := &plr.P2010001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P2010001", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP1010004(data interface{}) {
//	m := &explr.P1010004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P1010004", m, func(s *session.Session) bool {
//			if len(s.String(KeyUserID)) < 0 {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP1010005(data interface{}) {
//	m := &explr.P1010005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P1010005", m, func(s *session.Session) bool {
//			if len(s.String(KeyUserID)) < 0 {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) payP0000002(data interface{}) {
//	m := &plr.P0000002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P0000002", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//

//
//func (p *LinkServer) payP1000006(data interface{}) {
//	m := &plr.P1000006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P1000006", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100001(data interface{}) {
//	m := &plr.P3100001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100001", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100002(data interface{}) {
//	m := &plr.P3100002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100002", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100003(data interface{}) {
//	m := &plr.P3100003{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100003", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100004(data interface{}) {
//	m := &plr.P3100004{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100004", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100005(data interface{}) {
//	m := &plr.P3100005{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100005", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100006(data interface{}) {
//	m := &plr.P3100006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100006", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100007(data interface{}) {
//	m := &plr.P3100007{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100007", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100008(data interface{}) {
//	m := &plr.P3100008{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100008", m, func(s *session.Session) bool {
//			if s.Int32(KeyTableID) == m.GetTableID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100009(data interface{}) {
//	m := &plr.P3100009{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100009", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3100010(data interface{}) {
//	m := &plr.P3100010{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3100010", m, func(s *session.Session) bool {
//			if s.String(KeyUserID) == m.GetUserID() {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP3110001(data interface{}) {
//	m := &plr.P3110001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		p.group.Multicast("P3110001", m, func(s *session.Session) bool {
//			if (s.GameStatus() == StatusHall && len(s.String(KeyUserID)) > 0) || s.GameStatus() == StatusDuofu {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP1010006(data interface{}) {
//	m := &explr.P1010006{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P1010006", m, func(s *session.Session) bool {
//			if len(s.String(KeyUserID)) < 0 {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP5000001(data interface{}) {
//	m := &plr.P5000001{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P5000001", m, func(s *session.Session) bool {
//			if s.Int32(KeyGameID) == m.GameID {
//				return true
//			}
//			return false
//		})
//	}
//}
//
//func (p *LinkServer) onP5000002(data interface{}) {
//	m := &plr.P5000002{}
//	err := proto.Unmarshal(data.([]byte), m)
//	if err == nil {
//		ls := GLinkServers.FindLinkByType(TypeHall)
//		if ls == nil || ls.GetConnector() == nil {
//			return
//		}
//		ls.group.Multicast("P5000002", m, func(s *session.Session) bool {
//			if s.Int32(KeyGameID) == m.GameID {
//				return true
//			}
//			return false
//		})
//	}
//}
