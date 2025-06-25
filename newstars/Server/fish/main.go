package main

import (
	"database/sql"
	"flag"
	"newstars/Server/fish/conf"
	"newstars/Server/fish/core"
	"newstars/Server/fish/version"
	"newstars/framework/config"
	"newstars/framework/core/server"
	"newstars/framework/db"
	"newstars/framework/glog"
	"newstars/framework/model/data"
	"newstars/framework/redisx"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	flag.Parse()
	config.Viper()

	glog.Init(&config.GVA_CONFIG.Zap)

	db.InitGormDB("mysql", config.GVA_CONFIG)
	db.InitGormDBList(config.GVA_CONFIG.DBList)

	redisx.InitRedis(&config.GVA_CONFIG.Redis)

	conf.Init()
	data.InitBaseConfig()
}

func main() {
	glog.SInfo("NewFishServer main")
	// f, err := os.Create("cpuprofile")
	// if err != nil {
	// 	glog.SErrorf("creat profile failed.err:%v", err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	db, err := sql.Open(conf.Conf.Dbtype, conf.Conf.Dsn)
	if err != nil {
		glog.SFatalf("database config error:%v.program exit", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		glog.SFatalf("database config error:%v.program exit", err)
	}

	glog.SInfof("NewFishServer Go Go Go Version:%v", version.Version)
	db.SetConnMaxLifetime(10 * time.Minute)

	c := core.NewFishServer(db)
	server.Register(c)
	server.EnableDebug()
	server.Listen(conf.Conf.Host)
}
