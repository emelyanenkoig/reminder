package handlers

import (
	"emelyanenkoig/reminder/cache"
	"emelyanenkoig/reminder/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func CreateUser(uc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		err := c.ShouldBindJSON(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		uc.AddUser(user)

		c.JSON(http.StatusCreated, user)
	}
}

func GetUser(uc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		user, exist := uc.GetUser(uint(id))
		if exist != true {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		}
		c.JSON(http.StatusOK, user)
	}

}
