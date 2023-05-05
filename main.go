package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	. "github.com/frumple/mrt-api/gen/mywarp_main/table"
	. "github.com/go-jet/jet/v2/mysql"

	"github.com/gin-gonic/gin"
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

type WarpResult struct {
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

type CompanyResult struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

type WorldResult struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
}

func main() {
	db := initializeDatabase()
	defer db.Close()

	companies := loadCompanies()
	worlds := loadWorlds()

	engine := gin.Default()
	engine.Use(Database(db))
	engine.Use(StaticData(companies, worlds))

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

// TODO: Refactor these two functions with generics
func loadCompanies() *orderedmap.OrderedMap[string, CompanyResult] {
	companySlice := []CompanyResult{}
	companyMap := orderedmap.New[string, CompanyResult]()

	data, err := os.ReadFile(COMPANIES_PATH)
	checkForErrors(err)

	err = yaml.Unmarshal([]byte(data), &companySlice)
	checkForErrors(err)

	for _, value := range companySlice {
		companyMap.Set(value.ID, value)
	}

	return companyMap
}

func loadWorlds() *orderedmap.OrderedMap[string, WorldResult] {
	worldSlice := []WorldResult{}
	worldMap := orderedmap.New[string, WorldResult]()

	data, err := os.ReadFile(WORLDS_PATH)
	checkForErrors(err)

	err = yaml.Unmarshal([]byte(data), &worldSlice)
	checkForErrors(err)

	for _, value := range worldSlice {
		worldMap.Set(value.ID, value)
	}

	return worldMap
}

func getWarps(context *gin.Context) {
	warps := []WarpResult{}

	db := context.MustGet(CONTEXT_DB).(*sql.DB)
	companies := context.MustGet(CONTEXT_COMPANIES).(*orderedmap.OrderedMap[string, CompanyResult])
	worlds := context.MustGet(CONTEXT_WORLDS).(*orderedmap.OrderedMap[string, WorldResult])

	companyID := context.Query("company")
	playerUUID := context.Query("player")
	worldID := context.Query("world")

	statement := SELECT(
		Warp.WarpID.AS("warpResult.id"),
		Warp.Name.AS("warpResult.name"),
		Player.UUID.AS("warpResult.playerUUID"),
		World.UUID.AS("warpResult.worldUUID"),
		Warp.X.AS("warpResult.x"),
		Warp.Y.AS("warpResult.y"),
		Warp.Z.AS("warpResult.z"),
		Warp.Pitch.AS("warpResult.pitch"),
		Warp.Yaw.AS("warpResult.yaw"),
		Warp.CreationDate.AS("warpResult.creationDate"),
		Warp.Type.AS("warpResult.type"),
		Warp.Visits.AS("warpResult.visits"),
		Warp.WelcomeMessage.AS("warpResult.welcomeMessage"),
	).FROM(
		Warp.
			INNER_JOIN(Player, Warp.PlayerID.EQ(Player.PlayerID)).
			INNER_JOIN(World, Warp.WorldID.EQ(World.WorldID)),
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

		boolExpressions = append(boolExpressions, Player.UUID.EQ(String(playerUUID)))
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

		boolExpressions = append(boolExpressions, Warp.Name.LIKE(String(company.Pattern)))
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

		boolExpressions = append(boolExpressions, World.UUID.EQ(String(worldUUID)))
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
	companies := context.MustGet(CONTEXT_COMPANIES).(*orderedmap.OrderedMap[string, CompanyResult])
	context.JSON(http.StatusOK, companies)
}

func getWorlds(context *gin.Context) {
	worlds := context.MustGet(CONTEXT_WORLDS).(*orderedmap.OrderedMap[string, WorldResult])
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
