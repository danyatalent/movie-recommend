package handlers

import (
	"context"
	"errors"
	"github.com/danyatalent/movie-recommend/internal/apperror"
	"github.com/danyatalent/movie-recommend/internal/user"
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

type UserResponse struct {
	response.Response
	user.User `json:"user,omitempty"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func UserResponseOK(w http.ResponseWriter, r *http.Request, u user.User) {
	render.JSON(w, r, UserResponse{
		Response: response.OK(),
		User:     u,
	})
}

type UserCreator interface {
	CreateUser(ctx context.Context, user *user.User) (string, error)
}

func NewCreateUser(ctx context.Context, log *slog.Logger, creator UserCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req CreateUserRequest
		err := render.DecodeJSON(r.Body, &req)
		if request.BodyEmpty(err, log, w, r) {
			return
		}
		if err != nil {
			log.Error("failed to decode request body", logging.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))
		}

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("invalid request", logging.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}
		log.Info("request body decoded", slog.Any("request", req))
		id, err := creator.CreateUser(ctx, &user.User{Name: req.Username, Password: req.Password, Email: req.Email})
		if err != nil {
			if errors.Is(err, apperror.ErrEntityExists) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, response.Error("user already exist"))
				return
			}
			log.Error("failed to add user", logging.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to add user"))
			return
		}
		log.Info("user added", slog.String("uuid", id))
		w.WriteHeader(http.StatusCreated)
		UserResponseOK(w, r, user.User{
			ID:       id,
			Name:     req.Username,
			Password: req.Password,
			Email:    req.Email,
		})
	}
}

type UserUpdater interface {
	UpdateUserName(ctx context.Context, id, newName string) error
	UpdateUserPassword(ctx context.Context, id, newPass string) error
}

func NewUpdateUser(ctx context.Context, log *slog.Logger, updater UserUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log = log.With(
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		var req UpdateUserRequest
		err := render.DecodeJSON(r.Body, &req)
		if request.BodyEmpty(err, log, w, r) {
			return
		}
		if err != nil {
			log.Error("failed to decode request body", logging.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))
		}
		log.Info("request body decoded", slog.Any("req", req))
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("id is empty"))
		}

		// Validation request for update
		if req.Username == "" && req.Password != "" {
			if err = updater.UpdateUserPassword(ctx, id, req.Password); err != nil {
				if errors.Is(err, apperror.ErrEntityNotFound) {
					log.Info("entity not found")
					w.WriteHeader(http.StatusNotFound)
					render.JSON(w, r, response.Error("entity not found"))
					return
				}
				log.Error("failed to update user password", logging.Err(err))
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, response.Error("failed to update user password"))
				return
			}
			w.WriteHeader(http.StatusOK)
			UserResponseOK(w, r, user.User{
				ID:       id,
				Password: req.Password,
			})
			return
		} else if req.Username != "" && req.Password == "" {
			if err = updater.UpdateUserName(ctx, id, req.Username); err != nil {
				if errors.Is(err, apperror.ErrEntityNotFound) {
					log.Info("entity not found")
					w.WriteHeader(http.StatusNotFound)
					render.JSON(w, r, response.Error("entity not found"))
					return
				}
				log.Error("failed to update username", logging.Err(err))
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, response.Error("failed to update username"))
				return
			}
			w.WriteHeader(http.StatusOK)
			UserResponseOK(w, r, user.User{ID: id, Name: req.Username})
			return
		} else if req.Username == "" && req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("at least 1 field (username or password) must be not empty"))
			return
		}
		err = updater.UpdateUserPassword(ctx, id, req.Password)
		if err != nil {
			if errors.Is(err, apperror.ErrEntityNotFound) {
				log.Info("entity not found")
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("entity not found"))
				return
			}
			log.Error("failed to update user password", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to update user password"))
			return
		}
		err = updater.UpdateUserName(ctx, id, req.Username)
		if err != nil {
			if errors.Is(err, apperror.ErrEntityNotFound) {
				log.Info("entity not found")
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("entity not found"))
				return
			}
			log.Error("failed to update username", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to update username"))
			return
		}
		w.WriteHeader(http.StatusOK)
		UserResponseOK(w, r, user.User{ID: id, Name: req.Username, Password: req.Password})
	}
}

type UserGetter interface {
	GetUserByID(ctx context.Context, id string) (user.User, error)
}

func NewGetUserByID(ctx context.Context, log *slog.Logger, getter UserGetter) http.HandlerFunc {
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
			return
		}
		userByID, err := getter.GetUserByID(ctx, id)
		if err != nil {
			if errors.Is(err, apperror.ErrEntityNotFound) {
				log.Info("entity not found")
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, response.Error("entity not found"))
				return
			}
			log.Error("failed to get user by ID", logging.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get user by ID"))
			return
		}
		log.Info("got user by id", slog.Any("user", userByID))
		w.WriteHeader(http.StatusOK)
		UserResponseOK(w, r, userByID)
	}
}
