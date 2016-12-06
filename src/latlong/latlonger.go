package latlong

import (
	"math"
)

// A LatLonger can return its position on earth in terms of latitude and longitude
type LatLonger interface {
	Lat() float64
	Lon() float64
}

// Computes hsin of angle theta in radians
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance (in miles) between two LatLongers: a and b
func Distance(a, b LatLonger) float64 {
	latA, lonA := rad(a.Lat()), rad(a.Lon())
	latB, lonB := rad(b.Lat()), rad(b.Lon())

	r := 3958.76 // Earth's radius in miles
	h := hsin(latB-latA) + math.Cos(latA)*math.Cos(latB)*hsin(lonB-lonA)

	return 2 * r * math.Asin(math.Sqrt(h))
}
