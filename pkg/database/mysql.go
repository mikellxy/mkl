package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/mikellxy/mkl/config"
)

func init() {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return defaultTableName[:len(defaultTableName)-1]
	}
}

func GetDB() (*gorm.DB, error) {
	conf := config.Conf.MySQL
	connArg := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.User, conf.Password, conf.Host, conf.Port, conf.Database)
	return gorm.Open("mysql", connArg)
}
