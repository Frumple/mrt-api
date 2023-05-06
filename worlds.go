package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type World struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
}

func (world World) GetID() string {
	return world.ID
}

func loadWorlds() *orderedmap.OrderedMap[string, World] {
	return loadStaticData[World](WORLDS_PATH)
}

func getWorlds(context *gin.Context) {
	worlds := context.MustGet(CONTEXT_WORLDS).(*orderedmap.OrderedMap[string, World])
	context.JSON(http.StatusOK, orderedMapToValues(worlds))
}

func getWorldById(context *gin.Context) {
	worlds := context.MustGet(CONTEXT_WORLDS).(*orderedmap.OrderedMap[string, World])
	id := context.Param("id")

	world, exists := worlds.Get(id)
	if !exists {
		body := createErrorBody(
			"World not found",
			"",
		)
		context.JSON(http.StatusNotFound, body)
		return
	}

	context.JSON(http.StatusOK, world)
}
