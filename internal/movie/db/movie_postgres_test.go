package movie

import (
	"context"
	"github.com/danyatalent/movie-recommend/internal/config"
	"github.com/danyatalent/movie-recommend/internal/genre"
	"github.com/danyatalent/movie-recommend/internal/movie"
	"github.com/danyatalent/movie-recommend/pkg/client/postgresql"
	logger2 "github.com/danyatalent/movie-recommend/pkg/logger"
	"log/slog"
	"math"
	"testing"
)

func TestRepository_Create(t *testing.T) {
	logger := logger2.InitLogger(logger2.InfoLevel)
	client, err := postgresql.NewClient(context.TODO(), logger, config.Storage{
		Host:     "localhost",
		Port:     "5432",
		Database: "postgres",
		Username: "postgres",
		Password: "postgres",
	})
	if err != nil {
		t.Errorf("cannot create postgres repo")
	}
	m := movie.DTO{
		Name:        "Dune",
		Description: "some text",
		Duration:    19200,
		Rating:      8.0,
		DirectorID:  "82a276e5-b852-48d6-a023-7b119faa76e6",
		GenresID:    []string{"59457b31-89f8-4ade-b46c-731c61430c3e", "23bb4312-7fc3-4238-aaca-0d27b0a11fb3"},
	}
	repo := NewRepository(client, logger)
	if _, err := repo.CreateMovie(context.TODO(), &m); err != nil {
		t.Errorf("can't create movie, err: %v", err)
	}
}

func TestRepository_GetMovie(t *testing.T) {
	logger := logger2.InitLogger(logger2.InfoLevel)
	client, err := postgresql.NewClient(context.TODO(), logger, config.Storage{
		Host:     "localhost",
		Port:     "5432",
		Database: "postgres",
		Username: "postgres",
		Password: "postgres",
	})
	if err != nil {
		t.Errorf("cannot create postgres repo")
	}
	repo := NewRepository(client, logger)
	id := "dc26760a-42ba-4335-92f4-e9c0f1a2a838"
	m, err := repo.GetMovie(context.TODO(), id)
	logger.Info("test movie", slog.Any("movie", m))
	if err != nil {
		t.Errorf("can't get movie by id, err: %v", err)
	}
	if m.ID != id {
		t.Errorf("wrong id")
	}
	if m.Name != "Dune" {
		t.Errorf("wrong name")
	}
	if m.Description != "some text" {
		t.Errorf("wrong description")
	}
	if m.Duration != 19200 {
		t.Errorf("wrong duration")
	}
	epsilon := 0.0001
	if math.Abs(m.Rating-8.0) > epsilon {
		t.Errorf("wrong rating")
	}
	if m.DirectorID != "82a276e5-b852-48d6-a023-7b119faa76e6" {
		t.Errorf("wrong director id")
	}

	if !contains(m.Genres, "23bb4312-7fc3-4238-aaca-0d27b0a11fb3") {
		t.Errorf("did not find necessary id")
	}
	if !contains(m.Genres, "59457b31-89f8-4ade-b46c-731c61430c3e") {
		t.Errorf("did not find necessary id")
	}
}

func contains(genres []genre.Genre, id string) bool {
	for _, g := range genres {
		if g.ID == id {
			return true
		}
	}
	return false
}
