package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type World struct {
	ID   string `json:"id"`
	UUID string `json:"uuid"`
}

func (world World) GetID() string {
	return world.ID
}

func (world World) Render(writer http.ResponseWriter, request *http.Request) error {
	return nil
}

type WorldProvider struct {
	worlds *orderedmap.OrderedMap[string, World]
}

func (provider WorldProvider) getWorlds(writer http.ResponseWriter, request *http.Request) {
	err := render.RenderList(writer, request, toRenderList(orderedMapToValues(provider.worlds)))
	if err != nil {
		render.Render(writer, request, ErrorRender(err))
		return
	}
}

func (provider WorldProvider) getWorldById(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	company, exists := provider.worlds.Get(id)
	if !exists {
		render.Render(writer, request, ErrorNotFound)
		return
	}

	err := render.Render(writer, request, company)
	if err != nil {
		render.Render(writer, request, ErrorRender(err))
		return
	}
}

func worldsRouter(provider WorldProvider) http.Handler {
	router := chi.NewRouter()
	router.Get("/", provider.getWorlds)

	router.Route("/{id}", func(subrouter chi.Router) {
		subrouter.Get("/", provider.getWorldById)
	})
	return router
}

func loadWorlds() WorldProvider {
	worlds := loadStaticData[World](WORLDS_PATH)
	return WorldProvider{worlds: worlds}
}
