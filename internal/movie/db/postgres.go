package movie

import (
	"context"
	"errors"
	"fmt"
	"github.com/danyatalent/movie-recommend/internal/apperror"
	"github.com/danyatalent/movie-recommend/internal/genre"
	"github.com/danyatalent/movie-recommend/internal/movie"
	"github.com/danyatalent/movie-recommend/pkg/client/postgresql"
	logging "github.com/danyatalent/movie-recommend/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
	"sync"
)

type Repository struct {
	client postgresql.Client
	logger *slog.Logger
}

func (r *Repository) CreateMovie(ctx context.Context, movie *movie.DTO) (string, error) {
	queryMovies := "insert into movies(name, description, duration, rating, director_id) values ($1, $2, $3, $4, $5) returning id"
	r.logger.Info("creating movie", slog.String("query", queryMovies))
	errCh := make(chan error, len(movie.GenresID))

	if err := r.client.QueryRow(ctx, queryMovies, movie.Name, movie.Description,
		movie.Duration, movie.Rating, movie.DirectorID).Scan(&movie.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Code, pgErr.SQLState()))
			r.logger.Error("error due query", logging.Err(newErr))
			return "", newErr
		}
		return "", err
	}
	queryMoviesGenres := "insert into movies_genres(movie_id, genre_id) VALUES ($1, $2)"
	wg := sync.WaitGroup{}
	for _, genreID := range movie.GenresID {
		wg.Add(1)
		go func(genreID, id string) {
			defer wg.Done()
			if _, err := r.client.Exec(ctx, queryMoviesGenres, id, genreID); err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) {
					newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Code: %s, SQLState: %s",
						pgErr.Message, pgErr.Detail, pgErr.Code, pgErr.SQLState()))
					r.logger.Error("error due adding to movies_genres", logging.Err(newErr))
					errCh <- newErr
				} else {
					errCh <- err
				}
			} else {
				errCh <- nil
			}
		}(genreID, movie.ID)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return "", err
		}
	}

	return movie.ID, nil
}

func (r *Repository) GetMovie(ctx context.Context, id string) (movie.Movie, error) {
	queryMovies := "select id, name, description, duration, rating, director_id from movies where id=$1"
	r.logger.Info("getting movie by id")
	var m movie.Movie
	err := r.client.QueryRow(ctx, queryMovies, id).Scan(&m.ID, &m.Name, &m.Description, &m.Duration, &m.Rating, &m.DirectorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return movie.Movie{}, apperror.ErrEntityNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Code, pgErr.SQLState()))
			r.logger.Error("error due getting movie", logging.Err(err))
			return movie.Movie{}, newErr
		}
		return movie.Movie{}, err
	}

	genres, err := r.GetGenresByMovie(ctx, id)
	if err != nil {
		return movie.Movie{}, err
	}
	m.Genres = genres
	return m, nil

}

func (r *Repository) GetGenresByMovie(ctx context.Context, id string) ([]genre.Genre, error) {
	queryGenres := `select g.id, g.name 
					from genres g
					join movies_genres mg on g.id = mg.genre_id
					where mg.movie_id = $1
					`
	genres := make([]genre.Genre, 0)
	rows, err := r.client.Query(ctx, queryGenres, id)
	if err != nil {
		return nil, err
	}
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

func NewRepository(client postgresql.Client, logger *slog.Logger) *Repository {
	return &Repository{
		client: client,
		logger: logger,
	}
}
