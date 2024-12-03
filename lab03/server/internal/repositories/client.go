package repositories

import (
	"context"
	"database/sql"

	"github.com/cyllective/oauth-labs/lab03/server/internal/database"
	"github.com/cyllective/oauth-labs/lab03/server/internal/entities"
)

type ClientRepository struct {
	db *sql.DB
}

func NewClientRepository(db *sql.DB) *ClientRepository {
	return &ClientRepository{db}
}

func (r *ClientRepository) Get(ctx context.Context, id string) (*entities.Client, error) {
	var c entities.Client
	row := r.db.QueryRowContext(ctx, `SELECT id, config from clients WHERE id = ?`, id)
	if err := row.Scan(&c.ID, &c.Config); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ClientRepository) GetAll(ctx context.Context, ids ...string) ([]*entities.Client, error) {
	stmt := "SELECT id, config FROM clients WHERE id IN (" + database.MakePlaceholders(len(ids)) + ")"
	rows, err := r.db.QueryContext(ctx, stmt, database.StringSliceToArgs(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients := make([]*entities.Client, 0)
	for rows.Next() {
		var c entities.Client
		if err := rows.Scan(&c.ID, &c.Config); err != nil {
			return nil, err
		}
		clients = append(clients, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *ClientRepository) Exists(ctx context.Context, id string) bool {
	row := r.db.QueryRowContext(ctx, `SELECT 1 FROM clients WHERE id = ?`, id)
	var ok int
	if err := row.Scan(&ok); err != nil {
		return false
	}
	return ok == 1
}

func (r *ClientRepository) Register(ctx context.Context, request *entities.Client) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	_, _ = tx.Exec(`DELETE FROM clients WHERE id = ?`, request.ID)
	_, _ = tx.Exec(`INSERT INTO clients(id, config) VALUES(?, ?);`, request.ID, request.Config)
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
	}
	return err
}
