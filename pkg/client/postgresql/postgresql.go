package postgresql

import (
	"context"
	"fmt"
	"github.com/danyatalent/movie-recommend/internal/config"
	logging "github.com/danyatalent/movie-recommend/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func NewClient(log *slog.Logger, ctx context.Context, maxAttempts int, sc config.Storage) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", sc.Username, sc.Password, sc.Host, sc.Port, sc.Database)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Error("error connecting postgresql", logging.Err(err))
		os.Exit(1)
	}

	log.Info("successfully connected to db",
		slog.String("db", sc.Database),
		slog.String("username", sc.Username),
		slog.String("host", sc.Host),
		slog.String("port", sc.Port),
	)

	return pool, nil
}

//func DoWithTries(fn func() error, attempts int, delay time.Duration) (err error) {
//	for attempts > 0 {
//		if err = fn(); err != nil {
//			time.Sleep(delay)
//			attempts--
//
//			continue
//		}
//
//		return nil
//	}
//
//	return
//}
