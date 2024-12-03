package repositories

import (
	"context"
	"database/sql"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab01/server/internal/entities"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db}
}

func (r *RefreshTokenRepository) Get(ctx context.Context, id string) (*entities.RefreshToken, error) {
	stmt := `SELECT id, client_id, user_id, revoked FROM refresh_tokens WHERE id = ?`
	row := r.db.QueryRowContext(ctx, stmt, id)
	var t entities.RefreshToken
	if err := row.Scan(&t.ID, &t.ClientID, &t.UserID, &t.Revoked); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	stmt := `INSERT INTO refresh_tokens(id, client_id, user_id, data, revoked) VALUES(?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, stmt, token.ID, token.ClientID, token.UserID, token.Data, token.Revoked)
	return err
}

func (r *RefreshTokenRepository) RevokeAll(ctx context.Context, request *dto.RevokeRefreshTokens) error {
	stmt := `UPDATE refresh_tokens SET revoked = true WHERE client_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, stmt, request.ClientID, request.UserID)
	return err
}
