package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx"
)

type DiscoverProfile struct {
	Id             int32   `json:"id"`
	Age            int     `json:"age"`
	Name           string  `json:"name"`
	Gender         string  `json:"gender"`
	DistanceFromMe int     `json:"distanceFromMe"`
	Lat            float64 `json:"-"`
	Long           float64 `json:"-"`
}

type DiscoverFilters struct {
	MinAge  int
	MaxAge  int
	Genders []string
}

func scanDiscoverRows(r *pgx.Rows) (*DiscoverProfile, error) {
	profile := &DiscoverProfile{}
	err := r.Scan(
		&profile.Id,
		&profile.Age,
		&profile.Name,
		&profile.Gender,
		&profile.Lat,
		&profile.Long,
	)

	return profile, err
}

func (ps *PostgresStore) GetDiscoverProfiles(ctx context.Context, id int32, filters DiscoverFilters) ([]*DiscoverProfile, error) {
	slog.Info("Getting profiles")

	ctx, cancel := context.WithTimeoutCause(ctx, getTimeoutDuration(), ErrQueryTimedOut)
	defer cancel()

	// hack to return any gender if none supplied in filter as couldn't get query working properly in time
	if len(filters.Genders) == 0 {
		filters.Genders = []string{"male", "female", "other"}
	}
	query := `SELECT id, age, name, gender, lat, long FROM profiles 
				WHERE id NOT IN (SELECT unnest(swipedOn) FROM profiles WHERE id = $1)
				AND ($2 = 0 OR age <= $2)
				AND ($3 = 0 OR age >= $3)
				AND gender = ANY ($4)`

	slog.Info("Discover query", "q", query)
	rows, err := ps.PostgresConnection.QueryEx(ctx, query, nil, id, filters.MaxAge, filters.MinAge, filters.Genders)
	if err != nil {
		slog.Error("Error retrieving discover profiles", "error", err)
		return nil, ErrDatabaseError
	}

	var profiles []*DiscoverProfile
	for rows.Next() {
		profile, err := scanDiscoverRows(rows)
		if err != nil {
			slog.Error("Error scanning rows", "method", "getDiscoverProfiles", "error", err)
			return nil, ErrDatabaseError
		}

		profiles = append(profiles, profile)
	}

	slog.Info("Getting discover profiles complete", "len", len(profiles))
	return profiles, nil
}
