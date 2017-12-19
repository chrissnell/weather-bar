package noaa

import (
	"math"
)

const (
	earthRadius = 6371 // km
	toRadians   = (math.Pi / 180)
)

// Latitude and Longitude representation
type Point struct {
	Latitude  float64
	Longitude float64
}

// Calculates the distance between two points.
// The haversine formula is not exact because the earth is not a perfect sphere.
// However for most purposes the results are more than adequate.
func (p1 *Point) HaversineDistance(p2 *Point) float64 {
	dLat := (p2.Latitude - p1.Latitude) * toRadians
	dLng := (p2.Longitude - p1.Longitude) * toRadians

	lat1 := toRadians * p1.Latitude
	lat2 := toRadians * p2.Latitude

	a := haversin(dLat) + math.Cos(lat1)*math.Cos(lat2)*haversin(dLng)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c // d
}

func haversin(val float64) float64 {
	return math.Pow(math.Sin(val/2), 2)
}
