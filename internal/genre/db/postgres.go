package genre

import (
	"context"
	"errors"
	"fmt"
	"github.com/danyatalent/movie-recommend/internal/apperror"
	"github.com/danyatalent/movie-recommend/internal/genre"
	"github.com/danyatalent/movie-recommend/pkg/client/postgresql"
	logging "github.com/danyatalent/movie-recommend/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
)

type repository struct {
	client postgresql.Client
	logger *slog.Logger
}

func (r *repository) GetAllGenres(ctx context.Context, pageSize, pageNumber int) ([]genre.Genre, error) {
	q := "select id, name from genres order by id limit $1 offset $2"
	offset := (pageNumber - 1) * pageSize
	r.logger.Info("getting all genres", slog.String("query", q))
	rows, err := r.client.Query(ctx, q, pageSize, offset)
	if err != nil {
		return nil, err
	}

	genres := make([]genre.Genre, 0)

	for rows.Next() {
		var g genre.Genre
		err = rows.Scan(&g.ID, &g.Name)
		if err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return genres, nil
}

func (r *repository) UpdateGenre(ctx context.Context, id, newName string) error {
	q := "update genres set name=$1 where id=$2"
	r.logger.Debug("updating genre", slog.String("query", q))
	result, err := r.client.Exec(ctx, q, newName, id)
	if result.RowsAffected() == 0 {
		return apperror.ErrEntityNotFound
	}
	if err != nil {
		return fmt.Errorf("can't exec query: %s", q)
	}
	return nil
}

func (r *repository) DeleteGenre(ctx context.Context, id string) error {
	q := "delete from genres where id=$1"
	r.logger.Debug("deleting from genre", slog.String("query", q))
	_, err := r.client.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("can't delete from genres")
	}
	return nil
}

func (r *repository) GetGenreByID(ctx context.Context, id string) (genre.Genre, error) {
	q := "select id, name from genres where id = $1"
	r.logger.Debug("getting genre by id", slog.String("query", q))
	var g genre.Genre
	if err := r.client.QueryRow(ctx, q, id).Scan(&g.ID, &g.Name); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if errors.Is(err, pgx.ErrNoRows) {
				return genre.Genre{}, apperror.ErrEntityNotFound
			}
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s,  Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Code, pgErr.SQLState()))
			r.logger.Error("error due query", logging.Err(newErr))
			return genre.Genre{}, newErr
		}
		return genre.Genre{}, err
	}
	return g, nil
}

func (r *repository) CreateGenre(ctx context.Context, genre *genre.Genre) (string, error) {
	q := "insert into genres(name) values ($1) returning id"
	r.logger.Info("creating genre", slog.String("query", q))
	if err := r.client.QueryRow(ctx, q, genre.Name).Scan(&genre.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.SQLState() == apperror.ErrConstraintUniqueCode {
				return "", apperror.ErrEntityExists
			}

			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Code, pgErr.SQLState()))
			r.logger.Error("error due query", logging.Err(newErr))
			return "", newErr
		}
		return "", err
	}

	return genre.ID, nil
}

func NewRepository(client postgresql.Client, logger *slog.Logger) genre.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}
