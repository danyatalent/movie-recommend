package movie

import (
	"github.com/danyatalent/movie-recommend/internal/genre"
)

type DTO struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Duration    int      `json:"duration"`
	Rating      float64  `json:"rating"`
	DirectorID  string   `json:"director_id"`
	GenresID    []string `json:"genres_id"`
}

type Movie struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Duration    int           `json:"duration"`
	Rating      float64       `json:"rating"`
	DirectorID  string        `json:"director_id"`
	Genres      []genre.Genre `json:"genres"`
}
