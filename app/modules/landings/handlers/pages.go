package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/landings/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Pages struct {
	log     *logger.Logger
	landing *database.Repository[models.Landing]
	pages   *database.Repository[models.Page]
}

func NewPages(log *logger.Logger, landing *database.Repository[models.Landing], pages *database.Repository[models.Page]) *Pages {
	return &Pages{
		log:     log,
		landing: landing,
		pages:   pages,
	}
}

func (p *Pages) Page(w http.ResponseWriter, r *http.Request) {
	landingPath := chi.URLParam(r, "landing")
	pagePage := chi.URLParam(r, "page")

	landing, err := p.landing.First(database.Condition{
		In: map[string]any{
			"path":   landingPath,
			"active": true,
		},
	})

	if err != nil {
		p.log.Error("error on get landing from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if landing.ID == 0 {
		w.WriteHeader(http.StatusFound)
		return
	}

	page, err := p.pages.First(database.Condition{
		In: map[string]any{
			"path":       pagePage,
			"landing_id": landing.ID,
			"active":     true,
		},
	})

	if err != nil {
		p.log.Error("error on get page from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if landing.ID == 0 {
		w.WriteHeader(http.StatusFound)
		return
	}

	render.HTML(w, r, page.Body)
}
