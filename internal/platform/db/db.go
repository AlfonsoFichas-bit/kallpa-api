package db

import (
	"log/slog"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Init(dbName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return nil, err
	}

	return db, nil
}
