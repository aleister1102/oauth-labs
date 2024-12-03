package repositories

import (
	"context"
	"database/sql"
	"log"

	"github.com/cyllective/oauth-labs/lab02/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab02/server/internal/entities"
)

type ConsentRepository struct {
	db *sql.DB
}

func NewConsentRepository(db *sql.DB) *ConsentRepository {
	return &ConsentRepository{db}
}

func (c *ConsentRepository) Create(ctx context.Context, consent *dto.Consent) error {
	stmt := `INSERT INTO user_consents(client_id, user_id) VALUES(?, ?)`
	_, err := c.db.ExecContext(ctx, stmt, consent.ClientID, consent.UserID)
	return err
}

func (c *ConsentRepository) Delete(ctx context.Context, consent *dto.Consent) error {
	stmt := `DELETE FROM user_consents WHERE client_id = ? AND user_id = ?`
	_, err := c.db.ExecContext(ctx, stmt, consent.ClientID, consent.UserID)
	return err
}

func (c *ConsentRepository) Get(ctx context.Context, consent *dto.Consent) (*dto.Consent, error) {
	stmt := `SELECT client_id, user_id FROM user_consents WHERE client_id = ? AND user_id = ?`
	row := c.db.QueryRowContext(ctx, stmt, consent.ClientID, consent.UserID)
	var cons entities.Consent
	if err := row.Scan(&cons.ClientID, &cons.UserID); err != nil {
		log.Printf("[ConsentService.Get]: error %#v\n", err.Error())
		return nil, err
	}
	return &dto.Consent{ClientID: cons.ClientID, UserID: cons.UserID}, nil
}

func (c *ConsentRepository) GetAll(ctx context.Context, userID string) (*dto.UserConsents, error) {
	stmt := `SELECT client_id FROM user_consents WHERE user_id = ?`
	rows, err := c.db.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}
	clientIDs := make([]string, 0)
	for rows.Next() {
		var clientID string
		if err := rows.Scan(&clientID); err != nil {
			return nil, err
		}
		clientIDs = append(clientIDs, clientID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &dto.UserConsents{
		UserID:    userID,
		ClientIDs: clientIDs,
	}, nil
}

func (c *ConsentRepository) Exists(ctx context.Context, consent *dto.Consent) bool {
	stmt := `SELECT client_id, user_id FROM user_consents WHERE client_id = ? AND user_id = ?`
	row := c.db.QueryRowContext(ctx, stmt, consent.ClientID, consent.UserID)
	var cons entities.Consent
	if err := row.Scan(&cons.ClientID, &cons.UserID); err != nil {
		return false
	}
	return cons.ClientID == consent.ClientID && cons.UserID == consent.UserID
}
