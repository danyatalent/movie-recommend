package handlers

import (
	"context"
	"errors"
	"github.com/danyatalent/movie-recommend/internal/apperror"
	"github.com/danyatalent/movie-recommend/internal/director"
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

type DirectorRequest struct {
	FirstName string              `json:"first_name" validate:"required"`
	LastName  string              `json:"last_name" validate:"required"`
	Country   string              `json:"country" validate:"required"`
	BirthDate director.CustomDate `json:"birth_date" validate:"required"`
	HasOscar  bool                `json:"has_oscar" validate:"required"`
}

type DirectorResponse struct {
	response.Response
	Director director.Director `json:"director,omitempty"`
}

func DirectorResponseOK(w http.ResponseWriter, r *http.Request, director director.Director) {
	render.JSON(w, r, DirectorResponse{
		Response: response.OK(),
		Director: director,
	})
}

type Creator interface {
	CreateDirector(ctx context.Context, director *director.Director) (string, error)
}

func NewCreateDirector(ctx context.Context, log *slog.Logger, creator Creator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log = log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req DirectorRequest
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
		id, err := creator.CreateDirector(ctx, &director.Director{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			BirthDate: req.BirthDate,
			Country:   req.Country,
			HasOscar:  req.HasOscar,
		})
		if err != nil {
			if errors.Is(err, apperror.ErrEntityExists) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("director already exists"))
				return
			}
			log.Error("failed to add director", logging.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("director added", slog.String("id", id))
		w.WriteHeader(http.StatusCreated)
		DirectorResponseOK(w, r, director.Director{
			ID:        id,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			BirthDate: req.BirthDate,
			Country:   req.Country,
			HasOscar:  req.HasOscar,
		})
	}
}

type Getter interface {
	GetDirectorByID(ctx context.Context, id string) (director.Director, error)
}

func NewGetDirector(ctx context.Context, log *slog.Logger, getter Getter) http.HandlerFunc {
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
		directorByID, err := getter.GetDirectorByID(ctx, id)
		if err != nil {
			if errors.Is(err, apperror.ErrEntityNotFound) {
				log.Info("entity not found")
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("entity not found"))
				return
			}
			log.Error("failed to get director by ID", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get director by ID"))
			return
		}
		log.Info("got director by ID", slog.Any("director", directorByID))
		w.WriteHeader(http.StatusOK)
		DirectorResponseOK(w, r, directorByID)
	}
}
