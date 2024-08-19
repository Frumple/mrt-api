package main

import (
	"net/http"
	"time"

	"github.com/frumple/mrt-api/gen/mywarp_main/table"
	"github.com/go-chi/chi/v5"

	//lint:ignore ST1001 This dot import is intended for Jet SQL statements (SELECT, FROM, etc.)
	. "github.com/go-jet/jet/v2/mysql"
)

const MAX_WARPS_LIMIT = 2000

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

type WarpProvider interface {
	getWarps(writer http.ResponseWriter, request *http.Request)
	getWarpById(writer http.ResponseWriter, request *http.Request)
}

func warpsRouter(provider WarpProvider) http.Handler {
	router := chi.NewRouter()
	router.Get("/", provider.getWarps)

	router.Route("/{id}", func(subrouter chi.Router) {
		subrouter.Get("/", provider.getWarpById)
	})
	return router
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
