package main

import (
	"flag"
	"net/http"
	"newstars/Server/gate/core"
	"newstars/Server/gate/peer"
	"newstars/Server/gate/version"
	"newstars/framework/config"
	"newstars/framework/core/gate"
	"newstars/framework/db"
	"newstars/framework/glog"
	"newstars/framework/model/data"
	"newstars/framework/redisx"
	"time"
)

func init() {
	flag.Parse()
	config.Viper()

	glog.Init(&config.GVA_CONFIG.Zap)

	db.InitGormDB("mysql", config.GVA_CONFIG)
	db.InitGormDBList(config.GVA_CONFIG.DBList)

	redisx.InitRedis(&config.GVA_CONFIG.Redis)

	data.InitBaseConfig()
	peer.Init()
}

func main() {

	glog.SInfof("Gateserver Go Go Go Version:%v", version.Version)
	time.Sleep(2 * time.Second)
	c := core.NewGateCore()
	gate.Register(c)
	gate.EnableDebug()
	gate.SetCheckOriginFunc(func(_ *http.Request) bool { return true })
	gate.ListenWS(peer.Conf.Host)
}
