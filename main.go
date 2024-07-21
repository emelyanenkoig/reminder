package main

import (
	"emelyanenkoig/reminder/pkg/bot"
	"emelyanenkoig/reminder/pkg/cache"
	"emelyanenkoig/reminder/pkg/config"
	"emelyanenkoig/reminder/pkg/database"
	"emelyanenkoig/reminder/pkg/handlers"
	"emelyanenkoig/reminder/pkg/repository"
	"emelyanenkoig/reminder/pkg/services"
	"fmt"
	"log"
)

func main() {
	//cfg, err := config.LoadConfig()
	//if err != nil {
	//	log.Fatalf("Failed to load config: %v", err)
	//}
	cfg := config.LoadLocalConfig()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName,
		cfg.Database.Port, cfg.Database.SSLMode)
	db := database.InitDB(dsn)
	c := cache.NewCache()

	userRepo := repository.NewUserRepository(db, c)
	reminderRepo := repository.NewReminderRepository(db, c)

	userService := services.NewUserService(userRepo)
	reminderService := services.NewReminderService(reminderRepo)

	myBot, err := bot.NewBot(userService, reminderService)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	go myBot.Start()

	handlers.Run(userRepo, reminderRepo, cfg.Server.Host, cfg.Server.Port)
}
