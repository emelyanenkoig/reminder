package handlers

import (
	"emelyanenkoig/reminder/pkg/repository"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func Run(userRepo repository.UserRepository, reminderRepo repository.ReminderRepository, host string, port int) {

	router := gin.Default()

	userGroup := router.Group("/")
	{
		userGroup.GET("/:id", GetUser(userRepo))
		userGroup.POST("/", CreateUser(userRepo))
		userGroup.PUT("/:id", UpdateUser(userRepo))
		userGroup.DELETE("/:id", DeleteUser(userRepo))
		userGroup.GET("/", GetUsers(userRepo))

		reminderGroup := userGroup.Group("/:id/reminders")
		{
			reminderGroup.GET("/", GetRemindersByUser(reminderRepo))
			reminderGroup.POST("/", CreateReminder(reminderRepo))
			reminderGroup.GET("/:reminder_id", GetUserReminderById(reminderRepo))
			reminderGroup.PUT("/:reminder_id", UpdateReminder(reminderRepo))
			reminderGroup.DELETE("/:reminder_id", DeleteReminder(reminderRepo))
		}
	}
	err := router.Run(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatal(err)
	}
}
