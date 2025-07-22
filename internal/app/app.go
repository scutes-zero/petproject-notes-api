package app

import (
	"log/slog"
	"net/http"
	"notes-api/internal/config"
	"notes-api/internal/handlers/auth"
	"notes-api/internal/handlers/notes"
	"notes-api/internal/middleware"
	"notes-api/internal/storage"
	"notes-api/pkg/logger"

	"github.com/go-chi/chi/v5"
	chiMW "github.com/go-chi/chi/v5/middleware"
)

type App struct {
	config    *config.Config
	storage   *storage.Storage
	logger    *slog.Logger
	jwtSecret []byte
}

func NewApp(config *config.Config, storage *storage.Storage, logger *slog.Logger, jwtSecret []byte) *App {
	return &App{config: config, storage: storage, logger: logger, jwtSecret: jwtSecret}
}

func (a *App) AddRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(chiMW.RequestID)
	r.Use(middleware.LoggerMiddleware(a.logger))

	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", auth.RegisterHandler(a.logger, a.storage))
		r.Post("/signin", auth.LoginHandler(a.logger, a.storage, a.jwtSecret))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuthMiddleware(a.jwtSecret))

		r.Route("/notes", func(r chi.Router) {

			r.Get("/", notes.NotesHandler(a.logger, a.storage))
			r.Get("/{id}", notes.NoteHandler(a.logger, a.storage))
			r.Post("/", notes.CreateNoteHandler(a.logger, a.storage))
			r.Delete("/{id}", notes.DeleteNoteHandler(a.logger, a.storage))
			r.Put("/{id}", notes.UpdateNoteHandler(a.logger, a.storage))
		})
	})

	return r
}

func (a *App) Start() error {
	srv := http.Server{
		Addr:         a.config.HTTPServer.Address,
		Handler:      a.AddRoutes(),
		ReadTimeout:  a.config.HTTPServer.Timeout,
		WriteTimeout: a.config.HTTPServer.Timeout,
		IdleTimeout:  a.config.HTTPServer.IdleTimeout,
	}

	a.logger.Info("starting server", slog.String("address", a.config.HTTPServer.Address))

	if err := srv.ListenAndServe(); err != nil {
		a.logger.Error("failed to start server", logger.Err(err))
		return err
	}

	return nil
}
