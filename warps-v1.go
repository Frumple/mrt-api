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

type WarpProviderV1 struct {
	db              *sql.DB
	companyProvider CompanyProvider
	worldProvider   WorldProvider
}

func (provider WarpProviderV1) getWarps(writer http.ResponseWriter, request *http.Request) {
	warps := []Warp{}

	db := provider.db
	companiesByID := provider.companyProvider.companiesByID
	worldsByID := provider.worldProvider.worldsByID

	name := request.URL.Query().Get("name")
	playerUUID := request.URL.Query().Get("player")
	companyID := request.URL.Query().Get("company")
	worldID := request.URL.Query().Get("world")
	typeStr := request.URL.Query().Get("type")

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
		company, exists := companiesByID.Get(companyID)

		if !exists {
			detail := "The 'company' query parameter must be one of the IDs returned from the /companies endpoint."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		boolExpressions = append(boolExpressions, table.Warp.Name.LIKE(String(company.Pattern)))
	}

	// Filter by world
	if worldID != "" {
		world, exists := worldsByID.Get(worldID)

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

func (provider WarpProviderV1) getWarpById(writer http.ResponseWriter, request *http.Request) {
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
