package utils

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func ListDatabase(dsn string) ([]string, error) {
	slashIndex := strings.LastIndex(dsn, "/")
	dsnPre := dsn[:slashIndex+1]

	dsnPre = fmt.Sprintf("%s?charset=utf8&parseTime=True&loc=Local", dsnPre)

	db, err := sql.Open("mysql", dsnPre)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SHOW DATABASES;")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var databases []string
	for rows.Next() {
		name := ""
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		databases = append(databases, name)
	}
	return databases, nil
}

func CreateDatabase(dsn string) error {
	slashIndex := strings.LastIndex(dsn, "/")
	dsnPre := dsn[:slashIndex+1]
	dbName := strings.Split(dsn[slashIndex+1:], "?")[0]

	db, err := sql.Open("mysql", dsnPre)
	if err != nil {
		return err
	}

	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4;",
		dbName,
	)

	_, err = db.Exec(createSQL)
	return err
}

func DropDatabase(dsn string) error {
	slashIndex := strings.LastIndex(dsn, "/")
	dsnPre := dsn[:slashIndex+1]
	dbName := strings.Split(dsn[slashIndex+1:], "?")[0]

	dsnPre = fmt.Sprintf("%s?charset=utf8&parseTime=True&loc=Local", dsnPre)
	db, err := sql.Open("mysql", dsnPre)
	if err != nil {
		return err
	}

	dropSQL := fmt.Sprintf(
		"DROP DATABASE `%s`;",
		dbName,
	)

	_, err = db.Exec(dropSQL)
	return err
}
