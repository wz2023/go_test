package conf

import (
	"encoding/json"
	"io/ioutil"
	"newstars/framework/glog"
	"time"
)

// Conf  配置
var Conf struct {
	Host   string
	Dsn    string
	Dbtype string
	// FishFreedom []FishFreedom
	FishTide     []FishTide
	TypeInterval map[int32]int64
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

func init() {
	data, err := os.ReadFile("conf/conf.json")
	if err != nil {
		glog.SFatalf("%v", err)
	}
	err = json.Unmarshal(data, &Conf)
	if err != nil {
		glog.SFatalf("%v", err)
	}
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
}
