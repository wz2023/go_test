package main

import (
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"newstars/Server/blackjack/core"
	"newstars/Server/hall/version"
	"newstars/framework/config"
	"newstars/framework/core/server"
	"newstars/framework/db"
	"newstars/framework/glog"
	"newstars/framework/model/data"
	"newstars/framework/redisx"
)

func init() {
	flag.Parse()
	config.Viper()

	glog.Init(&config.GVA_CONFIG.Zap)

	db.InitGormDB("mysql", config.GVA_CONFIG)
	db.InitGormDBList(config.GVA_CONFIG.DBList)

	redisx.InitRedis(&config.GVA_CONFIG.Redis)

	data.InitBaseConfig()
}

func main() {
	glog.SInfof("blackjackServer main")
	glog.SInfof("blackjackServer Go Go Go Version:%v", version.Version)

	c := core.NewBlackJackServer()
	server.Register(c)
	server.EnableDebug()
	server.SetCheckOriginFunc(func(_ *http.Request) bool { return true })
	server.Listen(config.GVA_CONFIG.System.Host)
}
