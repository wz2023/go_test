package table

import (
	"database/sql"
	"fmt"
	"newstars/Server/fish/consts"
	"newstars/Server/fish/model"
	"newstars/framework/glog"
)

// FishTable modals
type FishTable struct {
	tid    int32
	rid    int32
	gameid int32
	counts int32
	seats  map[int32]*FishSeat
}

// FishSeat modals
type FishSeat struct {
	Sid    int32
	Uid    string
	Tid    int32
	Rid    int32
	Status int32
	Name   string
}

// GCheckIPArea 检查地址区域
var GCheckIPArea bool

// FishManager manage table
type FishManager struct {
	rid int32
	db  *sql.DB
	ts  map[int32]*FishTable
}

// NewFishManager table mgr
func NewFishManager(rid int32, db *sql.DB) *FishManager {
	return &FishManager{
		rid: rid,
		db:  db,
		ts:  make(map[int32]*FishTable),
	}
}

// NewFishTable create instance
func NewFishTable(tid, rid, gameid int32) *FishTable {
	seats := make(map[int32]*FishSeat)
	var i int32
	for i = 0; i < consts.FishSeatNumbers; i++ {
		seats[i] = NewFishSeat(i, tid, rid)
	}

	return &FishTable{
		tid:    tid,
		rid:    rid,
		gameid: gameid,
		seats:  seats,
	}
}

// NewFishSeat create seat
func NewFishSeat(sid, tid, rid int32) *FishSeat {
	return &FishSeat{
		Sid: sid,
		Tid: tid,
		Rid: rid,
	}
}

// Init init db tables
func (p *FishManager) Init() error {
	ts, err := model.QueryGameTables(p.rid, p.db)
	if err != nil {
		return err
	}
	for i := range ts {
		tid := ts[i].Gametableid
		rid := ts[i].Gameroomid
		gameid := ts[i].GameID
		p.ts[tid] = NewFishTable(tid, rid, gameid)
	}
	return nil
}

// TakeTable select an available table
func (p *FishManager) TakeTable(nick string, gameid int32) *FishTable {
	//找只有一个人的桌台
	for _, v := range p.ts {
		if v.gameid == gameid {
			if v.counts == 1 {

				return v
			}
		}
	}
	//找最多人的桌台
	var maxtable *FishTable
	var maxtablecount int32
	for _, v := range p.ts {
		if v.gameid == gameid {
			if v.counts < consts.FishSeatNumbers {
				if v.counts > maxtablecount {
					maxtablecount = v.counts
					maxtable = v
					if v.counts == consts.FishSeatNumbers-1 {
						return v
					}
				}
			}
		}

	}
	if maxtable != nil {
		return maxtable
	}

	//随机桌台
	for _, v := range p.ts {
		if v.gameid == gameid {
			if v.counts < consts.FishSeatNumbers {
				return v
			}
		}
	}
	return p.InsertTable(gameid)
}

// InsertTable insert a table
func (p *FishManager) InsertTable(gameid int32) *FishTable {
	m := &model.Gametable{
		Gameroomid:  p.rid,
		GameID:      gameid,
		Seatnumbers: consts.FishSeatNumbers,
	}
	err := model.InsertGameTable(m, p.db)
	if err != nil {
		glog.SErrorf("InsertGameTable failed.Error:%v", err)
		return nil
	}
	tid := m.Gametableid
	t := NewFishTable(tid, p.rid, gameid)
	p.ts[tid] = t
	return t
}

// GetTable query table by tid
func (p *FishManager) GetTable(tid int32) *FishTable {
	for _, v := range p.ts {
		if v.tid == tid {
			return v
		}
	}
	return nil
}

// ValidityCheck check area
func (p *FishTable) ValidityCheck(nick string) bool {
	if GCheckIPArea {
		for _, v := range p.seats {
			if v.Name == nick {
				return false
			}
		}
	}
	return true
}

// SitDown sit down
func (p *FishTable) SitDown(uid string, nick string) error {
	for _, v := range p.seats {
		if v.Status == consts.SeatStatusNone {
			v.Name = nick
			v.Status = consts.SeatStatusOk
			v.Uid = uid
			p.counts++
			return nil
		}
	}
	return fmt.Errorf("Not enough seat for sit down")
}

// SitUp sit up
func (p *FishTable) SitUp(uid string) error {
	for _, v := range p.seats {
		if v.Uid == uid {
			v.Uid = ""
			v.Status = consts.SeatStatusNone
			v.Name = ""
			p.counts--
			return nil
		}
	}
	return fmt.Errorf("Invalid uid for sit up")
}

// SitUpAll sit up all
func (p *FishTable) SitUpAll() {
	for _, v := range p.seats {
		v.Uid = ""
		v.Status = consts.SeatStatusNone
		v.Name = ""
	}
	p.counts = 0
}

// QuerySeat by uid
func (p *FishTable) QuerySeat(uid string) *FishSeat {
	for _, v := range p.seats {
		if v.Uid == uid {
			return v
		}
	}
	return nil
}

func (p *FishTable) GetSeats() map[int32]*FishSeat {
	return p.seats
}

func (p *FishTable) Gameid() int32 {
	return p.gameid
}
