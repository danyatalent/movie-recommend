package handlers

import (
	"context"
	"errors"
	"github.com/danyatalent/movie-recommend/internal/apperror"
	"github.com/danyatalent/movie-recommend/internal/genre"
	"github.com/danyatalent/movie-recommend/internal/movie"
	logging "github.com/danyatalent/movie-recommend/pkg/logger"
	"github.com/danyatalent/movie-recommend/pkg/request"
	"github.com/danyatalent/movie-recommend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

type RequestMovie struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Duration    int      `json:"duration" validate:"required"`
	Rating      float64  `json:"rating" validate:"required"`
	DirectorID  string   `json:"director_id" validate:"required"`
	GenresID    []string `json:"genres_id" validate:"required"`
}

type ResponseMovie struct {
	response.Response
	Movie movie.Movie `json:"movie,omitempty"`
}

func MovieResponseOK(w http.ResponseWriter, r *http.Request, movie2 movie.Movie) {
	render.JSON(w, r, ResponseMovie{
		Response: response.OK(),
		Movie:    movie2,
	})
}

type MovieGetter interface {
	GetMovie(ctx context.Context, id string) (movie.Movie, error)
}

func NewGetMovie(ctx context.Context, log *slog.Logger, getter MovieGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log = log.With(
			slog.String("request_id", middleware.GetReqID(ctx)),
		)
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("id is empty"))
			return
		}
		m, err := getter.GetMovie(ctx, id)
		if err != nil {
			if errors.Is(err, apperror.ErrEntityNotFound) {
				log.Info("entity not found")
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("entity not found"))
				return
			}
			log.Error("failed to get movie by ID", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get movie by ID"))
			return
		}
		log.Info("got director by ID", slog.Any("movie", m))
		w.WriteHeader(http.StatusOK)
		MovieResponseOK(w, r, m)
	}
}

type MovieCreator interface {
	CreateMovie(ctx context.Context, dto *movie.DTO) (string, error)
	GetGenresByMovie(ctx context.Context, id string) ([]genre.Genre, error)
}

func NewCreateMovie(ctx context.Context, log *slog.Logger, creator MovieCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log = log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req RequestMovie
		err := render.DecodeJSON(r.Body, &req)

		if request.BodyEmpty(err, log, w, r) {
			return
		}
		if err != nil {
			log.Error("failed to decode request body", logging.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("invalid request", logging.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}
		id, err := creator.CreateMovie(ctx, &movie.DTO{
			Name:        req.Name,
			Description: req.Description,
			Duration:    req.Duration,
			Rating:      req.Rating,
			DirectorID:  req.DirectorID,
			GenresID:    req.GenresID,
		})
		if err != nil {
			if errors.Is(err, apperror.ErrEntityExists) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("movie already exists"))
				return
			}
			log.Error("failed to create movie", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		genres, err := creator.GetGenresByMovie(ctx, id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("can't find such genres"))
			return
		}
		log.Info("movie created", slog.String("id", id))
		w.WriteHeader(http.StatusCreated)
		MovieResponseOK(w, r, movie.Movie{
			ID:          id,
			Name:        req.Name,
			Description: req.Description,
			Duration:    req.Duration,
			Rating:      req.Rating,
			DirectorID:  req.DirectorID,
			Genres:      genres,
		})
	}
}
