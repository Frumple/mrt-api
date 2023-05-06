package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/frumple/mrt-api/gen/mywarp_main/table"
	"github.com/gin-gonic/gin"

	//lint:ignore ST1001 This dot import is intended for Jet SQL statements (SELECT, FROM, etc.)
	. "github.com/go-jet/jet/v2/mysql"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Warp struct {
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

func getWarps(context *gin.Context) {
	warps := []Warp{}

	db := context.MustGet(CONTEXT_DB).(*sql.DB)
	companies := context.MustGet(CONTEXT_COMPANIES).(*orderedmap.OrderedMap[string, Company])
	worlds := context.MustGet(CONTEXT_WORLDS).(*orderedmap.OrderedMap[string, World])

	companyID := context.Query("company")
	playerUUID := context.Query("player")
	worldID := context.Query("world")

	statement := beginWarpSelectStatement()

	boolExpressions := []BoolExpression{}

	// Filter by player
	if playerUUID != "" {
		// Add hyphens to the UUID if they are missing
		if len(playerUUID) == 32 {
			playerUUID = fmt.Sprintf("%s-%s-%s-%s-%s", playerUUID[0:8], playerUUID[8:12], playerUUID[12:16], playerUUID[16:20], playerUUID[20:32])
		}

		if !isValidUUID(playerUUID) {
			body := createErrorBody(
				"Invalid player UUID format",
				"The 'player' query parameter must be a UUID that has 32 hexadecimal digits (with or without hyphens).",
			)
			context.JSON(http.StatusBadRequest, body)
			return
		}

		boolExpressions = append(boolExpressions, table.Player.UUID.EQ(String(playerUUID)))
	}

	// Filter by company
	if companyID != "" {
		company, exists := companies.Get(companyID)

		if !exists {
			body := createErrorBody(
				"Invalid company",
				"The 'company' query parameter must be one of the IDs returned from the /companies endpoint.",
			)
			context.JSON(http.StatusBadRequest, body)
			return
		}

		boolExpressions = append(boolExpressions, table.Warp.Name.LIKE(String(company.Pattern)))
	}

	// Filter by world
	if worldID != "" {
		world, exists := worlds.Get(worldID)

		if !exists {
			body := createErrorBody(
				"Invalid world",
				"The 'world' query parameter must be one of the IDs returned from the /worlds endpoint.",
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

func getWarpByName(context *gin.Context) {
	warps := []Warp{}

	db := context.MustGet(CONTEXT_DB).(*sql.DB)
	name := context.Param("name")

	statement := beginWarpSelectStatement()
	statement.WHERE(table.Warp.Name.EQ(String(name)))

	err := statement.Query(db, &warps)
	checkForErrors(err)

	if len(warps) == 0 {
		body := createErrorBody(
			"Warp not found",
			"",
		)
		context.JSON(http.StatusNotFound, body)
		return
	}

	context.JSON(http.StatusOK, warps[0])
}

func beginWarpSelectStatement() SelectStatement {
	return SELECT(
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
}
