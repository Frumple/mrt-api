package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

const CONTEXT_DB = "context_db"

// Middleware to pass database into a handler function
func Database(db *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set(CONTEXT_DB, db)
		context.Next()
	}
}
