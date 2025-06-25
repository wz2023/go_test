package db

import (
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"newstars/framework/config"
)

// gormMysqlByConfig 初始化Mysql数据库
func gormMysqlByConfig(dbType string, config *config.Mysql) *gorm.DB {
	m := config
	if m.Dbname == "" {
		return nil
	}
	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}
	gormConf := &gormConfig{
		DbType:    dbType,
		MysqlConf: config,
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), gormConf.Config(m.Prefix, m.Singular)); err != nil {
		return nil
	} else {
		db.InstanceSet("gorm:table_options", "ENGINE="+m.Engine)
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		return db
	}
}
