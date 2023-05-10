package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/frumple/mrt-api/gen/mywarp_main/table"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	//lint:ignore ST1001 This dot import is intended for Jet SQL statements (SELECT, FROM, etc.)
	. "github.com/go-jet/jet/v2/mysql"
)

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

func (warp Warp) Render(writer http.ResponseWriter, request *http.Request) error {
	return nil
}

type WarpProvider struct {
	db              *sql.DB
	companyProvider CompanyProvider
	worldProvider   WorldProvider
}

const MAX_LIMIT = 2000

// getWarps godoc
// @summary     List all warps
// @description List all warps. Maximum number of warps returned per request is 2000. Use the 'offset' query parameter to show further entries.
// @tags        Warps
// @produce     json
// @param       name     query    string false "Filter by warp name."
// @param       player   query    string false "Filter by player UUID (can be with or without hyphens)."
// @param       company  query    string false "Filter by company ID (from /companies)."
// @param       world    query    string false "Filter by world ID (from /worlds)."
// @param       order_by query    string false "Order by 'name', 'creation_date', or 'visits'."
// @param       sort_by  query    string false "Sort by 'asc' (ascending) or 'desc' (descending)."
// @param       limit    query    int    false "Limit number of warps returned. Maximum limit is 2000."
// @param       offset   query    int    false "Number of warps to skip before returning."
// @success     200      {array}  Warp
// @failure     400      {object} Error
// @router      /warps [get]
func (provider WarpProvider) getWarps(writer http.ResponseWriter, request *http.Request) {
	warps := []Warp{}

	db := provider.db
	companies := provider.companyProvider.companies
	worlds := provider.worldProvider.worlds

	name := request.URL.Query().Get("name")
	playerUUID := request.URL.Query().Get("player")
	companyID := request.URL.Query().Get("company")
	worldID := request.URL.Query().Get("world")

	orderBy := request.URL.Query().Get("order_by")
	sortBy := request.URL.Query().Get("sort_by")

	limitStr := request.URL.Query().Get("limit")
	offsetStr := request.URL.Query().Get("offset")

	statement := beginWarpSelectStatement()

	boolExpressions := []BoolExpression{}

	// Filter by name
	if name != "" {
		boolExpressions = append(boolExpressions, table.Warp.Name.EQ(String(name)))
	}

	// Filter by player
	if playerUUID != "" {
		// Add hyphens to the UUID if they are missing
		if len(playerUUID) == 32 {
			playerUUID = fmt.Sprintf("%s-%s-%s-%s-%s", playerUUID[0:8], playerUUID[8:12], playerUUID[12:16], playerUUID[16:20], playerUUID[20:32])
		}

		if !isValidUUID(playerUUID) {
			detail := "The 'player' query parameter must be a UUID that has 32 hexadecimal digits (with or without hyphens)."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		boolExpressions = append(boolExpressions, table.Player.UUID.EQ(String(playerUUID)))
	}

	// Filter by company
	if companyID != "" {
		company, exists := companies.Get(companyID)

		if !exists {
			detail := "The 'company' query parameter must be one of the IDs returned from the /companies endpoint."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		boolExpressions = append(boolExpressions, table.Warp.Name.LIKE(String(company.Pattern)))
	}

	// Filter by world
	if worldID != "" {
		world, exists := worlds.Get(worldID)

		if !exists {
			detail := "The 'world' query parameter must be one of the IDs returned from the /worlds endpoint."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		worldUUID := world.UUID

		boolExpressions = append(boolExpressions, table.World.UUID.EQ(String(worldUUID)))
	}

	// Combine all filters
	if len(boolExpressions) > 0 {
		combinedBoolExpression := boolExpressions[0]
		for i := 1; i < len(boolExpressions); i++ {
			combinedBoolExpression = combinedBoolExpression.AND(boolExpressions[i])
		}

		statement.WHERE(combinedBoolExpression)
	}

	var column Column
	var orderByClause OrderByClause

	// Order by name, creation date, or visits
	if orderBy != "" {
		switch orderBy {
		case "name":
			column = table.Warp.Name
		case "creation_date":
			column = table.Warp.CreationDate
		case "visits":
			column = table.Warp.Visits
		default:
			detail := "The 'order_by' query parameter must be one of 'name', 'creation_date', or 'visits'."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}
	} else {
		column = table.Warp.WarpID
	}

	// Sort by ascending or descending
	if sortBy != "" {
		switch sortBy {
		case "asc":
			orderByClause = column.ASC()
		case "desc":
			orderByClause = column.DESC()
		default:
			detail := "The 'sort_by' query parameter must be one of 'asc' or 'desc'."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}
	} else {
		orderByClause = column.ASC()
	}

	statement.ORDER_BY(orderByClause)

	// Limit to a number of records
	limit := MAX_LIMIT

	// Use a different limit if specified
	if limitStr != "" {
		new_limit, err := strconv.Atoi(limitStr)
		if err != nil || new_limit < 0 || new_limit > MAX_LIMIT {
			detail := fmt.Sprintf("The 'limit' query parameter must be an unsigned integer within the following range: 0 <= limit <= %d.", MAX_LIMIT)
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		limit = new_limit
	}

	statement.LIMIT(int64(limit))

	// Offset number of records
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			detail := "The 'offset' query parameter must be an unsigned integer."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		statement.OFFSET(int64(offset))
	}

	err := statement.Query(db, &warps)
	checkForErrors(err)

	err = render.RenderList(writer, request, toRenderList(warps))
	if err != nil {
		render.Render(writer, request, ErrorRender(err))
		return
	}
}

// getWarpById  godoc
// @summary     Get warp by ID
// @description Get warp by ID.
// @tags        Warps
// @produce     json
// @param       id  path     int   true "Warp ID"
// @success     200 {object} Warp
// @failure     400 {object} Error
// @failure     404 {object} Error
// @router      /warps/{id} [get]
func (provider WarpProvider) getWarpById(writer http.ResponseWriter, request *http.Request) {
	warps := []Warp{}

	idStr := chi.URLParam(request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		detail := "The 'id' parameter must be an unsigned integer."
		render.Render(writer, request, ErrorBadRequest(detail))
		return
	}

	statement := beginWarpSelectStatement()

	statement.WHERE(table.Warp.WarpID.EQ(Int(int64(id))))

	err = statement.Query(provider.db, &warps)
	checkForErrors(err)

	if len(warps) == 0 {
		render.Render(writer, request, ErrorNotFound)
		return
	}

	err = render.Render(writer, request, warps[0])
	if err != nil {
		render.Render(writer, request, ErrorRender(err))
		return
	}
}

func beginWarpSelectStatement() SelectStatement {
	return SELECT(
		table.Warp.WarpID.AS("warp.ID"),
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

func warpsRouter(provider WarpProvider) http.Handler {
	router := chi.NewRouter()
	router.Get("/", provider.getWarps)

	router.Route("/{id}", func(subrouter chi.Router) {
		subrouter.Get("/", provider.getWarpById)
	})
	return router
}
