package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/lokalize/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Languages struct {
	log       *logger.Logger
	languages *database.Repository[models.Language]
}

func NewLanguages(log *logger.Logger, languages *database.Repository[models.Language]) *Languages {
	return &Languages{
		log:       log,
		languages: languages,
	}
}

type LanguagesResponse struct {
	Items []Language `json:"items"`
}

type Language struct {
	Id     uint   `json:"id"`
	Name   string `json:"name"`
	Locale string `json:"locale"`
}

func (l *Languages) List(w http.ResponseWriter, r *http.Request) {
	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		l.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	languages, err := l.languages.Find(database.Condition{
		In: map[string]any{
			"application_id": appid,
		},
	})
	if err != nil {
		l.log.Error("error on get payment from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := LanguagesResponse{
		Items: []Language{},
	}

	for _, lang := range languages {
		resp.Items = append(resp.Items, Language{
			Id:     lang.ID,
			Name:   lang.Name,
			Locale: lang.Locale,
		})
	}

	render.JSON(w, r, resp)
}
