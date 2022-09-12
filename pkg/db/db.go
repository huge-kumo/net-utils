package db

import (
	"database/sql"
	"github.com/huge-kumo/net-utils/config"
	"github.com/huge-kumo/net-utils/pkg/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

var orm *gorm.DB

func init() {
	var err error
	var db *sql.DB

	defer func() {
		if err != nil {
			log.GetInstance().Error("[系统错误] 数据库模版初始化失败 " + err.Error())
			os.Exit(0)
		}
	}()

	if orm, err = gorm.Open(sqlite.Open(config.Conf.Storage.FilePath), &gorm.Config{}); err != nil {
		return
	}
	if db, err = orm.DB(); err != nil {
		return
	}
	if err = db.Ping(); err != nil {
		return
	}

	return

}

func GetInstance() *gorm.DB {
	return orm
}
