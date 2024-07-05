package server

import (
	"math"
	"sort"

	"github.com/chammond14/muzz/internal/db"
)

func sortProfilesByLocation(profiles []*db.DiscoverProfile, location db.Location) {
	for _, profile := range profiles {
		profile.DistanceFromMe = getDistanceInKm(location.Lat, location.Long, profile.Lat, profile.Long)
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].DistanceFromMe < profiles[j].DistanceFromMe
	})
}

func getDistanceInKm(lat1 float64, lon1 float64, lat2 float64, lon2 float64) int {
	radLat1 := float64(math.Pi * lat1 / 180)
	radLat2 := float64(math.Pi * lat2 / 180)

	theta := float64(lon1 - lon2)
	radTheta := float64(math.Pi * theta / 180)

	dist := math.Sin(radLat1)*math.Sin(radLat2) + math.Cos(radLat1)*math.Cos(radLat2)*math.Cos(radTheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515
	dist = dist * 1.609344

	return int(dist)
}
