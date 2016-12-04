// Copyright (c) 2015 Petr Lozhkin (im7mortal@gmail.com)
//
// Permission is hereby granted, free of charge, to any person
// obtaining a copy of this software and associated documentation
// files (the "Software"), to deal in the Software without
// restriction, including without limitation the rights to use, copy,
// modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
// BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
// ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// The original source for this package can be found here:
// https://github.com/im7mortal/UTM
//
// This file is a modified version of https://github.com/im7mortal/UTM/blob/279c65143efbea68f1d402f3d72943287f0ab8c8/utm.go
//
// Package UTM is bidirectional UTM-WGS84 converter for go
package utm

import (
	"encoding/json"
	"errors"
	"fmt"
	"latlong"
	"math"
	"os"
	"unicode"
)

const (
	k0 float64 = 0.9996
	e  float64 = 0.00669438
	r          = 6378137
)

var e2 = e * e
var e3 = e2 * e
var e_p2 = e / (1.0 - e)

var sqrt_e = math.Sqrt(1 - e)

var _e = (1 - sqrt_e) / (1 + sqrt_e)
var _e2 = _e * _e
var _e3 = _e2 * _e
var _e4 = _e3 * _e
var _e5 = _e4 * _e

var m1 = (1 - e/4 - 3*e2/64 - 5*e3/256)
var m2 = (3*e/8 + 3*e2/32 + 45*e3/1024)
var m3 = (15*e2/256 + 45*e3/1024)
var m4 = (35 * e3 / 3072)

var p2 = (3./2*_e - 27./32*_e3 + 269./512*_e5)
var p3 = (21./16*_e2 - 55./32*_e4)
var p4 = (151./96*_e3 - 417./128*_e5)
var p5 = (1097. / 512 * _e4)

type zone_letter struct {
	zone   int
	letter string
}

const x = math.Pi / 180

func rad(d float64) float64 { return d * x }
func deg(r float64) float64 { return r / x }

var zone_letters = []zone_letter{
	{84, " "},
	{72, "X"},
	{64, "W"},
	{56, "V"},
	{48, "U"},
	{40, "T"},
	{32, "S"},
	{24, "R"},
	{16, "Q"},
	{8, "P"},
	{0, "N"},
	{-8, "M"},
	{-16, "L"},
	{-24, "K"},
	{-32, "J"},
	{-40, "H"},
	{-48, "G"},
	{-56, "F"},
	{-64, "E"},
	{-72, "D"},
	{-80, "C"},
}

// Coordinate contains coordinates in the Universal Transverse
// Mercator coordinate system
type Coordinate struct {
	Easting    float64
	Northing   float64
	ZoneNumber int
	ZoneLetter string
}

// ToLatLong converts Universal Transverse Mercator (UTM) coordinates to a latitude and longitude
func (coordinate *Coordinate) ToLatLong() (latlong.Coordinate, error) {
	zoneLetterExist := !(coordinate.ZoneLetter == "")

	if !zoneLetterExist {
		return latlong.Coordinate{}, errors.New("ZoneLetter field needs to be set")
	}

	if !(100000 <= coordinate.Easting && coordinate.Easting < 1000000) {
		err := errors.New("easting out of range (must be between 100.000 m and 999.999 m")
		return latlong.Coordinate{}, err
	}
	if !(0 <= coordinate.Northing && coordinate.Northing <= 10000000) {
		err := errors.New("northing out of range (must be between 0 m and 10.000.000 m)")
		return latlong.Coordinate{}, err
	}
	if !(1 <= coordinate.ZoneNumber && coordinate.ZoneNumber <= 60) {
		err := errors.New("zone number out of range (must be between 1 and 60)")
		return latlong.Coordinate{}, err
	}

	zoneLetter := unicode.ToUpper(rune(coordinate.ZoneLetter[0]))
	if !('C' <= zoneLetter && zoneLetter <= 'X') || zoneLetter == 'I' || zoneLetter == 'O' {
		err := errors.New("zone letter out of range (must be between C and X)")
		return latlong.Coordinate{}, err
	}

	northernValue := (zoneLetter >= 'N')
	x := coordinate.Easting - 500000
	y := coordinate.Northing

	if !northernValue {
		y -= 10000000
	}

	m := y / k0
	mu := m / (r * m1)

	p_rad := (mu +
		p2*math.Sin(2*mu) +
		p3*math.Sin(4*mu) +
		p4*math.Sin(6*mu) +
		p5*math.Sin(8*mu))

	p_sin := math.Sin(p_rad)
	p_sin2 := p_sin * p_sin

	p_cos := math.Cos(p_rad)

	p_tan := p_sin / p_cos
	p_tan2 := p_tan * p_tan
	p_tan4 := p_tan2 * p_tan2

	ep_sin := 1 - e*p_sin2
	ep_sin_sqrt := math.Sqrt(1 - e*p_sin2)

	n := r / ep_sin_sqrt
	rad := (1 - e) / ep_sin

	c := _e * p_cos * p_cos
	c2 := c * c

	d := x / (n * k0)
	d2 := d * d
	d3 := d2 * d
	d4 := d3 * d
	d5 := d4 * d
	d6 := d5 * d

	latitude := (p_rad - (p_tan/rad)*
		(d2/2-
			d4/24*(5+3*p_tan2+10*c-4*c2-9*e_p2)) +
		d6/720*(61+90*p_tan2+298*c+45*p_tan4-252*e_p2-3*c2))

	longitude := (d -
		d3/6*(1+2*p_tan2+c) +
		d5/120*(5-2*c+28*p_tan2-3*c2+8*e_p2+24*p_tan4)) / p_cos

	return latlong.Coordinate{deg(latitude), deg(longitude) + float64(zone_number_to_central_longitude(coordinate.ZoneNumber))}, nil

}

