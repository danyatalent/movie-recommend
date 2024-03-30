package handlers

import (
	"context"
	"errors"
	"github.com/danyatalent/movie-recommend/internal/apperror"
	"github.com/danyatalent/movie-recommend/internal/genre"
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

type GenreResponse struct {
	response.Response
	Genre genre.Genre `json:"genre,omitempty"`
}

type GenreRequest struct {
	Name  string `json:"name" validate:"required_without_all=Limit Page" example:"Comedy"`
	Limit int    `json:"limit" validate:"required_without=Name" example:"5"`
	Page  int    `json:"page" validate:"required_without=Name" example:"1"`
}

func GenreResponseOK(w http.ResponseWriter, r *http.Request, genre genre.Genre) {
	render.JSON(w, r, GenreResponse{
		Response: response.OK(),
		Genre:    genre,
	})
}

// TODO: handle errors; add status codes into headers

// NewCreateGenre godoc
//
// @Summary create genre
// @Description create genre by json
// @Tags genres
// @Accept json
// @Produce json
// @Param genre body GenreRequest true "Genre"
// @Success 200 {object} GenreResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /genres [post]
func NewCreateGenre(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log = log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req GenreRequest

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
		// &genre.Genre{...} - not sure if it's good
		id, err := repository.CreateGenre(ctx, &genre.Genre{
			Name: req.Name,
		})
		if err != nil {
			if errors.Is(err, apperror.ErrEntityExists) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("genre already exists"))
				return
			}
			log.Error("failed to add genre", logging.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("genre added", slog.String("uuid", id))

		w.WriteHeader(http.StatusCreated)
		GenreResponseOK(w, r, genre.Genre{
			ID:   id,
			Name: req.Name,
		})
	}

}

// NewGetGenreByID godoc
//
// @Summary get genre
// @Description get genre by id
// @Tags genres
// @Accept json
// @Produce json
// @Param id path int true "Genre ID"
// @Success 200 {object} GenreResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /genres/{id} [get]
func NewGetGenreByID(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log = log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("id is empty"))
		}
		genreByID, err := repository.GetGenreByID(ctx, id)
		if err != nil {
			if errors.Is(err, apperror.ErrEntityNotFound) {
				log.Info("entity not found")
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("entity not found"))
				return
			}

			log.Error("failed to get genreByID", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("got genreByID", slog.Any("genreByID", genreByID))
		w.WriteHeader(http.StatusOK)
		GenreResponseOK(w, r, genreByID)
	}
}

// NewGetAllGenres godoc
//
// @Summary get genres
// @Description get genres by page and limit
// @Tags genres
// @Accept json
// @Produce json
// @Param genre body GenreRequest true "Genre"
// @Success 200 {object} GenreResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /genres [get]
func NewGetAllGenres(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log = log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req GenreRequest

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

		allGenres, err := repository.GetAllGenres(ctx, req.Limit, req.Page)
		if err != nil {
			log.Error("failed to get all genres", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("successfully got all genres")
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, allGenres)
	}
}

// NewUpdateGenre godoc
//
// @Summary update genre
// @Description update genre by json
// @Tags genres
// @Accept json
// @Produce json
// @Param genre body GenreRequest true "Update Genre"
// @Param id path int true "Genre ID"
// @Success 200 {object} GenreResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /genres/{id} [put]
func NewUpdateGenre(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log = log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req GenreRequest
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

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("invalid request", logging.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("id is empty"))
		}
		err = repository.UpdateGenre(ctx, id, req.Name)
		if err != nil {
			if errors.Is(err, apperror.ErrEntityNotFound) {
				log.Info("entity not found")
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("entity not found"))
				return
			}
			log.Error("failed to update genre", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("genre updated", slog.String("id", id), slog.String("name", req.Name))

		w.WriteHeader(http.StatusOK)
		GenreResponseOK(w, r, genre.Genre{
			ID:   id,
			Name: req.Name,
		})
	}
}

// NewDeleteGenre godoc
//
// @Summary delete genre
// @Description delete genre by id
// @Tags genres
// @Accept json
// @Produce json
// @Param id path int true "Genre ID"
// @Success 200 {object} GenreResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /genres/{id} [delete]
func NewDeleteGenre(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log = log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("id is empty"))
		}
		err := repository.DeleteGenre(ctx, id)
		if err != nil {
			log.Error("failed to delete genre", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("internal error"))
			return
		}
		log.Info("successfully deleted genre", slog.String("id", id))
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, GenreResponse{
			Response: response.OK(),
		})
	}
}
