package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Company struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
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

func (provider CompanyProvider) getCompanies(writer http.ResponseWriter, request *http.Request) {
	err := render.RenderList(writer, request, toRenderList(orderedMapToValues(provider.companies)))
	if err != nil {
		render.Render(writer, request, ErrorRender(err))
		return
	}
}

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
