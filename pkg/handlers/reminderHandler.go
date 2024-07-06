package handlers

import (
	"emelyanenkoig/reminder/pkg/models"
	"emelyanenkoig/reminder/pkg/repository"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func CreateReminder(repo *repository.ReminderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// parse user id
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

		reminder.UserID = uint(userID)
		reminder.CreatedAt = time.Now()
		reminder.UpdatedAt = time.Now()

		err = repo.CreateReminder(&reminder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reminder"})
			return
		}

		c.JSON(http.StatusCreated, reminder)
	}
}

func GetRemindersByUser(repo *repository.ReminderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		reminders, err := repo.GetRemindersByUser(uint(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reminders"})
			return
		}

		c.JSON(http.StatusOK, reminders)
	}
}

func GetUserReminderById(repo *repository.ReminderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		reminderIDStr := c.Param("reminder_id")
		reminderID, err := strconv.ParseUint(reminderIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder ID"})
			return
		}

		reminder, err := repo.GetReminderByUserId(uint(userID), uint(reminderID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
			return
		}

		c.JSON(http.StatusOK, reminder)
	}
}

func UpdateReminder(repo *repository.ReminderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		reminderIDStr := c.Param("reminder_id")
		reminderID, err := strconv.ParseUint(reminderIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder ID"})
			return
		}

		var updatedReminder models.Reminder
		if err := c.ShouldBindJSON(&updatedReminder); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = repo.UpdateReminder(uint(userID), uint(reminderID), &updatedReminder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reminder"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Reminder updated successfully"})
	}
}

func DeleteReminder(repo *repository.ReminderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		reminderIDStr := c.Param("reminder_id")
		reminderID, err := strconv.ParseUint(reminderIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder ID"})
			return
		}

		err = repo.DeleteReminder(uint(userID), uint(reminderID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reminder"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Reminder deleted successfully"})
	}
}
