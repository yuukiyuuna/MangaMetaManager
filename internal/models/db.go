package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Optimize SQLite for concurrency
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.Exec("PRAGMA journal_mode=WAL;")
		sqlDB.Exec("PRAGMA busy_timeout=5000;")
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(10)
	}

	// Auto migrate models
	err = DB.AutoMigrate(&ProxySettings{}, &ProviderProxyStrategy{}, &MangaSeries{}, &MangaBook{}, &LibraryFolder{}, &AppSettings{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}
