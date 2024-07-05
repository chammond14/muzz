package db

import (
	"context"
	"log/slog"
	"slices"

	"github.com/jackc/pgx"
)

type SwipeDetails struct {
	Id          int32
	SwipedOn    []int32
	SwipedYesBy []int32
}

func scanSwipeRows(r *pgx.Rows) (SwipeDetails, error) {
	profile := &SwipeDetails{}
	err := r.Scan(
		&profile.Id,
		&profile.SwipedOn,
		&profile.SwipedYesBy,
	)

	return *profile, err
}

func (ps *PostgresStore) Swipe(ctx context.Context, userId int32, swipedUserId int32, liked bool) (bool, int, error) {
	slog.Info("Swiping profile", "swiper", userId, "swiped user", swipedUserId, "liked", liked)

	ctx, cancel := context.WithTimeoutCause(ctx, getTimeoutDuration(), ErrQueryTimedOut)
	defer cancel()

	profiles, err := ps.getSwipeProfiles(ctx, userId, swipedUserId)
	if err != nil {
		return false, 0, err
	}

	swiperProfile := profiles[slices.IndexFunc(profiles, func(p *SwipeDetails) bool { return p.Id == userId })]
	swipedProfile := profiles[slices.IndexFunc(profiles, func(p *SwipeDetails) bool { return p.Id == swipedUserId })]

	swiperProfile.SwipedOn = append(swiperProfile.SwipedOn, swipedUserId)

	isMatch := liked && slices.Contains(swiperProfile.SwipedYesBy, swipedUserId)

	tx, err := ps.PostgresConnection.Begin()
	if err != nil {
		slog.Info("Error beginning transaction", "error", err)
		return false, 0, ErrDatabaseError
	}

	defer tx.RollbackEx(ctx)

	if isMatch {
		slog.Info("Swiped yes and matched, updating profiles")

		index := slices.Index(swiperProfile.SwipedYesBy, swipedUserId)
		slices.Delete(swiperProfile.SwipedYesBy, index, index+1)

		createMatchQuery := `INSERT INTO matches (user1Id, user2Id) VALUES ($1, $2) RETURNING id`
		updateSwiperQuery := `UPDATE profiles SET swipedYesBy = $1, swipedOn = $2 WHERE id = $3`

		tx.ExecEx(ctx, updateSwiperQuery, nil, swiperProfile.SwipedYesBy, swiperProfile.SwipedOn, userId)
		slog.Info("Updating swiper profile complete")

		row := tx.QueryRowEx(ctx, createMatchQuery, nil, userId, swipedUserId)
		slog.Info("Updating swiped profile complete")

		match := &Match{}
		err = match.scanRow(row)
		if err != nil {
			slog.Error("Error creating match", "error", err)

			return false, 0, ErrDatabaseError
		}

		tx.CommitEx(ctx)
		slog.Info("Swiping profile complete")
		return true, match.Id, nil
	} else if liked {
		slog.Info("Swiped yes but no match, updating profiles")

		swipedProfile.SwipedYesBy = append(swipedProfile.SwipedYesBy, userId)
		updateSwiperQuery := `UPDATE profiles SET swipedOn = $1 WHERE id = $2`
		updateSwipedQuery := `UPDATE profiles SET swipedYesBy = $1 WHERE id = $2`

		tx.ExecEx(ctx, updateSwiperQuery, nil, swiperProfile.SwipedOn, userId)
		slog.Info("Updating swiper profile complete")

		tx.ExecEx(ctx, updateSwipedQuery, nil, swipedProfile.SwipedYesBy, swipedUserId)
		slog.Info("Updating swiped profile complete")

		tx.CommitEx(ctx)
		slog.Info("Swiping profile complete")
		return false, 0, nil
	}

	updateSwiperQuery := `UPDATE profiles SET swipedOn = $1 WHERE id = $2`
	tx.ExecEx(ctx, updateSwiperQuery, nil, swiperProfile.SwipedOn, userId)
	tx.CommitEx(ctx)
	return false, 0, nil
}

func (ps *PostgresStore) getSwipeProfiles(ctx context.Context, userId1 int32, userId2 int32) ([]*SwipeDetails, error) {
	slog.Info("Getting profiles for swipe", "swiper", userId1, "swiped user", userId2)

	usersQuery := `SELECT id, swipedOn, swipedYesBy FROM profiles WHERE id in ($1, $2)`
	rows, err := ps.PostgresConnection.QueryEx(ctx, usersQuery, nil, userId1, userId2)
	if err != nil {
		slog.Error("Error retrieving swiped profiles", "error", err)
		return nil, ErrDatabaseError
	}

	var profiles []*SwipeDetails
	for rows.Next() {
		profile, err := scanSwipeRows(rows)
		if err != nil {
			slog.Error("Error scanning rows", "method", "getSwipeProfiles", "error", err)
			return nil, ErrDatabaseError
		}

		profiles = append(profiles, &profile)
	}

	if len(profiles) < 2 {
		return nil, ErrSwipeRequestInvalid
	}

	return profiles, nil
}
