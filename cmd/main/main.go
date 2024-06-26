package main

import (
	"context"
	"fmt"
	_ "github.com/danyatalent/movie-recommend/docs"
	"github.com/danyatalent/movie-recommend/internal/config"
	director "github.com/danyatalent/movie-recommend/internal/director/db"
	genre "github.com/danyatalent/movie-recommend/internal/genre/db"
	"github.com/danyatalent/movie-recommend/internal/handlers"
	movie "github.com/danyatalent/movie-recommend/internal/movie/db"
	user "github.com/danyatalent/movie-recommend/internal/user/db"
	"github.com/danyatalent/movie-recommend/pkg/client/postgresql"
	logging "github.com/danyatalent/movie-recommend/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/swaggo/http-swagger"
	"log"
	"log/slog"
	"net/http"
	"os"
)

// @title Movie JSON API
// @version 1.0
// @description API Server for MovieRecommendation Service

// @host 158.160.124.149:3000
// @BasePath /
func main() {
	// Get configuration
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	cfg := config.GetConfig()
	ctx := context.Background()
	address := os.Getenv("ADDRESS")

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
	userRepository := user.NewRepository(postgresPool, logger)
	directorRepository := director.NewRepository(postgresPool, logger)
	movieRepository := movie.NewRepository(postgresPool, logger)

	// Init router and middlewares
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.URLFormat)

	// TODO: check context for DB operations
	// genres routing
	r.Route("/genres", func(r chi.Router) {
		r.Get("/{id}", handlers.NewGetGenreByID(ctx, logger, genreRepository))
		r.Post("/", handlers.NewCreateGenre(ctx, logger, genreRepository))
		r.Get("/", handlers.NewGetAllGenres(ctx, logger, genreRepository))
		r.Put("/{id}", handlers.NewUpdateGenre(ctx, logger, genreRepository))
		r.Delete("/{id}", handlers.NewDeleteGenre(ctx, logger, genreRepository))
	})

	// user routing
	r.Route("/users", func(r chi.Router) {
		r.Get("/{id}", handlers.NewGetUserByID(ctx, logger, userRepository))
		r.Post("/", handlers.NewCreateUser(ctx, logger, userRepository))
		r.Put("/{id}", handlers.NewUpdateUser(ctx, logger, userRepository))
	})

	// director routing
	r.Route("/directors", func(r chi.Router) {
		r.Get("/{id}", handlers.NewGetDirector(ctx, logger, directorRepository))
		r.Post("/", handlers.NewCreateDirector(ctx, logger, directorRepository))
	})

	// movie routing
	r.Route("/movies", func(r chi.Router) {
		r.Get("/{id}", handlers.NewGetMovie(ctx, logger, movieRepository))
		r.Post("/", handlers.NewCreateMovie(ctx, logger, movieRepository))
	})
	swaggerURL := fmt.Sprintf("http://%s/swagger/doc.json", address)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(swaggerURL),
	))
	// Configuration of server
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	// TODO: graceful shutdown
	logger.Info("starting server", slog.String("address", srv.Addr))
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server is down", logging.Err(err))
		os.Exit(1)
	}

}
