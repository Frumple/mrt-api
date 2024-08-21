package main

import (
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
	companies *orderedmap.OrderedMap[string, Company]
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

	companies := orderedMapToValues(provider.companies)

	if mode != "" {
		filtered_companies := []Company{}
		mode_exists := false

		for i := range companies {
			if string(companies[i].Mode) == mode {
				filtered_companies = append(filtered_companies, companies[i])
				mode_exists = true
			}
		}

		if !mode_exists {
			detail := "The 'mode' query parameter must be one of 'warp_rail', 'bus', 'air', 'sea', or 'other'."
			render.Render(writer, request, ErrorBadRequest(detail))
			return
		}

		companies = filtered_companies
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

	company, exists := provider.companies.Get(id)
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
	return CompanyProvider{companies: companies}
}
