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
