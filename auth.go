package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func authRequired(c *gin.Context) {
	token := c.GetHeader("Authorization")

	if token == "" {
		c.AbortWithStatusJSON(401, gin.H{
			"error": "missing authorization token",
		})
		return
	}

	var userID uint64

	err := db.QueryRow(`
		SELECT id
		FROM users
		WHERE session_token = ?
	`, token).Scan(&userID)

	if err == sql.ErrNoRows {
		c.AbortWithStatusJSON(401, gin.H{
			"error": "invalid session",
		})
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Set("user_id", userID)

	c.Next()
}