package server

import (
	"testing"

	"github.com/chammond14/muzz/internal/db"
)

func Test_sortProfilesByLocationSortsLocationsCorrectly(t *testing.T) {
	profile1 := &db.DiscoverProfile{
		Id:             1,
		Age:            50,
		Gender:         "Female",
		Name:           "Tester",
		DistanceFromMe: 100,
		Lat:            -0.11406225575119984,
		Long:           51.55781040589032,
	}

	profile2 := &db.DiscoverProfile{
		Id:             2,
		Age:            100,
		Gender:         "Female",
		Name:           "Bob",
		DistanceFromMe: 20,
		Lat:            -0.08768348444653988,
		Long:           51.508050972200834,
	}

	profiles := []*db.DiscoverProfile{profile1, profile2}

	sortProfilesByLocation(profiles, db.Location{Lat: -0.08768348444653988, Long: 51.508050972200834})

	if profiles[0].DistanceFromMe > profiles[1].DistanceFromMe {
		t.Errorf("expected first profile in slice to have lesser distance")
	}
}
