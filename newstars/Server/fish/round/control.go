package round

import (
	"database/sql"
	"errors"
	"newstars/Server/fish/consts"
	"newstars/Server/fish/model"
	"newstars/framework/glog"
	"newstars/framework/util/decimal"
	"sync"
	"time"
)

// RevenueControl 库存
type RevenueControl struct {
	total        map[int32]decimal.Decimal // 库存
	reven        map[int32]decimal.Decimal // 营收
	captureRates map[int32]float64         // 捕获概率
	enterAmount  map[int32]float64         // 入场金额
	newTotal     map[int32]decimal.Decimal // 新手库存
	mu           sync.Mutex
	newmu        sync.Mutex //新手库存锁
	cmu          sync.Mutex
	uid          string
}

// NewRevenueControl create
func NewRevenueControl(db *sql.DB) *RevenueControl {
	renven := &RevenueControl{
		total:        make(map[int32]decimal.Decimal),
		reven:        make(map[int32]decimal.Decimal),
		captureRates: make(map[int32]float64),
		enterAmount:  make(map[int32]float64),
		newTotal:     make(map[int32]decimal.Decimal),
	}
	ins, err := model.QueryInventory(db)
	if err != nil {

		glog.SFatalf("landlordsRooms init failed.err:%v", err.Error())
	}
	insmap := make(map[int32]*model.Inventory)
	for _, v := range ins {
		if v.KindID == consts.FishKindID {
			insmap[v.RoomID] = v
		}
	}
	rooms, err := model.QueryRoomListByKind(consts.FishKindID, db)
	if err != nil {
		glog.SFatalf("QueryRoomListByKind failed kindid:%v", consts.FishKindID)
	}
	for _, v := range rooms {
		if inv, ok := insmap[v.Gameroomid]; ok {
			renven.total[v.Gameroomid] = decimal.NewFromFloat(inv.PoolAmount)
			// renven.enterAmount[v.Gameroomid], _ = decimal.New(int64(v.Minenteramount), 0).Float64()
			if v.Minenteramount == 0 {
				renven.enterAmount[v.Gameroomid], _ = decimal.NewFromFloat(consts.RoomMinEnterAmount).Float64()
			} else {
				renven.enterAmount[v.Gameroomid], _ = decimal.New(int64(v.Minenteramount), 0).Float64()
			}
			renven.reven[v.Gameroomid] = decimal.NewFromFloat(inv.Revenue)
		} else {
			item := &model.Inventory{}
			item.KindID = consts.FishKindID
			item.PoolAmount = 0
			item.Revenue = 0
			item.RoomID = v.Gameroomid
			item.TableID = 0
			item.UpdateTime = time.Now().Unix()
			renven.total[v.Gameroomid] = decimal.Zero
			if v.Minenteramount == 0 {
				renven.enterAmount[v.Gameroomid], _ = decimal.NewFromFloat(consts.RoomMinEnterAmount).Float64()
			} else {
				renven.enterAmount[v.Gameroomid], _ = decimal.New(int64(v.Minenteramount), 0).Float64()
			}
			err = model.InsertInventory(item, db)
		}
		exists, err := model.CheckRoomConfigExists(db, v.Gameroomid)
		if err == nil && !exists {
			model.InitRoomConfig(db, v.Gameroomid, consts.CaptureRate)
		}

		var newtotal float64
		renven.captureRates[v.Gameroomid], newtotal, err = model.QueryRoomCaptureRate(db, v.Gameroomid)
		renven.newTotal[v.Gameroomid] = decimal.NewFromFloat(newtotal)
		if err != nil {
			renven.captureRates[v.Gameroomid] = consts.CaptureRate
		}
	}
	return renven
}

// SetAmount 库存
func (p *RevenueControl) SetAmount(rid int32, amount float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.total[rid]; ok {
		p.total[rid] = decimal.NewFromFloat(amount)
	}
}

// SetCaptureRate 捕获概率
func (p *RevenueControl) SetCaptureRate(roomid int32, rate float64) {
	p.cmu.Lock()
	defer p.cmu.Unlock()
	p.captureRates[roomid] = rate
}

// GetRoomCaptureRate 捕获概率
func (p *RevenueControl) GetRoomCaptureRate(roomid int32) float64 {
	p.cmu.Lock()
	defer p.cmu.Unlock()
	return p.captureRates[roomid]
}

// getExtraCaptureRate 获取额外库存概率
func (p *RevenueControl) getExtraCaptureRate(roomid int32) float64 {
	total := p.GetToTal(roomid)
	if total < 50+p.enterAmount[roomid]*40 {
		return -0.1
	} else if total < 100+p.enterAmount[roomid]*50 {
		return -0.05
	} else if total > 200+p.enterAmount[roomid]*200 {
		return 0.1
	} else if total > 100+p.enterAmount[roomid]*100 {
		return 0.05
	}
	return 0.0
}

func (p *RevenueControl) GetCaptureRates() map[int32]float64 {
	p.cmu.Lock()
	defer p.cmu.Unlock()
	return p.captureRates
}

// UpdateAmount 更新
func (p *RevenueControl) UpdateAmount(rid int32, payoff decimal.Decimal, reven decimal.Decimal) (float64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.total[rid]; ok {
		p.total[rid] = p.total[rid].Add(payoff)
		p.reven[rid] = p.reven[rid].Add(reven)
		total, _ := p.total[rid].Float64()
		return total, nil
	}
	return 0, errors.New("rid is invalid")
}

// GetToTal 获取总量
func (p *RevenueControl) GetToTal(rid int32) float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	var total float64
	if _, ok := p.total[rid]; ok {
		total, _ = p.total[rid].Float64()
		return total
	}
	return total
}

// SetNewAmount 新手库存
func (p *RevenueControl) SetNewAmount(rid int32, amount float64) {
	p.newmu.Lock()
	defer p.newmu.Unlock()
	if _, ok := p.total[rid]; ok {
		p.newTotal[rid] = decimal.NewFromFloat(amount)
	}
}

// UpdateNewAmount 更新
func (p *RevenueControl) UpdateNewAmount(rid int32, payoff decimal.Decimal, reven decimal.Decimal) (float64, error) {
	p.newmu.Lock()
	defer p.newmu.Unlock()
	if _, ok := p.newTotal[rid]; ok {
		p.newTotal[rid] = p.newTotal[rid].Add(payoff)
		// p.reven[rid] = p.reven[rid].Add(reven)
		total, _ := p.newTotal[rid].Float64()
		return total, nil
	}
	return 0, errors.New("rid is invalid")
}

// GetNewToTal 获取总量
func (p *RevenueControl) GetNewToTal(rid int32) float64 {
	p.newmu.Lock()
	defer p.newmu.Unlock()
	var total float64
	if _, ok := p.newTotal[rid]; ok {
		total, _ = p.newTotal[rid].Float64()
		return total
	}
	return total
}

// GetRevenue reve
func (p *RevenueControl) GetRevenue(rid int32) float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	var reven float64
	if _, ok := p.reven[rid]; ok {
		reven, _ = p.reven[rid].Float64()
		return reven
	}
	return reven
}

// SetUID set uid
func (p *RevenueControl) SetUID(uid string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.uid = uid
}

// UID uid
func (p *RevenueControl) UID() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.uid
}
