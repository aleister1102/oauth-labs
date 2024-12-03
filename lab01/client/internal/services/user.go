package services

import (
	"context"
	"crypto/sha256"
	"database/sql"

	"github.com/google/uuid"

	"github.com/cyllective/oauth-labs/lab01/client/internal/dto"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db}
}

func (u *UserService) makeID(email string) (string, error) {
	id := sha256.Sum256([]byte(email))
	uid, err := uuid.FromBytes(id[:16])
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}

func (u *UserService) Exists(ctx context.Context, id string) bool {
	row := u.db.QueryRowContext(ctx, `SELECT 1 from users WHERE id = ?`, id)
	var exists int
	if err := row.Scan(&exists); err != nil {
		return false
	}
	return exists == 1
}

func (u *UserService) Create(ctx context.Context, profile *dto.Profile) error {
	userID, err := u.makeID(profile.Email)
	if err != nil {
		return err
	}
	if !u.Exists(ctx, userID) {
		stmt := `INSERT INTO users(id, firstname, lastname, email, extra) VALUES(?, ?, ?, ?, ?)`
		_, err := u.db.ExecContext(ctx, stmt, userID, profile.Firstname, profile.Lastname, profile.Email, profile.Extra)
		return err
	}
	return nil
}

func (u *UserService) Get(ctx context.Context, email string) (*dto.Profile, error) {
	userID, err := u.makeID(email)
	if err != nil {
		return nil, err
	}
	stmt := `SELECT id, firstname, lastname, email, extra FROM users WHERE id = ?`
	row := u.db.QueryRowContext(ctx, stmt, userID)
	var profile dto.Profile
	if err := row.Scan(&profile.ID, &profile.Firstname, &profile.Lastname, &profile.Email, &profile.Extra); err != nil {
		return nil, err
	}
	return &profile, nil
}
