package handlers

import (
	"emelyanenkoig/reminder/pkg/repository"
	"github.com/gin-gonic/gin"
	"log"
)

func Run(userRepo *repository.UserRepository, reminderRepo *repository.ReminderRepository) {

	router := gin.Default()

	userGroup := router.Group("/")
	{
		userGroup.GET("/:id", GetUser(userRepo))
		userGroup.POST("/", CreateUser(userRepo))

		reminderGroup := userGroup.Group("/:id/reminders")
		{
			reminderGroup.GET("/", GetRemindersByUser(reminderRepo))
			reminderGroup.POST("/", CreateReminder(reminderRepo))
			reminderGroup.GET("/:reminder_id", GetUserReminderById(reminderRepo))
			reminderGroup.PUT("/:reminder_id", UpdateReminder(reminderRepo))
			reminderGroup.DELETE("/:reminder_id", DeleteReminder(reminderRepo))
		}
	}
	err := router.Run("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
