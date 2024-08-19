package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/frumple/mrt-api/gen/mywarp_main/table"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	//lint:ignore ST1001 This dot import is intended for Jet SQL statements (SELECT, FROM, etc.)
	. "github.com/go-jet/jet/v2/mysql"
)

type WarpResponse struct {
	Pagination WarpResponsePagination `json:"pagination"`
	Result     []Warp                 `json:"result"`
}

func (response WarpResponse) Render(writer http.ResponseWriter, request *http.Request) error {
	return nil
}

type WarpResponsePagination struct {
	Limit     int `json:"limit"`
	Offset    int `json:"offset"`
	Hits      int `json:"hits"`
	TotalHits int `json:"total_hits"`
}

func (pagination WarpResponsePagination) Render(writer http.ResponseWriter, request *http.Request) error {
	return nil
}

type WarpProviderV2 struct {
	db              *sql.DB
	companyProvider CompanyProvider
	worldProvider   WorldProvider
}

// getWarps godoc
// @summary     List all warps
// @description List all warps. Maximum number of warps returned per request is 2000. Use the 'offset' query parameter to show further entries.
// @tags        Warps
// @produce     json
// @param       name     query    string false "Filter by warp name."
// @param       player   query    string false "Filter by player UUID (can be with or without hyphens)."
// @param       company  query    string false "Filter by company ID (from /companies)."
// @param       world    query    string false "Filter by world ID (from /worlds)."
// @param       type     query    int    false "Filter by type (0 = private, 1 = public)."
// @param       order_by query    string false "Order by 'name', 'creation_date', or 'visits'."
// @param       sort_by  query    string false "Sort by 'asc' (ascending) or 'desc' (descending)."
// @param       limit    query    int    false "Limit number of warps returned. Maximum limit is 2000."
// @param       offset   query    int    false "Number of warps to skip before returning."
// @success     200      {object} WarpResponse
// @failure     400      {object} Error
// @router      /warps [get]
func (provider WarpProviderV2) getWarps(writer http.ResponseWriter, request *http.Request) {
	warps := []Warp{}

	db := provider.db
	companies := provider.companyProvider.companies
	worlds := provider.worldProvider.worlds

	name := request.URL.Query().Get("name")
	playerUUID := request.URL.Query().Get("player")
	companyID := request.URL.Query().Get("company")
	worldID := request.URL.Query().Get("world")
	typeStr := request.URL.Query().Get("type")

	orderBy := request.URL.Query().Get("order_by")
	sortBy := request.URL.Query().Get("sort_by")

	limitStr := request.URL.Query().Get("limit")
	offsetStr := request.URL.Query().Get("offset")

	selectStatement := beginWarpSelectStatement()
	countStatement := beginWarpCountStatement()

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

	// Filter by type
	if typeStr != "" {
		typeInt, err := strconv.Atoi(typeStr)
		if err != nil || typeInt < 0 || typeInt > 1 {
			detail := "The 'type' query parameter must be either 0 (private) or 1 (public)."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		boolExpressions = append(boolExpressions, table.Warp.Type.EQ(Int(int64(typeInt))))
	}

	// Combine all filters
	if len(boolExpressions) > 0 {
		combinedBoolExpression := boolExpressions[0]
		for i := 1; i < len(boolExpressions); i++ {
			combinedBoolExpression = combinedBoolExpression.AND(boolExpressions[i])
		}

		selectStatement.WHERE(combinedBoolExpression)
		countStatement.WHERE(combinedBoolExpression)
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

	selectStatement.ORDER_BY(orderByClause)

	// Limit to a number of records
	limit := MAX_WARPS_LIMIT

	// Use a different limit if specified
	if limitStr != "" {
		new_limit, err := strconv.Atoi(limitStr)
		if err != nil || new_limit < 0 || new_limit > MAX_WARPS_LIMIT {
			detail := fmt.Sprintf("The 'limit' query parameter must be an unsigned integer within the following range: 0 <= limit <= %d.", MAX_WARPS_LIMIT)
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		limit = new_limit
	}

	selectStatement.LIMIT(int64(limit))

	// Offset number of records
	offset := 0

	if offsetStr != "" {
		new_offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			detail := "The 'offset' query parameter must be an unsigned integer."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		offset = new_offset
	}

	selectStatement.OFFSET(int64(offset))

	err := selectStatement.Query(db, &warps)
	checkForErrors(err)

	countResult := CountResult{}

	err = countStatement.Query(db, &countResult)
	checkForErrors(err)

	hits := len(warps)
	total_hits := int(countResult.Count)

	pagination := WarpResponsePagination{limit, offset, hits, total_hits}
	response := WarpResponse{pagination, warps}

	err = render.Render(writer, request, response)
	// err = render.RenderList(writer, request, toRenderList(warps))
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
func (provider WarpProviderV2) getWarpById(writer http.ResponseWriter, request *http.Request) {
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

type CountResult struct {
	Count uint32 `json:"count"`
}

func beginWarpCountStatement() SelectStatement {
	return SELECT(
		COUNT(table.Warp.WarpID).AS("count_result.count"),
	).FROM(
		table.Warp.
			INNER_JOIN(table.Player, table.Warp.PlayerID.EQ(table.Player.PlayerID)).
			INNER_JOIN(table.World, table.Warp.WorldID.EQ(table.World.WorldID)),
	)
}
