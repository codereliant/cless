package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DBFileName = "cless.sqlite3"

func NewSqliteDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(DBFileName), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
