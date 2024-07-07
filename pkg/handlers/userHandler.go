package handlers

import (
	"emelyanenkoig/reminder/pkg/models"
	"emelyanenkoig/reminder/pkg/repository"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func CreateUser(repo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		err := c.ShouldBindJSON(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = repo.CreateUser(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errFailedCreateUser})
			return
		}

		c.JSON(http.StatusCreated, user)
	}
}

func GetUser(repo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errInvalidUserID})
			return
		}

		user, err := repo.GetUserById(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": errUserNotFound})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func GetUsers(repo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := repo.GetUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errFailedFetchUsers})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func UpdateUser(repo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		err := c.ShouldBindJSON(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		idStr := c.Param("id")
		userId, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}

		// Load user from DB to ensure all fields are present and current
		var existingUser models.User
		err = repo.DB.Where("id = ?", userId).First(&existingUser).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		// Update fields
		existingUser.Username = user.Username
		existingUser.Reminders = user.Reminders

		err = repo.UpdateUser(uint(userId), &existingUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
			return
		}

		c.JSON(http.StatusOK, existingUser)
	}
}

func DeleteUser(repo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("id")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errInvalidUserID})
			return
		}

		err = repo.DeleteUser(uint(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errFailedDeleteUser})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}
