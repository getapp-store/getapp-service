package handlers

import (
	"go.uber.org/zap"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/billing/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Products struct {
	log          *logger.Logger
	products     *database.Repository[models.Product]
	applications *database.Repository[applications.Application]
}

func NewProducts(log *logger.Logger, products *database.Repository[models.Product], applications *database.Repository[applications.Application]) *Products {
	return &Products{
		log:          log,
		products:     products,
		applications: applications,
	}
}

type Product struct {
	Id     uint   `json:"id"`
	Name   string `json:"name"`
	Title  string `json:"title"`
	Amount int    `json:"amount"`
}

type ProductsResponse struct {
	Items []Product `json:"items"`
}

func (p *Products) List(w http.ResponseWriter, r *http.Request) {
	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		p.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	application, err := p.applications.First(database.Condition{
		In: map[string]any{"id": appid},
	})
	if err != nil {
		p.log.Error("error on get application from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if application.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	items, err := p.products.Find(database.Condition{
		In: map[string]any{
			"application_id": application.ID,
			"active":         true,
		},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ProductsResponse{
		Items: []Product{},
	}

	for _, item := range items {
		resp.Items = append(resp.Items, Product{
			Id:     item.ID,
			Name:   item.Name,
			Title:  item.Title,
			Amount: item.Amount,
		})
	}

	render.JSON(w, r, resp)
}
