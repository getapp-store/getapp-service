package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"

	applications "ru/kovardin/getapp/app/modules/applications/models"
	"ru/kovardin/getapp/app/modules/warehouse/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Items struct {
	log   *logger.Logger
	items *database.Repository[models.Item]
	apps  *database.Repository[applications.Application]
}

func NewItems(log *logger.Logger, items *database.Repository[models.Item], apps *database.Repository[applications.Application]) *Items {
	return &Items{
		log:   log,
		items: items,
		apps:  apps,
	}
}

type ListResponse struct {
	Items map[string]json.RawMessage `json:"items"`
	Page  int                        `json:"page"`
	Size  int                        `json:"size"`
}

type ItemResponse struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

func (i *Items) Item(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		i.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	application, err := i.apps.First(database.Condition{
		In: map[string]any{"id": appid},
	})
	if err != nil {
		i.log.Error("error on get application from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if application.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	item, err := i.items.First(database.Condition{
		In: map[string]any{
			"application_id": appid,
			"active":         true,
			"key":            key,
		},
	})

	if err != nil {
		i.log.Error("error on get items from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if item.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp := ItemResponse{
		Key:   item.Key,
		Value: []byte(item.Value),
	}

	render.JSON(w, r, resp)
}

func (i *Items) Search(w http.ResponseWriter, r *http.Request) {
	// text search in Value
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	query := r.URL.Query().Get("query")

	if size == 0 {
		size = 10 // default size
	}

	start := page * size
	end := page*size + size

	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		i.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	application, err := i.apps.First(database.Condition{
		In: map[string]any{"id": appid},
	})
	if err != nil {
		i.log.Error("error on get application from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if application.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	condition := database.Condition{
		In: map[string]any{
			"application_id": appid,
			"active":         true,
		},
		Where: []database.Where{
			database.Where{
				Condition: "value LIKE ?",
				Values:    []any{"%" + query + "%"},
			},
		},
		Pagination: database.Paginating{
			Start: start,
			End:   end,
		},
		Sorting: database.Sorting{
			Sort:  "created_at",
			Order: "ASC",
		},
	}

	if query != "" {
		condition.Where = []database.Where{
			database.Where{
				Condition: "value LIKE ?",
				Values:    []any{"%" + query + "%"},
			},
		}
	}

	items, err := i.items.Find(condition)

	if err != nil {
		i.log.Error("error on get items from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := ListResponse{
		Items: map[string]json.RawMessage{},
		Page:  page,
		Size:  size,
	}

	for _, a := range items {
		resp.Items[a.Key] = []byte(a.Value)
	}

	render.JSON(w, r, resp)
}
