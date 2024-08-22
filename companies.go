package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type TransportMode string

const (
	WarpRail TransportMode = "warp_rail"
	Bus      TransportMode = "bus"
	Air      TransportMode = "air"
	Sea      TransportMode = "sea"
	Other    TransportMode = "other"
)

var transportModes = []TransportMode{
	WarpRail,
	Bus,
	Air,
	Sea,
	Other,
}

type Company struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Pattern string        `json:"pattern"`
	Mode    TransportMode `json:"mode"`
}

func (company Company) GetID() string {
	return company.ID
}

func (company Company) Render(writer http.ResponseWriter, request *http.Request) error {
	return nil
}

type CompanyProvider struct {
	companies       []Company
	companiesByID   *orderedmap.OrderedMap[string, Company]
	companiesByMode *orderedmap.OrderedMap[TransportMode, []Company]
}

// getCompanies godoc
// @summary     List all companies
// @description List all companies (defined in https://github.com/Frumple/mrt-api/blob/main/data/companies.yml).
// @tags        Companies
// @produce     json
// @param       mode query   string false "Filter by transport mode: `warp_rail`, `bus`, `air`, `sea`, or `other`."
// @success     200  {array} Company
// @router      /companies [get]
func (provider CompanyProvider) getCompanies(writer http.ResponseWriter, request *http.Request) {
	mode := request.URL.Query().Get("mode")

	companies := provider.companies

	if mode != "" {
		mode_exists := false
		companies, mode_exists = provider.companiesByMode.Get(TransportMode(mode))

		if !mode_exists {
			detail := "The 'mode' query parameter must be one of 'warp_rail', 'bus', 'air', 'sea', or 'other'."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}
	}

	err := render.RenderList(writer, request, toRenderList(companies))
	if err != nil {
		render.Render(writer, request, ErrorRender(err))
		return
	}
}

// getCompanyById godoc
// @summary       Get company by ID
// @description   Get company by ID (defined in https://github.com/Frumple/mrt-api/blob/main/data/companies.yml).
// @tags          Companies
// @produce       json
// @param         id  path     string  true "Company ID"
// @success       200 {object} Company
// @failure       404 {object} Error
// @router        /companies/{id} [get]
func (provider CompanyProvider) getCompanyById(writer http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	company, exists := provider.companiesByID.Get(id)
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

func companiesRouter(provider CompanyProvider) http.Handler {
	router := chi.NewRouter()
	router.Get("/", provider.getCompanies)

	router.Route("/{id}", func(subrouter chi.Router) {
		subrouter.Get("/", provider.getCompanyById)
	})

	return router
}

func loadCompanies() CompanyProvider {
	companies := loadStaticData[Company](COMPANIES_PATH)
	companiesByID := staticDataToOrderedMap(companies)
	companiesByMode := orderedmap.New[TransportMode, []Company]()

	// Populate map of transport modes to list of companies
	for i := range transportModes {
		companiesByMode.Set(transportModes[i], []Company{})
	}
	for i := range companies {
		company := companies[i]
		list, exists := companiesByMode.Get(company.Mode)

		if !exists {
			message := fmt.Sprintf("The company '%s' has an invalid mode: '%s'", company.ID, company.Mode)
			panic(message)
		}

		companiesByMode.Set(company.Mode, append(list, company))
	}

	return CompanyProvider{
		companies:       companies,
		companiesByID:   companiesByID,
		companiesByMode: companiesByMode,
	}
}
