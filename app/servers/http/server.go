package http

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"go.uber.org/fx"

	"ru/kovardin/getapp/app/servers/http/config"
	"ru/kovardin/getapp/pkg/database"
)

type Server struct {
	db     *database.Database
	config config.Config
	serv   *http.Server

	modules []Routers
}

type Routers interface {
	Routes(r chi.Router)
}

func New(lc fx.Lifecycle, db *database.Database, config config.Config) *Server {
	s := &Server{
		db:      db,
		config:  config,
		modules: []Routers{},
	}

	s.serv = &http.Server{Addr: config.Address}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.serv.Handler = s.routing()
			s.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.Stop(ctx)
			return nil
		},
	})

	return s
}

func (s *Server) Start() {
	go func() {
		err := s.serv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.serv.Shutdown(ctx)
}

func (s *Server) routing() http.Handler {
	logger := httplog.NewLogger("httplog-example", httplog.Options{
		JSON:             true,
		LogLevel:         slog.LevelDebug,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
		// TimeFieldFormat: time.RFC850,
		QuietDownRoutes: []string{
			"/",
			"/ping",
		},
		QuietDownPeriod: 10 * time.Second,
		// SourceFieldName: "source",
	})

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(httplog.RequestLogger(logger))
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler, middleware.Recoverer, middleware.NoCache)

	for _, m := range s.modules {
		m.Routes(r)
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("^)"))
	})

	return r
}

func (s *Server) Routers(r Routers) {
	s.modules = append(s.modules, r)
}
