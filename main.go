package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type WarpRailCompany struct {
	Id      string
	Name    string
	Pattern string
}

func main() {
	engine := gin.Default()

	v1 := engine.Group("/v1")
	{
		v1.GET("/warp_rail_companies", getWarpRailCompanies)
		v1.GET("/warps", getWarps)
	}

	engine.Run()
}

func checkForErrors(err error) {
	if err != nil {
		panic(err)
	}
}

func getWarpRailCompanies(context *gin.Context) {
	var companies []WarpRailCompany

	data, err := os.ReadFile("yaml/warp_rail_companies.yml")
	checkForErrors(err)

	err = yaml.Unmarshal([]byte(data), &companies)
	checkForErrors(err)

	context.JSON(http.StatusOK, companies)
}

func getWarps(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{})
}
