package main

import (
	"emelyanenkoig/reminder/cache"
	"emelyanenkoig/reminder/handlers"
	"github.com/gin-gonic/gin"
	"log"
)

func Run() {
	uc := cache.NewCache()
	router := gin.Default()

	userGroup := router.Group("/")
	{
		userGroup.GET("/:id", handlers.GetUser(uc))
		userGroup.POST("/", handlers.CreateUser(uc))

		reminderGroup := userGroup.Group("/:id/reminders")
		{
			reminderGroup.GET("/", handlers.GetRemindersByUser(uc))
			reminderGroup.POST("/", handlers.CreateReminder(uc))
			reminderGroup.GET("/:reminder_id", handlers.GetReminderById(uc))
		}
	}
	err := router.Run("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
