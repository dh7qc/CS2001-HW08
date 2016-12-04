// Package nvector is a bidirectional n-vector converter for go
//
// Reference for n-vector can be found here: https://en.wikipedia.org/wiki/N-vector
package nvector

import (
	"encoding/json"
	"errors"
	"fmt"
	"latlong"
	"math"
)

// Convert angle in radians to angle in degrees
func deg(rad float64) float64 { return rad * 180 / math.Pi }

// Convert angle in degrees to angle in radians
func rad(deg float64) float64 { return deg * math.Pi / 180 }

// Coordinate represents a position on earth in the n-vector
// horizontal position representation
type Coordinate struct {
	X, Y, Z float64
}

// Convert an n-vector Coordinate to its corresponding LatLon
func (c *Coordinate) ToLatLong() latlong.Coordinate {
	lat := deg(math.Atan2(c.Z, math.Hypot(c.X, c.Y)))
	lon := deg(math.Atan2(c.Y, c.X))
	return latlong.Coordinate{lat, lon}
}

// Convert a LatLongto its corresponding n-vector Coordinate
func ToCoordinate(l latlong.LatLonger) Coordinate {
	rlat, rlon := rad(l.Lat()), rad(l.Lon())

	return Coordinate{
		X: deg(math.Cos(rlat) * math.Cos(rlon)),
		Y: deg(math.Cos(rlat) * math.Sin(rlon)),
		Z: deg(math.Sin(rlat)),
	}
}

func (c Coordinate) Lat() float64 {
	return c.ToLatLong().Latitude
}

func (c Coordinate) Lon() float64 {
	return c.ToLatLong().Longitude
}

// Unmarshals a Coordinate from JSON.
func (c *Coordinate) UnmarshalJSON(b []byte) error {
	// Try to unmarshal the JSON object
	obj := make(map[string]interface{})
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	// Check the number of fields
	if len(obj) > 3 {
		return errors.New(fmt.Sprintf("Too many fields for: nvector.Coordinate"))
	}
	if len(obj) < 3 {
		return errors.New(fmt.Sprintf("Not enough fields for: nvector.Coordinate"))
	}

	// Check the value for "X" key (if there is one)
	if _, ok := obj["X"]; !ok {
		return errors.New("Missing field: \"X\"")
	}
	if _, ok := obj["X"].(float64); !ok {
		return errors.New("Wrong type for field: \"X\"")
	}

	// Check the value for "Y" key (if there is one)
	if _, ok := obj["Y"]; !ok {
		return errors.New("Missing field: \"Y\"")
	}
	if _, ok := obj["Y"].(float64); !ok {
		return errors.New("Wrong type for field: \"Y\"")
	}

	// Check the value for "Z" key (if there is one)
	if _, ok := obj["Z"]; !ok {
		return errors.New("Missing field: \"Y\"")
	}
	if _, ok := obj["Z"].(float64); !ok {
		return errors.New("Wrong type for field: \"Y\"")
	}

	// Everything's OK! Time to populate the fields.
	c.X = obj["X"].(float64)
	c.Y = obj["Y"].(float64)
	c.Z = obj["Z"].(float64)

	// No error.
	return nil
}
