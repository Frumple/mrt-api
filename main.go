package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/frumple/mrt-api/docs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	_ "github.com/go-sql-driver/mysql"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

const (
	DB_CONFIG_PATH = "config/db_config.yml"
	COMPANIES_PATH = "data/companies.yml"
	WORLDS_PATH    = "data/worlds.yml"
)

const MAX_THROTTLE = 3

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type StaticData interface {
	Company | World
	GetID() string
}

// @title                    Minecart Rapid Transit Server API
// @version                  2.0.0
// @description              Provides data from the Minecart Rapid Transit (MRT) server.

// @host                     api.minecartrapidtransit.net
// @BasePath	               /api/v2

// @externalDocs.description GitHub Repository
// @externalDocs.url         https://github.com/Frumple/mrt-api

func main() {
	db := initializeDatabase()
	defer db.Close()

	companyProvider := loadCompanies()
	worldProvider := loadWorlds()
	warpProviderV1 := WarpProviderV1{
		db:              db,
		companyProvider: companyProvider,
		worldProvider:   worldProvider,
	}
	warpProviderV2 := WarpProviderV2{
		db:              db,
		companyProvider: companyProvider,
		worldProvider:   worldProvider,
	}

	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Heartbeat("/ping"))
	router.Use(middleware.Throttle(MAX_THROTTLE))
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Mount("/warps", warpsRouter(warpProviderV1))
			r.Mount("/companies", companiesRouter(companyProvider))
			r.Mount("/worlds", worldsRouter(worldProvider))
		})

		r.Route("/v2", func(r chi.Router) {
			r.Mount("/warps", warpsRouter(warpProviderV2))
			r.Mount("/companies", companiesRouter(companyProvider))
			r.Mount("/worlds", worldsRouter(worldProvider))
		})
	})

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	addr := ":8080"
	log.Println("API server started on: ", addr)
	log.Fatal(http.ListenAndServe(addr, router))
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

func loadStaticData[V StaticData](yamlFilePath string) []V {
	vSlice := []V{}

	data, err := os.ReadFile(yamlFilePath)
	checkForErrors(err)

	err = yaml.Unmarshal([]byte(data), &vSlice)
	checkForErrors(err)

	return vSlice
}

func staticDataToOrderedMap[V StaticData](vSlice []V) *orderedmap.OrderedMap[string, V] {
	vMap := orderedmap.New[string, V]()

	for _, value := range vSlice {
		vMap.Set(value.GetID(), value)
	}

	return vMap
}

func toRenderList[V render.Renderer](vSlice []V) []render.Renderer {
	list := []render.Renderer{}
	for _, v := range vSlice {
		list = append(list, v)
	}
	return list
}

func checkForErrors(err error) {
	if err != nil {
		panic(err)
	}
}

type Error struct {
	Error          error `json:"-"`
	HTTPStatusCode int   `json:"-"`

	Message   string `json:"message"`
	Detail    string `json:"detail,omitempty"`
	ErrorText string `json:"error,omitempty"`
}

func (response *Error) Render(writer http.ResponseWriter, request *http.Request) error {
	render.Status(request, response.HTTPStatusCode)
	return nil
}

func ErrorBadRequest(detail string) render.Renderer {
	return &Error{
		HTTPStatusCode: 400,
		Message:        "Bad request.",
		Detail:         detail,
	}
}

func ErrorRender(err error) render.Renderer {
	return &Error{
		Error:          err,
		HTTPStatusCode: 422,
		Message:        "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrorNotFound = &Error{
	HTTPStatusCode: 404,
	Message:        "Resource not found.",
}
