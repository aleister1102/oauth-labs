package repositories

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/cyllective/oauth-labs/lab00/server/internal/dto"
)

type AccessTokenRepository struct {
	db *sql.DB
}

func NewAcccessTokenRepository(db *sql.DB) *AccessTokenRepository {
	return &AccessTokenRepository{db}
}

func (a *AccessTokenRepository) Exists(ctx context.Context, id string) bool {
	row := a.db.QueryRowContext(ctx, `SELECT 1 FROM access_tokens WHERE id = ?`, id)
	var ok int
	if err := row.Scan(&ok); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		panic(err)
	}
	return ok == 1
}

func (a *AccessTokenRepository) Create(ctx context.Context, request *dto.CreateAccessToken) error {
	stmt := `INSERT INTO access_tokens(id, user_id, client_id, data) VALUES(?, ?, ?, ?)`
	_, err := a.db.ExecContext(ctx, stmt, request.ID, request.UserID, request.ClientID, request.EncryptedJWT)
	return err
}

func (a *AccessTokenRepository) Delete(ctx context.Context, id string) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM access_tokens WHERE id = ?`, id)
	if err != nil {
		log.Printf("[AccessTokenRepository.Delete]: failed to delete access token: %s", err)
	}
	return err
}

func (a *AccessTokenRepository) DeleteAll(ctx context.Context, request *dto.DeleteAccessTokens) error {
	_, err := a.db.ExecContext(ctx, `DELETE FROM access_tokens WHERE client_id = ? AND user_id = ?`, request.ClientID, request.UserID)
	return err
}

func (a *AccessTokenRepository) RevokeAll(ctx context.Context, request *dto.RevokeAccessTokens) error {
	_, err := a.db.ExecContext(ctx, `UPDATE access_tokens SET revoked = true WHERE client_id = ? AND user_id = ?`, request.ClientID, request.UserID)
	return err
}

func (a *AccessTokenRepository) IsRevoked(ctx context.Context, id string) bool {
	row := a.db.QueryRowContext(ctx, `SELECT revoked FROM access_tokens WHERE id = ?`, id)
	var revoked bool
	if err := row.Scan(&revoked); err != nil {
		// On any error, we assume it is revoked.
		return true
	}
	return revoked
}
