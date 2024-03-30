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
	ID          string        `json:"id" example:"dc26760a-42ba-4335-92f4-e9c0f1a2a838"`
	Name        string        `json:"name" example:"Dune"`
	Description string        `json:"description" example:"some text"`
	Duration    int           `json:"duration" example:"19200"`
	Rating      float64       `json:"rating" example:"7.5"`
	DirectorID  string        `json:"director_id" example:"0ac7ee25-2ebf-4edb-91eb-3d160a0428a8"`
	Genres      []genre.Genre `json:"genres"`
}
