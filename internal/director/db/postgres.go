package director

import (
	"context"
	"errors"
	"fmt"
	"github.com/danyatalent/movie-recommend/internal/apperror"
	"github.com/danyatalent/movie-recommend/internal/director"
	"github.com/danyatalent/movie-recommend/pkg/client/postgresql"
	logging "github.com/danyatalent/movie-recommend/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
	"time"
)

type Repository struct {
	client postgresql.Client
	logger *slog.Logger
}

func NewRepository(client postgresql.Client, logger *slog.Logger) *Repository {
	return &Repository{
		client: client,
		logger: logger,
	}
}

func (r *Repository) CreateDirector(ctx context.Context, director *director.Director) (string, error) {
	q := "insert into directors(first_name, last_name, country, birth_date, has_oscar) VALUES ($1, $2, $3, $4, $5) returning id"
	r.logger.Info("creating director", slog.String("query", q))
	y, m, d := director.BirthDate.Date()
	//r.logger.Info("test date", slog.Any("year", y), slog.Any("month", m), slog.Any("day", d))

	date := fmt.Sprintf("%d-%v-%d", y, m, d)
	r.logger.Info("date", slog.String("date", date))
	if err := r.client.QueryRow(ctx, q, director.FirstName, director.LastName, director.Country,
		date, director.HasOscar).Scan(&director.ID); err != nil {
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
	return director.ID, nil
}

func (r *Repository) GetDirectorByID(ctx context.Context, id string) (director.Director, error) {
	q := "select id, first_name, last_name, country, birth_date, has_oscar from directors where id=$1"
	r.logger.Info("getting director by ID", slog.String("query", q))
	var d director.Director
	var date time.Time
	err := r.client.QueryRow(ctx, q, id).Scan(&d.ID, &d.FirstName, &d.LastName, &d.Country, &date, &d.HasOscar)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return director.Director{}, apperror.ErrEntityNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Code, pgErr.SQLState()))
			r.logger.Error("error due query", logging.Err(newErr))
			return director.Director{}, newErr
		}
		return director.Director{}, err
	}
	d.BirthDate = director.CustomDate{Time: date}
	return d, nil
}
