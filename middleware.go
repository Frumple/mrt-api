package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const (
	CONTEXT_DB        = "db"
	CONTEXT_COMPANIES = "companies"
	CONTEXT_WORLDS    = "worlds"
)

func DatabaseMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set(CONTEXT_DB, db)
		context.Next()
	}
}

func StaticDataMiddleware(companies *orderedmap.OrderedMap[string, Company], worlds *orderedmap.OrderedMap[string, World]) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set(CONTEXT_COMPANIES, companies)
		context.Set(CONTEXT_WORLDS, worlds)
		context.Next()
	}
}
