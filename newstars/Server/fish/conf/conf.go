package conf

import (
	"encoding/json"
	"newstars/framework/glog"
	"newstars/framework/util/decimal"
	"os"
	"time"
)

// Conf  配置
var Conf struct {
	Host   string
	Dsn    string
	Dbtype string
	// FishFreedom []FishFreedom
	FishTide        []FishTide
	TypeInterval    map[int32]int64
	RevenueRate     float64
	DecRevenueRate  decimal.Decimal
	ExtraCommission map[int]float64
}

// type FishFreedom struct {
// 	KindID string
// 	Paths  string
// 	IsBoss bool
// }

type TideData struct {
	TimeAxis int32
	KindIDs  string
	Paths    string
}

type FishTide struct {
	TideData []TideData
	TimeUnit time.Duration
	Delay    int64
}

func Init() {
	data, err := os.ReadFile("conf/conf.json")
	if err != nil {
		glog.SFatalf("%v", err)
	}
	err = json.Unmarshal(data, &Conf)
	if err != nil {
		glog.SFatalf("%v", err)
	}

	if Conf.RevenueRate < 0 {
		Conf.RevenueRate = 0
	}
	Conf.DecRevenueRate = decimal.NewFromFloat(Conf.RevenueRate)
}

func Refresh() {
	data, err := os.ReadFile("conf/conf.json")
	if err != nil {
		glog.SFatalf("%v", err)
	}
	err = json.Unmarshal(data, &Conf)
	if err != nil {
		glog.SFatalf("%v", err)
	}
	if Conf.RevenueRate < 0 {
		Conf.RevenueRate = 0
	}
	Conf.DecRevenueRate = decimal.NewFromFloat(Conf.RevenueRate)
}
