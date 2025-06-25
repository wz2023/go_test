package model

import (
	"database/sql"
	"math/rand"
	"newstars/framework/glog"
	"newstars/framework/util/decimal"
	"sync"
)

type InventoryConfig struct {
	GameKindID    int32
	ControlType   int32
	LowThreshold  float64
	HighThreshold float64
	Percent       float64
}

type InveManager struct {
	inves []*InventoryConfig
	sync.Mutex
}

var Inves InveManager = InveManager{inves: make([]*InventoryConfig, 0)}

func (p *InveManager) load(db *sql.DB) {
	inves, err := QueryInventoryConfig(db)
	if err != nil {
		glog.SInfof("query inventory config err %v", err)
		return
	}
	p.Lock()
	p.inves = inves
	p.Unlock()
}

func (p *InveManager) trigProtect(gameKindID int32, inventory decimal.Decimal) bool {
	p.Lock()
	defer p.Unlock()
	num := 0
	for _, v := range Inves.inves {
		if v.GameKindID != gameKindID || v.ControlType != 0 {
			continue
		}
		num++
		low := decimal.NewFromFloat(v.LowThreshold)
		high := decimal.NewFromFloat(v.HighThreshold)
		if low.GreaterThan(high) {
			return inventory.LessThan(high)
		}
		if inventory.GreaterThanOrEqual(low) && inventory.LessThan(high) {
			r := rand.Float64()
			if r <= v.Percent/100 {
				return true
			}
		}
	}
	if num == 0 {
		return inventory.LessThan(decimal.Zero)
	}
	return false
}

// TrigProtect TrigProtect
func TrigProtect(gameKindID int32, inventory decimal.Decimal) bool {
	return Inves.trigProtect(gameKindID, inventory)
}
