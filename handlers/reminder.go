package handlers

import (
	"emelyanenkoig/reminder/cache"
	"emelyanenkoig/reminder/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func CreateReminder(uc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var reminder models.Reminder
		err = c.ShouldBindJSON(&reminder)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, exist := uc.GetUser(uint(userID))
		if !exist {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
			return
		}

		reminder.ID = uint(len(user.Reminders) + 1)
		reminder.CreatedAt = time.Now()
		reminder.UpdatedAt = time.Now()

		for i := 0; i < len(user.Reminders); i++ {
			if reminder.ID == user.Reminders[i].ID {
				c.JSON(http.StatusNotFound, gin.H{"error": "Reminder already exists"})
				return
			}
		}
		user.Reminders = append(user.Reminders, reminder)
		uc.AddUser(user)
		c.JSON(http.StatusCreated, reminder)
	}
}

func GetRemindersByUser(uc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		user, exist := uc.GetUser(uint(userID))
		if exist != true {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		}
		c.JSON(http.StatusOK, user.Reminders)
	}
}

func GetReminderById(uc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdStr := c.Param("id")
		userId, err := strconv.ParseUint(userIdStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		user, exist := uc.GetUser(uint(userId))
		if exist != true {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		}

		reminderIdStr := c.Param("reminder_id")
		reminderId, err := strconv.ParseUint(reminderIdStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder ID"})
			return
		}

		for _, reminder := range user.Reminders {
			if reminder.ID == uint(reminderId) {
				c.JSON(http.StatusOK, reminder)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No such reminder ID"})
			}
		}

	}
}
