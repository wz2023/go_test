package db

import (
	"gorm.io/gorm"
	"newstars/framework/config"
	"sync"
)

var (
	GormDB     *gorm.DB
	gormDBList map[string]*gorm.DB
	lock       sync.RWMutex
)

func InitGormDB(dbType string, config *config.Config) {
	switch dbType {
	case "mysql":
		GormDB = gormMysqlByConfig(dbType, &config.Mysql)
	case "pgsql":
		GormDB = gormPgSqlByConfig(dbType, &config.Pgsql)
	default:
		GormDB = gormMysqlByConfig(dbType, &config.Mysql)
	}
}

func InitGormDBList(dbList []config.SpecializedDB) {
	dbMap := make(map[string]*gorm.DB)
	for _, info := range dbList {
		if info.Disable {
			continue
		}
		switch info.Type {
		case "mysql":
			dbMap[info.AliasName] = gormMysqlByConfig(info.Type, &config.Mysql{GeneralDB: info.GeneralDB})
		case "pgsql":
			dbMap[info.AliasName] = gormPgSqlByConfig(info.Type, &config.Pgsql{GeneralDB: info.GeneralDB})
		default:
			continue
		}
	}
	if sysDB, ok := dbMap["system"]; ok {
		GormDB = sysDB
	}
	gormDBList = dbMap
}

func GetDBByName(dbname string) *gorm.DB {
	lock.RLock()
	defer lock.RUnlock()
	db, ok := gormDBList[dbname]
	if !ok || db == nil {
		panic("db no init")
	}
	return db
}
