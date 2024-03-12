package request

import (
	"errors"
	"github.com/danyatalent/movie-recommend/pkg/response"
	"github.com/go-chi/render"
	"io"
	"log/slog"
	"net/http"
)

func BodyEmpty(err error, log *slog.Logger, w http.ResponseWriter, r *http.Request) bool {
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.Error("empty request"))
		return true
	}
	return false
}
