package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/cyllective/oauth-labs/lab05/server/internal/dto"
)

type UserRepository struct {
	db *sql.DB
}

var (
	ErrUserNotFound = fmt.Errorf("user not found")
)

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

func (u *UserRepository) Get(ctx context.Context, id string) (*dto.User, error) {
	stmt := `SELECT id, username, password, firstname, lastname, email, extra FROM users WHERE id = ?`
	row := u.db.QueryRowContext(ctx, stmt, id)
	var user dto.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Firstname, &user.Lastname, &user.Email, &user.Extra)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepository) GetByUsername(ctx context.Context, username string) (*dto.User, error) {
	stmt := `SELECT id, username, password, firstname, lastname, email, extra FROM users WHERE username = ?`
	row := u.db.QueryRowContext(ctx, stmt, username)
	var user dto.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Firstname, &user.Lastname, &user.Email, &user.Extra)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepository) Create(ctx context.Context, username string, password string) error {
	stmt := `INSERT INTO users(id, username, password) VALUES(?, ?, ?)`
	_, err := u.db.ExecContext(ctx, stmt, uuid.NewString(), username, password)
	return err
}

func (u *UserRepository) Exists(ctx context.Context, id string) bool {
	row := u.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id = ?`, id)
	cnt := 0
	if err := row.Scan(&cnt); err != nil {
		return false
	}
	return cnt == 1
}

func (u *UserRepository) ExistsUsername(ctx context.Context, username string) bool {
	row := u.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE username = ?`, username)
	cnt := 0
	err := row.Scan(&cnt)
	if err != nil {
		return false
	}
	return cnt > 0
}

func (u *UserRepository) Update(ctx context.Context, request *dto.UpdateProfile) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}

	if request.Firstname != "" {
		_, _ = tx.Exec(`UPDATE users SET firstname = ? WHERE id = ?`, request.Firstname, request.UserID)
	}
	if request.Lastname != "" {
		_, _ = tx.Exec(`UPDATE users SET lastname = ? WHERE id = ?`, request.Lastname, request.UserID)
	}
	if request.Email != "" {
		_, _ = tx.Exec(`UPDATE users SET email = ? WHERE id = ?`, request.Email, request.UserID)
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
	}
	return err
}
