package genre

import "context"

type Repository interface {
	CreateGenre(ctx context.Context, genre *Genre) (string, error)
	GetGenreByID(ctx context.Context, id string) (Genre, error)
	GetAllGenres(ctx context.Context) ([]Genre, error)
	UpdateGenre(ctx context.Context, id, newName string) error
	DeleteGenre(ctx context.Context, id string) error
}