// ToCoordinate converts a LatLonger to Universal Transverse Mercator coordinates
func ToCoordinate(point latlong.LatLonger) (coord Coordinate, err error) {
	if !(-80.0 <= point.Lat() && point.Lat() <= 84.0) {
		err = errors.New("latitude out of range (must be between 80 deg S and 84 deg N)")
		return
	}
	if !(-180.0 <= point.Lon() && point.Lon() <= 180.0) {
		err = errors.New("longitude out of range (must be between 180 deg W and 180 deg E)")
		return
	}

	lat_rad := rad(point.Lat())
	lat_sin := math.Sin(lat_rad)
	lat_cos := math.Cos(lat_rad)

	lat_tan := lat_sin / lat_cos
	lat_tan2 := lat_tan * lat_tan
	lat_tan4 := lat_tan2 * lat_tan2

	coord.ZoneNumber = latlon_to_zone_number(point.Lat(), point.Lon())

	coord.ZoneLetter = latitude_to_zone_letter(point.Lat())

	lon_rad := rad(point.Lon())
	central_lon := zone_number_to_central_longitude(coord.ZoneNumber)
	central_lon_rad := rad(float64(central_lon))

	n := r / math.Sqrt(1-e*lat_sin*lat_sin)
	c := e_p2 * lat_cos * lat_cos

	a := lat_cos * (lon_rad - central_lon_rad)
	a2 := a * a
	a3 := a2 * a
	a4 := a3 * a
	a5 := a4 * a
	a6 := a5 * a
	m := r * (m1*lat_rad -
		m2*math.Sin(2*lat_rad) +
		m3*math.Sin(4*lat_rad) -
		m4*math.Sin(6*lat_rad))
	coord.Easting = k0*n*(a+
		a3/6*(1-lat_tan2+c)+
		a5/120*(5-18*lat_tan2+lat_tan4+72*c-58*e_p2)) + 500000
	coord.Northing = k0 * (m + n*lat_tan*(a2/2+
		a4/24*(5-lat_tan2+9*c+4*c*c)+
		a6/720*(61-58*lat_tan2+lat_tan4+600*c-330*e_p2)))

	if point.Lat() < 0 {
		coord.Northing += 10000000
	}

	return
}

func latitude_to_zone_letter(latitude float64) string {
	for _, zone_letter := range zone_letters {
		if latitude >= float64(zone_letter.zone) {
			return zone_letter.letter
		}
	}
	return " "
}

func latlon_to_zone_number(latitude float64, longitude float64) int {
	if 56 <= latitude && latitude <= 64 && 3 <= longitude && longitude <= 12 {
		return 32
	}

	if 72 <= latitude && latitude <= 84 && longitude >= 0 {
		if longitude <= 9 {
			return 31
		} else if longitude <= 21 {
			return 33
		} else if longitude <= 33 {
			return 35
		} else if longitude <= 42 {
			return 37
		}
	}

	return int((longitude+180)/6) + 1
}

func zone_number_to_central_longitude(zone_number int) int {
	return (zone_number-1)*6 - 180 + 3
}

func (c Coordinate) Lat() float64 {
	coord, err := c.ToLatLong()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return coord.Latitude
}

func (c Coordinate) Lon() float64 {
	coord, err := c.ToLatLong()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return coord.Longitude
}

// Unmarshals a Coordinate from JSON.
func (c *Coordinate) UnmarshalJSON(b []byte) error {
	// Try to unmarshal the JSON object
	obj := make(map[string]interface{})
	if err := json.Unmarshal(b, &obj); err != nil {
		return err
	}

	// Check the number of fields
	if len(obj) > 4 {
		return errors.New(fmt.Sprintf("Too many fields for: utm.Coordinate"))
	}
	if len(obj) < 4 {
		return errors.New(fmt.Sprintf("Not enough fields for: utm.Coordinate"))
	}

	// Check the value for "Easting" key (if there is one)
	if _, ok := obj["Easting"]; !ok {
		return errors.New("Missing field: \"Easting\"")
	}
	if _, ok := obj["Easting"].(float64); !ok {
		return errors.New("Wrong type for field: \"Easting\"")
	}

	// Check the value for "Northing" key (if there is one)
	if _, ok := obj["Northing"]; !ok {
		return errors.New("Missing field: \"Northing\"")
	}
	if _, ok := obj["Northing"].(float64); !ok {
		return errors.New("Wrong type for field: \"Northing\"")
	}

	// Check the value for "ZoneNumber" key (if there is one)
	if _, ok := obj["ZoneNumber"]; !ok {
		return errors.New("Missing field: \"ZoneNumber\"")
	}
	if _, ok := obj["ZoneNumber"].(float64); !ok {
		return errors.New("Wrong type for field: \"ZoneNumber\"")
	}

	// Check the value for "ZoneLetter" key (if there is one)
	if _, ok := obj["ZoneLetter"]; !ok {
		return errors.New("Missing field: \"ZoneLetter\"")
	}
	if _, ok := obj["ZoneLetter"].(string); !ok {
		return errors.New("Wrong type for field: \"ZoneLetter\"")
	}

	// Everything's OK! Time to populate the fields.
	c.Easting = obj["Easting"].(float64)
	c.Northing = obj["Northing"].(float64)
	c.ZoneNumber = int(obj["ZoneNumber"].(float64))
	c.ZoneLetter = obj["ZoneLetter"].(string)

	// No error.
	return nil
}
