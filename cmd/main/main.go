package main

import (
	"context"
	"fmt"
	"github.com/danyatalent/movie-recommend/internal/config"
	genre "github.com/danyatalent/movie-recommend/internal/genre/db"
	"github.com/danyatalent/movie-recommend/internal/handlers"
	"github.com/danyatalent/movie-recommend/pkg/client/postgresql"
	logging "github.com/danyatalent/movie-recommend/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

var (
	configPath string
)

func main() {
	// Get configuration
	cfg := config.GetConfig()

	// Init logger
	logger := logging.InitLogger(cfg.LogLevel)
	logger.Info("config parsed", slog.String("log-level", cfg.LogLevel))

	// Connect to Database
	postgresPool, err := postgresql.NewClient(logger, context.TODO(), 3, cfg.Storage)
	if err != nil {
		logger.Error("cannot connect to postgres", logging.Err(err))
	}
	// Close connection
	defer postgresPool.Close()

	// Testing connection genre
	genreRepository := genre.NewRepository(postgresPool, logger)
	//g := genre2.Genre{
	//	Name: "Comedy",
	//}
	//_, err = genreRepository.CreateGenre(context.Background(), &g)
	//if err != nil {
	//	logger.Error("can't create genre", logging.Err(err))
	//}

	// Testing connection movie
	//repository := movie.NewRepository(postgresPool, logger)
	//m := mv.Movie{
	//	Name:        "Interstellar",
	//	Description: "movie about space",
	//	Duration:    2000 * time.Second,
	//	Rating:      8.65,
	//}
	//err = repository.Create(context.Background(), &m)
	//if err != nil {
	//	logger.Error("can't create movie", logging.Err(err))
	//}

	// Testing connection in general
	//var greeting string
	//err = postgresPool.QueryRow(context.Background(), "select 'Hello world!'").Scan(&greeting)
	//if err != nil {
	//	logger.Error("QueryRow failed", logging.Err(err))
	//}
	//logger.Info(greeting)

	// Init router and middlewares
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.URLFormat)

	// create genre
	r.Post("/genre", handlers.NewCreateGenre(context.Background(), logger, genreRepository))

	// get all genres
	r.Get("/genres", handlers.NewGetAllGenres(context.TODO(), logger, genreRepository))

	// get genre by id
	r.Get("/genres/{id}", handlers.NewGetGenreByID(context.Background(), logger, genreRepository))

	r.Put("/genres/{id}", handlers.NewUpdateGenre(context.Background(), logger, genreRepository))
	r.Delete("/genres/{id}", handlers.NewDeleteGenre(context.Background(), logger, genreRepository))
	// /{name} - hello {name}
	r.Get("/{name}", NameHandler(logger))

	// Configuration of server
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	logger.Info("starting server", slog.String("address", srv.Addr))
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server is down", logging.Err(err))
		os.Exit(1)
	}

}

// NameHandler - Handler function for /{name}
func NameHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if name := chi.URLParam(r, "name"); name != "" {
			if _, err := w.Write([]byte(fmt.Sprintf("Hello %s", name))); err != nil {
				log.Error("can't write name", logging.Err(err))
			}
		}
	}
}
