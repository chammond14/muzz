package db

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx"
)

// Checkout describes a checkout as stored in the db.
type Profile struct {
	Id       int32
	Age      int
	Name     string
	Gender   string
	Email    string
	Password string
	Location Location
}

type Location struct {
	Lat  float64
	Long float64
}

func (p *Profile) scanRow(r *pgx.Row) error {
	return r.Scan(
		&p.Id,
		&p.Age,
		&p.Name,
		&p.Gender,
		&p.Email,
		&p.Password,
		&p.Location.Lat,
		&p.Location.Long,
	)
}

type Session struct {
	Token     string
	UserId    int32
	Timestamp time.Time
}

func (s *Session) scanRow(r *pgx.Row) error {
	return r.Scan(
		&s.Token,
		&s.UserId,
		&s.Timestamp,
	)
}

type Match struct {
	Id int
}

func (m *Match) scanRow(r *pgx.Row) error {
	return r.Scan(
		&m.Id,
	)
}

// ProfileStore describes an interface which any data store must implement to achieve required functionality
type ProfileStore interface {
	CreateProfile(context.Context, int, string, string, string, string, Location) (*Profile, error)
	GetDiscoverProfiles(context.Context, int32, DiscoverFilters) ([]*DiscoverProfile, error)
	GetSession(context.Context, string) (int32, error)
	Login(context.Context, string, string) (string, error)
	Swipe(context.Context, int32, int32, bool) (bool, int, error)
}

func (ps *PostgresStore) CreateProfile(ctx context.Context, age int, name string, gender string, email string, password string, location Location) (*Profile, error) {
	slog.Info("Creating profile")

	ctx, cancel := context.WithTimeoutCause(ctx, getTimeoutDuration(), ErrQueryTimedOut)
	defer cancel()

	query := `INSERT INTO profiles (age, name, gender, email, password, lat, long) 
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id, age, name, gender, email, password, lat, long`

	row := ps.PostgresConnection.QueryRowEx(ctx, query, nil, age, name, gender, email, password, location.Lat, location.Long)
	profile := &Profile{}
	err := profile.scanRow(row)
	if err != nil {
		slog.Error("Error creating profile", "error", err)
		return nil, ErrDatabaseError
	}

	slog.Info("Creating profile complete")
	return profile, nil
}

func (ps *PostgresStore) Login(ctx context.Context, email string, password string) (string, error) {
	slog.Info("Logging in")

	ctx, cancel := context.WithTimeoutCause(ctx, getTimeoutDuration(), ErrQueryTimedOut)
	defer cancel()

	query := `SELECT id, age, name, gender, email, password, lat, long FROM profiles WHERE email = $1 AND password = $2`
	row := ps.PostgresConnection.QueryRowEx(ctx, query, nil, email, password)

	profile := &Profile{}
	err := profile.scanRow(row)
	if err != nil {
		slog.Error("Error logging in", "error", err)
		if err == pgx.ErrNoRows {
			return "", ErrLoginFailed
		}

		return "", ErrDatabaseError
	}

	sessionToken := uuid.New().String()

	query = `INSERT INTO sessions (token, userId) VALUES ($1, $2)
	ON CONFLICT (userId) DO UPDATE SET token = $1, expiresAt = DEFAULT`

	_, err = ps.PostgresConnection.ExecEx(ctx, query, nil, sessionToken, profile.Id)
	if err != nil {
		slog.Error("Error creating session", "error", err)
		return "", ErrDatabaseError
	}

	return sessionToken, nil
}

func (ps *PostgresStore) GetSession(ctx context.Context, token string) (int32, error) {
	slog.Info("Getting session")

	ctx, cancel := context.WithTimeoutCause(ctx, getTimeoutDuration(), ErrQueryTimedOut)
	defer cancel()

	query := `SELECT token, userId, expiresAt FROM sessions WHERE token = $1 AND expiresAt > now()`

	row := ps.PostgresConnection.QueryRowEx(ctx, query, nil, token)
	session := &Session{}
	err := session.scanRow(row)
	slog.Info("Time", "stamp", session.Timestamp)
	if err != nil {
		slog.Error("Error finding session", "error", err)
		return 0, ErrNoValidSession
	}

	slog.Info("Getting session complete")
	return session.UserId, nil
}

func getTimeoutDuration() time.Duration {
	t, err := time.ParseDuration(os.Getenv("DB_TIMEOUT_SECONDS"))
	if err != nil {
		slog.Info("Could not load DB_TIMEOUT_SECONDS variable")
		return time.Duration(time.Second * 10)
	}

	return t
}
