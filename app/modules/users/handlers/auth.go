package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ru/kovardin/getapp/app/modules/users/models"
	"ru/kovardin/getapp/pkg/database"
	"ru/kovardin/getapp/pkg/logger"
)

type Authorization struct {
	log       *logger.Logger
	auth      *database.Repository[models.Auth]
	templates map[string]*template.Template
}

func NewAuthorization(
	log *logger.Logger,
	auth *database.Repository[models.Auth],
) *Authorization {
	return &Authorization{
		log:  log,
		auth: auth,
		templates: map[string]*template.Template{
			"choose": template.Must(template.ParseFiles(
				"templates/users/choose.gohtml",
			)),
		},
	}
}

func (a *Authorization) Choose(w http.ResponseWriter, r *http.Request) {
	application, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		a.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authorizations, err := a.auth.Find(database.Condition{
		In: map[string]any{
			"application_id": application,
		},
		Where: []database.Where{
			{
				Condition: "active = true",
			},
		},
	})
	if err != nil {
		a.log.Error("error on getting auth", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(authorizations) == 0 {
		a.log.Error("auth not found", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := a.templates["choose"].ExecuteTemplate(w, "choose", struct {
		Title          string
		Application    int
		Authorizations []models.Auth
	}{
		Title:          "Authorization",
		Application:    application,
		Authorizations: authorizations,
	}); err != nil {
		a.log.Error("auth error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
