// Package latlong contains types and functions for working with
// Latitude/Longitude coordinates
//
// Reference for latitude and longitude can be found here:
//     - https://en.wikipedia.org/wiki/Latitude
//     - https://en.wikipedia.org/wiki/Longitude
package latlong

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

// Convert angle in radians to angle in degrees
func deg(rad float64) float64 { return rad * 180 / math.Pi }

// Convert angle in degrees to angle in radians
func rad(deg float64) float64 { return deg * math.Pi / 180 }

// Coordinate represents a position on earth by latitude and longitude
type Coordinate struct {
	Latitude  float64
	Longitude float64
}

func (c Coordinate) Lat() float64 {
	return c.Latitude
}

func (c Coordinate) Lon() float64 {
	return c.Longitude
}

// Unmarshals a Coordinate from JSON.
func (c *Coordinate) UnmarshalJSON(b []byte) error {
	// Try to unmarshal the JSON object
	obj := make(map[string]interface{})
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	// Check the number of fields
	if len(obj) > 2 {
		return errors.New(fmt.Sprintf("Too many fields for: latlong.Coordinate"))
	}
	if len(obj) < 2 {
		return errors.New(fmt.Sprintf("Not enough fields for: latlong.Coordinate"))
	}

	// Check the value for "Latitude" key (if there is one)
	if _, ok := obj["Latitude"]; !ok {
		return errors.New("Missing field: \"Latitude\"")
	}
	if _, ok := obj["Latitude"].(float64); !ok {
		return errors.New("Wrong type for field: \"Latitude\"")
	}

	// Check the value for "Longitude" key (if there is one)
	if _, ok := obj["Longitude"]; !ok {
		return errors.New("Missing field: \"Longitude\"")
	}
	if _, ok := obj["Longitude"].(float64); !ok {
		return errors.New("Wrong type for field: \"Longitude\"")
	}

	// Everything's OK! Time to populate the fields.
	c.Latitude = obj["Latitude"].(float64)
	c.Longitude = obj["Longitude"].(float64)

	// No error.
	return nil
}
