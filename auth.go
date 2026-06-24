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
	var role string

	err := db.QueryRow(`
		SELECT id, role
		FROM users
		WHERE session_token = ?
	`, token).Scan(&userID, &role)

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
	c.Set("role", role)

	c.Next()
}
