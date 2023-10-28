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

type Phrases struct {
	log       *logger.Logger
	phrases   *database.Repository[models.Phrase]
	languages *database.Repository[models.Language]
}

func NewPhrases(log *logger.Logger, phrases *database.Repository[models.Phrase], languages *database.Repository[models.Language]) *Phrases {
	return &Phrases{
		log:       log,
		phrases:   phrases,
		languages: languages,
	}
}

func (p *Phrases) List(w http.ResponseWriter, r *http.Request) {
	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		p.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	locale := chi.URLParam(r, "locale")
	if locale == "" {
		p.log.Error("error on get locale")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	language, err := p.languages.First(database.Condition{
		In: map[string]any{
			"locale":         locale,
			"application_id": appid,
		},
	})

	if err != nil {
		p.log.Error("error on get language", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	phrases, err := p.phrases.Find(database.Condition{
		In: map[string]any{
			"language_id": language.ID,
		},
	})
	if err != nil {
		p.log.Error("error on get payment from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := map[string]string{}

	for _, phrase := range phrases {
		resp[phrase.Key] = phrase.Value
	}

	render.JSON(w, r, resp)
}
