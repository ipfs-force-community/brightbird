package utils

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

func CreateDatabase(dsn string) error {
	slashIndex := strings.LastIndex(dsn, "/")
	dsnPre := dsn[:slashIndex+1]
	dbName := strings.Split(dsn[slashIndex+1:], "?")[0]

	dsnPre = fmt.Sprintf("%s?charset=utf8&parseTime=True&loc=Local", dsnPre)
	db, err := gorm.Open(mysql.Open(dsnPre), nil)
	if err != nil {
		return err
	}

	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4;",
		dbName,
	)

	return db.Exec(createSQL).Error
}

func DropDatabase(dsn string) error {
	slashIndex := strings.LastIndex(dsn, "/")
	dsnPre := dsn[:slashIndex+1]
	dbName := strings.Split(dsn[slashIndex+1:], "?")[0]

	dsnPre = fmt.Sprintf("%s?charset=utf8&parseTime=True&loc=Local", dsnPre)
	db, err := gorm.Open(mysql.Open(dsnPre), nil)
	if err != nil {
		return err
	}

	createSQL := fmt.Sprintf(
		"DROP DATABASE `%s`;",
		dbName,
	)

	return db.Exec(createSQL).Error
}
