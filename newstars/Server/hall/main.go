package main

import (
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"newstars/Server/hall/conf"
	"newstars/Server/hall/core"
	"newstars/Server/hall/version"
	"newstars/framework/config"
	"newstars/framework/core/server"
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

	conf.Init()
	data.InitBaseConfig()
}

func main() {
	glog.SInfof("Hallserver main")
	glog.SInfof("Hallserver Go Go Go Version:%v", version.Version)
	db, err := sql.Open(conf.Conf.Dbtype, conf.Conf.Dsn)
	if err != nil {
		glog.SFatalf("database config error:%v.program exit", err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		glog.SFatalf("database config error:%v.program exit", err)
	}

	db2, err := sql.Open(conf.Conf.Dbtype, conf.Conf.Dsn2)
	if err != nil {
		glog.SFatalf("database config error:%v.program exit", err)
	}
	defer db2.Close()
	err = db2.Ping()
	if err != nil {
		glog.SFatalf("database config error:%v.program exit", err)
	}

	db.SetConnMaxLifetime(10 * time.Minute)
	db2.SetConnMaxLifetime(10 * time.Minute)

	c := core.NewHallCore(db, db2)
	server.Register(c)
	server.EnableDebug()
	server.SetCheckOriginFunc(func(_ *http.Request) bool { return true })
	server.Listen(conf.Conf.Host)
}
