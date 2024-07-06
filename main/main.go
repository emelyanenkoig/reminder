package main

import (
	"emelyanenkoig/reminder/pkg/cache"
	"emelyanenkoig/reminder/pkg/database"
	"emelyanenkoig/reminder/pkg/handlers"
	"emelyanenkoig/reminder/pkg/repository"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=reminder port=5432 sslmode=disable"
	db := database.InitDB(dsn)
	c := cache.NewCache()

	userRepo := repository.NewUserRepository(db, c)
	reminderRepo := repository.NewReminderRepository(db, c)

	handlers.Run(userRepo, reminderRepo)
}
