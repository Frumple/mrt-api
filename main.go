package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

const (
	DB_CONFIG_PATH = "config/db_config.yml"
	COMPANIES_PATH = "data/companies.yml"
	WORLDS_PATH    = "data/worlds.yml"
)

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type StaticData interface {
	Company | World
	GetID() string
}

func main() {
	db := initializeDatabase()
	defer db.Close()

	companies := loadCompanies()
	worlds := loadWorlds()

	engine := gin.Default()
	engine.Use(DatabaseMiddleware(db))
	engine.Use(StaticDataMiddleware(companies, worlds))

	v1 := engine.Group("/v1")
	{
		v1.GET("/warps", getWarps)
		v1.GET("/warps/:name", getWarpByName)

		v1.GET("/companies", getCompanies)
		v1.GET("/companies/:id", getCompanyById)

		v1.GET("/worlds", getWorlds)
		v1.GET("/worlds/:id", getWorldById)
	}

	engine.Run()
}

func initializeDatabase() *sql.DB {
	connectionString := buildConnectionString()

	db, err := sql.Open("mysql", connectionString)
	checkForErrors(err)

	// sql.Open() doesn't establish any connection yet
	// Use db.Ping() to ensure that the database is available and ready
	db.Ping()
	checkForErrors(err)

	return db
}

func buildConnectionString() string {
	db_config := DbConfig{}

	data, err := os.ReadFile(DB_CONFIG_PATH)
	checkForErrors(err)

	err = yaml.Unmarshal([]byte(data), &db_config)
	checkForErrors(err)

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", db_config.User, db_config.Password, db_config.Host, db_config.Port, db_config.Database)
}

func loadStaticData[V StaticData](yamlFilePath string) *orderedmap.OrderedMap[string, V] {
	vSlice := []V{}

	data, err := os.ReadFile(yamlFilePath)
	checkForErrors(err)

	err = yaml.Unmarshal([]byte(data), &vSlice)
	checkForErrors(err)

	return staticDataSliceToOrderedMap(vSlice)
}

func staticDataSliceToOrderedMap[V StaticData](vSlice []V) *orderedmap.OrderedMap[string, V] {
	vMap := orderedmap.New[string, V]()

	for _, value := range vSlice {
		vMap.Set(value.GetID(), value)
	}

	return vMap
}

func orderedMapToValues[K comparable, V any](vMap *orderedmap.OrderedMap[K, V]) []V {
	vSlice := []V{}
	for pair := vMap.Oldest(); pair != nil; pair = pair.Next() {
		vSlice = append(vSlice, pair.Value)
	}

	return vSlice
}

func checkForErrors(err error) {
	if err != nil {
		panic(err)
	}
}

func createErrorBody(message string, detail string) *orderedmap.OrderedMap[string, string] {
	body := orderedmap.New[string, string]()
	body.Set("message", message)
	body.Set("detail", detail)
	return body
}
