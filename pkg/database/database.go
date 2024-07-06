package database

import (
	"emelyanenkoig/reminder/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func InitDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Reminder{})
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	return db
}
