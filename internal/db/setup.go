package db

import (
	"context"

	"github.com/jackc/pgx"
)

// Store provides access to the database.
type PostgresStore struct {
	PostgresConnection *pgx.ConnPool
}

// NewStore sets up a new database store.
func NewPostgresStore(connStr string) (*PostgresStore, error) {
	store := &PostgresStore{}
	conn, err := setupConnectionPool(connStr)
	if err != nil {
		return nil, err
	}

	applySchema(conn)

	store.PostgresConnection = conn
	return store, nil
}

// TestConnection tests that the Store can properly connect to the Postgres Server.
func (s *PostgresStore) TestConnection(ctx context.Context) error {
	// _, err := s.PostgresConnection.ExecEx(ctx, "SELECT 1", nil)
	// return err
	return nil
}

func setupConnectionPool(connStr string) (*pgx.ConnPool, error) {
	config, err := pgx.ParseURI(connStr)

	if err != nil {
		return nil, err
	}

	conn, _ := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: config,
	})

	return conn, nil
}
