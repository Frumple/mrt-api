package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/frumple/mrt-api/gen/mywarp_main/table"
	"github.com/gin-gonic/gin"

	//lint:ignore ST1001 This dot import is intended for Jet SQL statements (SELECT, FROM, etc.)
	. "github.com/go-jet/jet/v2/mysql"
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

type Warp struct {
	ID             uint32    `json:"id" sql:"primary_key"`
	Name           string    `json:"name"`
	PlayerUUID     string    `json:"playerUUID"`
	WorldUUID      string    `json:"worldUUID"`
	X              float64   `json:"x"`
	Y              float64   `json:"y"`
	Z              float64   `json:"z"`
	Pitch          float64   `json:"pitch"`
	Yaw            float64   `json:"yaw"`
	CreationDate   time.Time `json:"creationDate"`
	Type           uint8     `json:"type"`
	Visits         uint32    `json:"visits"`
	WelcomeMessage *string   `json:"welcomeMessage"`
}

type Company struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

type World struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
}

type StaticData interface {
	Company | World
	GetID() string
}

func (company Company) GetID() string {
	return company.ID
}

func (world World) GetID() string {
	return world.ID
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
		v1.GET("/companies", getCompanies)
		v1.GET("/worlds", getWorlds)
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

func loadCompanies() *orderedmap.OrderedMap[string, Company] {
	return loadStaticData[Company](COMPANIES_PATH)
}

func loadWorlds() *orderedmap.OrderedMap[string, World] {
	return loadStaticData[World](WORLDS_PATH)
}

func loadStaticData[V StaticData](yamlFilePath string) *orderedmap.OrderedMap[string, V] {
	vSlice := []V{}
	vMap := orderedmap.New[string, V]()

	data, err := os.ReadFile(yamlFilePath)
	checkForErrors(err)

	err = yaml.Unmarshal([]byte(data), &vSlice)
	checkForErrors(err)

	for _, value := range vSlice {
		vMap.Set(value.GetID(), value)
	}

	return vMap
}

func getWarps(context *gin.Context) {
	warps := []Warp{}

	db := context.MustGet(CONTEXT_DB).(*sql.DB)
	companies := context.MustGet(CONTEXT_COMPANIES).(*orderedmap.OrderedMap[string, Company])
	worlds := context.MustGet(CONTEXT_WORLDS).(*orderedmap.OrderedMap[string, World])

	companyID := context.Query("company")
	playerUUID := context.Query("player")
	worldID := context.Query("world")

	statement := SELECT(
		table.Warp.WarpID.AS("warp.id"),
		table.Warp.Name,
		table.Player.UUID.AS("warp.playerUUID"),
		table.World.UUID.AS("warp.worldUUID"),
		table.Warp.X,
		table.Warp.Y,
		table.Warp.Z,
		table.Warp.Pitch,
		table.Warp.Yaw,
		table.Warp.CreationDate,
		table.Warp.Type,
		table.Warp.Visits,
		table.Warp.WelcomeMessage,
	).FROM(
		table.Warp.
			INNER_JOIN(table.Player, table.Warp.PlayerID.EQ(table.Player.PlayerID)).
			INNER_JOIN(table.World, table.Warp.WorldID.EQ(table.World.WorldID)),
	)

	boolExpressions := []BoolExpression{}

	if playerUUID != "" {
		// Add hyphens to the UUID if they are missing
		if len(playerUUID) == 32 {
			playerUUID = fmt.Sprintf("%s-%s-%s-%s-%s", playerUUID[0:8], playerUUID[8:12], playerUUID[12:16], playerUUID[16:20], playerUUID[20:32])
		}

		if !isValidUUID(playerUUID) {
			body := createErrorBody(
				"Invalid UUID format",
				"The player must be a UUID that has 32 hexadecimal digits (with or without hyphens).",
			)
			context.JSON(http.StatusBadRequest, body)
			return
		}

		boolExpressions = append(boolExpressions, table.Player.UUID.EQ(String(playerUUID)))
	}

	if companyID != "" {
		company, exists := companies.Get(companyID)

		if !exists {
			body := createErrorBody(
				"Invalid company",
				"The company must be one of the entries returned from the /companies endpoint.",
			)
			context.JSON(http.StatusBadRequest, body)
			return
		}

		boolExpressions = append(boolExpressions, table.Warp.Name.LIKE(String(company.Pattern)))
	}

	if worldID != "" {
		world, exists := worlds.Get(worldID)

		if !exists {
			body := createErrorBody(
				"Invalid world",
				"The world must be one of the entries returned from the /worlds endpoint.",
			)
			context.JSON(http.StatusBadRequest, body)
			return
		}

		worldUUID := world.UUID

		boolExpressions = append(boolExpressions, table.World.UUID.EQ(String(worldUUID)))
	}

	if len(boolExpressions) > 0 {
		combinedBoolExpression := boolExpressions[0]
		for i := 1; i < len(boolExpressions); i++ {
			combinedBoolExpression = combinedBoolExpression.AND(boolExpressions[i])
		}

		statement.WHERE(combinedBoolExpression)
	}

	err := statement.Query(db, &warps)
	checkForErrors(err)

	context.JSON(http.StatusOK, warps)
}

func getCompanies(context *gin.Context) {
	companies := context.MustGet(CONTEXT_COMPANIES).(*orderedmap.OrderedMap[string, Company])
	context.JSON(http.StatusOK, companies)
}

func getWorlds(context *gin.Context) {
	worlds := context.MustGet(CONTEXT_WORLDS).(*orderedmap.OrderedMap[string, World])
	context.JSON(http.StatusOK, worlds)
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
