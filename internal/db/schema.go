package db

import (
	"log/slog"
	"os"

	"github.com/jackc/pgx"
)

const schema = `
	CREATE TABLE IF NOT EXISTS profiles (
		id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		age INTEGER NOT NULL,
		name TEXT NOT NULL,
		gender TEXT NOT NULL,
		swipedOn integer[],
		swipedYesBy integer[],
		lat float,
		long float,
		createdAt timestamp not null default current_timestamp
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT NOT NULL,
		userId INTEGER NOT NULL PRIMARY KEY REFERENCES profiles (id),
		expiresAt timestamp not null default current_timestamp + (20 * interval '1 minute')
	);

	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		user1Id INTEGER NOT NULL REFERENCES profiles (id),
		user2Id INTEGER NOT NULL REFERENCES profiles (id),
		matchedAt timestamp not null default current_timestamp
	);`

// applySchema applies the schema to a connection and seeds the database
func applySchema(postgresConnection *pgx.ConnPool) error {
	tx, err := postgresConnection.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.Exec(schema); err != nil {
		return err
	}

	isLocal := os.Getenv("isLocal") == "true"
	if isLocal {
		seedDatabase(tx)
	}

	return tx.Commit()
}

func seedDatabase(tx *pgx.Tx) {
	slog.Info("Seeding database")

	tx.Exec(seed1)
	tx.Exec(seed2)
	tx.Exec(seed3)
	tx.Exec(seed4)
}
