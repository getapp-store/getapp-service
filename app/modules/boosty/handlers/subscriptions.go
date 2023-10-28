package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/boosty/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Subscriptions struct {
	log           *logger.Logger
	subscriptions *database.Repository[models.Subscription]
	applications  *database.Repository[applications.Application]
}

func NewSubscriptions(
	log *logger.Logger,
	subscriptions *database.Repository[models.Subscription],
	applications *database.Repository[applications.Application],
) *Subscriptions {
	return &Subscriptions{
		log:           log,
		subscriptions: subscriptions,
		applications:  applications,
	}
}

type Subscription struct {
	Id       uint   `json:"id"`
	External int    `json:"external"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Blog     string `json:"blog"`
	Amount   int    `json:"amount"`
	Active   bool   `json:"active"`
}

type SubscriptionsResponse struct {
	Items []Subscription `json:"items"`
}

func (s *Subscriptions) Subscriptions(w http.ResponseWriter, r *http.Request) {
	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		s.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	application, err := s.applications.First(database.Condition{
		In: map[string]any{"id": appid},
	})
	if err != nil {
		s.log.Error("error on get application from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if application.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	items, err := s.subscriptions.Find(database.Condition{
		In: map[string]any{
			"\"Blog\".application_id": application.ID,
			"subscriptions.active":    true,
			"\"Blog\".active":         true,
		},
		Preload: []string{
			"Blog",
		},
		Joins: []string{
			"Blog",
		},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := SubscriptionsResponse{
		Items: []Subscription{},
	}

	for _, item := range items {
		resp.Items = append(resp.Items, Subscription{
			Id:       item.ID,
			External: item.External,
			Name:     item.Name,
			Title:    item.Title,
			Amount:   item.Amount,
			Blog:     item.Blog.Title,
			Active:   item.Active,
		})
	}

	render.JSON(w, r, resp)
}
