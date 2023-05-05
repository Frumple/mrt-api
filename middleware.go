package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const (
	CONTEXT_DB        = "context_db"
	CONTEXT_COMPANIES = "context_companies"
	CONTEXT_WORLDS    = "worlds"
)

// Middleware to pass database into a handler function
func Database(db *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set(CONTEXT_DB, db)
		context.Next()
	}
}

// Middleware to pass YAML data into a handler function
func StaticData(companies *orderedmap.OrderedMap[string, CompanyResult], worlds *orderedmap.OrderedMap[string, WorldResult]) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set(CONTEXT_COMPANIES, companies)
		context.Set(CONTEXT_WORLDS, worlds)
		context.Next()
	}
}
