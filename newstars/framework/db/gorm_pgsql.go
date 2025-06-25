package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"newstars/framework/config"
)

// gormPgSqlByConfig 初始化 Postgresql 数据库 通过参数
func gormPgSqlByConfig(dbType string, conf *config.Pgsql) *gorm.DB {
	if conf.Dbname == "" {
		return nil
	}
	pgsqlConfig := postgres.Config{
		DSN:                  conf.Dsn(), // DSN data source name
		PreferSimpleProtocol: false,
	}
	gormConf := &gormConfig{
		DbType:    dbType,
		PgsqlConf: conf,
	}
	if db, err := gorm.Open(postgres.New(pgsqlConfig), gormConf.Config(conf.Prefix, conf.Singular)); err != nil {
		panic(err)
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
		sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
		return db
	}
}
