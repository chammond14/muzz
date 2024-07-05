package db

import "errors"

var (
	ErrQueryTimedOut       = errors.New("query timed out")
	ErrDatabaseError       = errors.New("could not access data store")
	ErrSwipeRequestInvalid = errors.New("failed to swipe on profile")
	ErrNoValidSession      = errors.New("no valid session")
	ErrLoginFailed         = errors.New("could not log in")
)
