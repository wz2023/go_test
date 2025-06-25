package conf

import (
	"database/sql"
	"encoding/json"
	"newstars/framework/glog"
	"os"
)

// Conf  配置
var Conf struct {
	Host              string
	Dsn               string
	Dsn2              string
	Dbtype            string
	SmsAcc            string
	SmsKey            string
	Products          map[string]float64
	GuestBaseWealth   float32
	BindMobileWealth  float32
	AccountBaseWealth float32
	BindSaleHost      string
	IPCheckInterval   int64
	IPUnlockInterval  int64
	IPLimitCount      int32
}

// ExchangeConf 兑换配置
var ExchangeConf struct {
	AliPayExchange    bool
	AlipayLimitAmount float64
	AlipayLimitCount  int32
	AilPayInterval    int32

	BankCardExchange    bool
	BankCardLimitCount  int32
	BankCardLimitAmount float64
	BankCardInterval    int32
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

	if Conf.IPLimitCount == 0 {
		Conf.IPLimitCount = 5
	}

	if Conf.IPCheckInterval == 0 {
		Conf.IPCheckInterval = 3600
	}

	if Conf.IPUnlockInterval == 0 {
		Conf.IPUnlockInterval = 86400
	}
}

type ConfigExchange struct {
	Type     int32
	Open     bool    // 开启
	Money    float64 // 限额
	Count    int32   // 限次
	Interval int32   // 间隔
}

func LoadExchangeConf(db *sql.DB) error {
	const (
		Limit1Type = iota
		BankCardType
		Limit2Type
	)

	var ualipay string
	err := db.QueryRow(`select exchange_alipay from global_config_t`).Scan(&ualipay)
	if err != nil {
		glog.SErrorf("LoadExchangeConf failed.err:%v", err)
		return err
	}
	var alipayconf = make([]*ConfigExchange, 3)
	err = json.Unmarshal([]byte(ualipay), &alipayconf)
	if err != nil {
		glog.SErrorf("%v", err)
		return err
	}

	ExchangeConf.BankCardLimitAmount = 0
	ExchangeConf.BankCardLimitCount = 0
	ExchangeConf.BankCardExchange = false
	ExchangeConf.BankCardInterval = 0
	ExchangeConf.AlipayLimitAmount = 0
	ExchangeConf.AlipayLimitCount = 0
	ExchangeConf.AliPayExchange = false
	ExchangeConf.AilPayInterval = 0

	for _, v := range alipayconf {
		if v.Type == BankCardType {
			ExchangeConf.BankCardLimitAmount = v.Money
			ExchangeConf.BankCardLimitCount = v.Count
			ExchangeConf.BankCardExchange = v.Open
			ExchangeConf.BankCardInterval = v.Interval
		} else if v.Type == Limit1Type || v.Type == Limit2Type {
			if v.Open {
				ExchangeConf.AlipayLimitAmount = v.Money
				ExchangeConf.AlipayLimitCount = v.Count
				ExchangeConf.AliPayExchange = v.Open
				ExchangeConf.AilPayInterval = v.Interval
			}
		}
	}
	return nil
}
