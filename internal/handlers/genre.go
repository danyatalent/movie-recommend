package handlers

import (
	"context"
	"errors"
	"github.com/danyatalent/movie-recommend/internal/genre"
	logging "github.com/danyatalent/movie-recommend/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"io"
	"log/slog"
	"net/http"
)

type Response struct {
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	genre.Genre `json:"genre,omitempty"`
}

// TODO: handle errors; add status codes into headers

func NewCreateGenre(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		req := struct {
			Name string `json:"name"`
		}{}
		err := render.DecodeJSON(r.Body, &req)

		if RequestBodyEmpty(err, log, w, r) {
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		// &genre.Genre{...} - not sure if it's good
		id, err := repository.CreateGenre(ctx, &genre.Genre{Name: req.Name})
		if err != nil {
			log.Error("failed to add genre", logging.Err(err))
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "failed to add genre",
			})
			return
		}
		log.Info("genre added", slog.String("uuid", id))

		render.JSON(w, r, Response{
			Status: "OK",
			Genre:  genre.Genre{ID: id},
		})
	}

}

func NewGetGenreByID(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")
			render.JSON(w, r, Response{Status: "Error", Error: "id is empty"})
		}
		genreByID, err := repository.GetGenreByID(ctx, id)
		if err != nil {
			log.Error("failed to get genreByID", logging.Err(err))
			render.JSON(w, r, Response{Status: "Error", Error: "failed to get genreByID"})
			return
		}
		log.Info("got genreByID", slog.Any("genreByID", genreByID))
		render.JSON(w, r, Response{Status: "OK", Genre: genreByID})
	}
}

func NewGetAllGenres(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		allGenres, err := repository.GetAllGenres(ctx)
		if err != nil {
			log.Error("failed to get all genres", logging.Err(err))
			render.JSON(w, r, Response{Status: "Error", Error: "failed to get all genres"})
			return
		}
		log.Info("successfully got all genres")
		render.JSON(w, r, allGenres)
	}
}

func NewUpdateGenre(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		req := struct {
			NewName string `json:"name"`
		}{}
		err := render.DecodeJSON(r.Body, &req)
		if RequestBodyEmpty(err, log, w, r) {
			return
		}
		log.Info("request body decoded", slog.Any("request", req))
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")
			render.JSON(w, r, Response{Status: "Error", Error: "id is empty"})
		}
		err = repository.UpdateGenre(ctx, id, req.NewName)
		if err != nil {
			log.Error("failed to update genre", logging.Err(err))
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "failed to update genre",
			})
			return
		}
		log.Info("genre updated", slog.String("id", id), slog.String("name", req.NewName))

		render.JSON(w, r, Response{
			Status: "OK",
			Genre:  genre.Genre{ID: id, Name: req.NewName},
		})
	}
}

func NewDeleteGenre(ctx context.Context, log *slog.Logger, repository genre.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")
			render.JSON(w, r, Response{Status: "Error", Error: "id is empty"})
		}
		err := repository.DeleteGenre(ctx, id)
		if err != nil {
			log.Error("failed to delete genre", logging.Err(err))
			render.JSON(w, r, Response{
				Status: "Error",
				Error:  "failed to delete genre",
			})
			return
		}
		log.Info("successfully deleted genre", slog.String("id", id))
		render.JSON(w, r, Response{
			Status: "OK",
		})
	}
}

func RequestBodyEmpty(err error, log *slog.Logger, w http.ResponseWriter, r *http.Request) bool {
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")

		render.JSON(w, r, Response{
			Status: "Error",
			Error:  "empty request",
		})
		return true
	}
	return false
}
