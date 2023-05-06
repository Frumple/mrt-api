package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Company struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

func (company Company) GetID() string {
	return company.ID
}

func loadCompanies() *orderedmap.OrderedMap[string, Company] {
	return loadStaticData[Company](COMPANIES_PATH)
}

func getCompanies(context *gin.Context) {
	companies := context.MustGet(CONTEXT_COMPANIES).(*orderedmap.OrderedMap[string, Company])
	context.JSON(http.StatusOK, orderedMapToValues(companies))
}

func getCompanyById(context *gin.Context) {
	companies := context.MustGet(CONTEXT_COMPANIES).(*orderedmap.OrderedMap[string, Company])
	id := context.Param("id")

	company, exists := companies.Get(id)
	if !exists {
		body := createErrorBody(
			"Company not found",
			"",
		)
		context.JSON(http.StatusNotFound, body)
		return
	}

	context.JSON(http.StatusOK, company)
}
