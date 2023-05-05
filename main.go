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
	"gopkg.in/yaml.v3"
)

const (
	DB_CONFIG_PATH           = "config/db_config.yml"
	WARP_RAIL_COMPANIES_PATH = "data/warp_rail_companies.yml"
)

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type WarpRailCompanyResult struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
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

func main() {
	db := initializeDatabase()
	defer db.Close()

	engine := gin.Default()
	engine.Use(Database(db))

	v1 := engine.Group("/v1")
	{
		v1.GET("/warp_rail_companies", getWarpRailCompanies)
		v1.GET("/warps", getWarps)
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
	var db_config DbConfig

	data, err := os.ReadFile(DB_CONFIG_PATH)
	checkForErrors(err)

	err = yaml.Unmarshal([]byte(data), &db_config)
	checkForErrors(err)

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", db_config.User, db_config.Password, db_config.Host, db_config.Port, db_config.Database)
}

func getWarpRailCompanies(context *gin.Context) {
	var companies []WarpRailCompanyResult

	data, err := os.ReadFile(WARP_RAIL_COMPANIES_PATH)
	checkForErrors(err)

	err = yaml.Unmarshal([]byte(data), &companies)
	checkForErrors(err)

	context.JSON(http.StatusOK, companies)
}

func getWarps(context *gin.Context) {
	var warps []WarpResult

	db := context.MustGet(CONTEXT_DB).(*sql.DB)
	playerUUID := context.Query("playeruuid")

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

	if playerUUID != "" {
		// Add dashes to the UUID if they are missing
		if len(playerUUID) == 32 {
			playerUUID = fmt.Sprintf("%s-%s-%s-%s-%s", playerUUID[0:8], playerUUID[8:12], playerUUID[12:16], playerUUID[16:20], playerUUID[20:32])
		}

		if !isValidUUID(playerUUID) {
			context.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid UUID",
			})
			return
		}

		statement = statement.WHERE(Player.UUID.EQ(String(playerUUID)))
	}

	err := statement.Query(db, &warps)
	checkForErrors(err)

	context.JSON(http.StatusOK, warps)
}

func checkForErrors(err error) {
	if err != nil {
		panic(err)
	}
}
