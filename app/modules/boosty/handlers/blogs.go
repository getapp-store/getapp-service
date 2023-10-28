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

type Blogs struct {
	log          *logger.Logger
	applications *database.Repository[applications.Application]
	blogs        *database.Repository[models.Blog]
}

func NewBlogs(log *logger.Logger, applications *database.Repository[applications.Application], blogs *database.Repository[models.Blog]) *Blogs {
	return &Blogs{
		log:          log,
		applications: applications,
		blogs:        blogs,
	}
}

type BlogResponse struct {
	Id    uint   `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (b *Blogs) Blog(w http.ResponseWriter, r *http.Request) {
	appid, err := strconv.Atoi(chi.URLParam(r, "application"))
	if err != nil {
		b.log.Error("error on get application", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	application, err := b.applications.First(database.Condition{
		In: map[string]any{"id": appid},
	})
	if err != nil {
		b.log.Error("error on get application from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if application.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	blog, err := b.blogs.First(database.Condition{
		In: map[string]any{
			"application_id": appid,
			"active":         true,
		},
	})
	if err != nil {
		b.log.Error("error on get blog from db", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if blog.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	render.JSON(w, r, BlogResponse{
		Id:    blog.ID,
		Name:  blog.Name,
		Title: blog.Title,
		Url:   blog.Url,
	})
}
