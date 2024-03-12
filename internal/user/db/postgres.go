package user

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/danyatalent/movie-recommend/internal/apperror"
	"github.com/danyatalent/movie-recommend/internal/user"
	"github.com/danyatalent/movie-recommend/pkg/client/postgresql"
	logging "github.com/danyatalent/movie-recommend/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
)

type Repository struct {
	client postgresql.Client
	logger *slog.Logger
}

func (r *Repository) UpdateUserPassword(ctx context.Context, id, newPass string) error {
	q := "update users set password=$2 where id=$1"
	r.logger.Info("update user password", slog.String("id", id))
	newPasswordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(newPass)))
	result, err := r.client.Exec(ctx, q, id, newPasswordHash)
	if result.RowsAffected() == 0 {
		return apperror.ErrEntityNotFound
	}
	if err != nil {
		return fmt.Errorf("can't update user pass: %s", id)
	}
	return nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (user.User, error) {
	q := "select id, name, email from users where id=$1"
	r.logger.Debug("getting user from users", slog.String("id", id))
	var u user.User
	err := r.client.QueryRow(ctx, q, id).Scan(&u.ID, &u.Name, &u.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, apperror.ErrEntityNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Code, pgErr.SQLState()))
			r.logger.Error("error due query", logging.Err(newErr))
			return user.User{}, newErr
		}
		return user.User{}, err
	}
	return u, nil
}

//  PUT /users/{id} - {name: ... pass: ...} {name: ...} {pass: ...}

func (r *Repository) UpdateUserName(ctx context.Context, id, newName string) error {
	q := "update users set name=$2 where id=$1"
	r.logger.Info("update username", slog.String("id", id))
	result, err := r.client.Exec(ctx, q, id, newName)
	if result.RowsAffected() == 0 {
		return apperror.ErrEntityNotFound
	}
	if err != nil {
		return fmt.Errorf("can't update user name: %s", id)
	}
	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	q := "delete from users where id=$1"
	r.logger.Debug("deleting user")
	result, err := r.client.Exec(ctx, q, id)
	if result.RowsAffected() == 1 {
		return apperror.ErrEntityNotFound
	}
	if err != nil {
		return fmt.Errorf("can't delete user: %s", id)
	}
	return nil
}

func (r *Repository) CreateUser(ctx context.Context, user *user.User) (string, error) {
	//passwordHash := sha256.New()
	//passwordHash.Write([]byte(user.Password))
	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(user.Password)))
	r.logger.Info("password hash", slog.String("passwordHash", passwordHash))

	q := "insert into users(name, password, email) values ($1, $2, $3) returning id"
	r.logger.Info("creating user", slog.String("query", q))
	if err := r.client.QueryRow(ctx, q, user.Name, passwordHash, user.Email).Scan(&user.ID); err != nil {
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

	return user.ID, nil
}

func NewRepository(client postgresql.Client, logger *slog.Logger) *Repository {
	return &Repository{
		client: client,
		logger: logger,
	}
}
