package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"latlong"
	"log"
	"nvector"
	"os"
	"utm"
)

var (
	// True if we want to see debug output, otherwise false.
	// Set by the user with the -debug flag
	debug bool
)

// parseCLIArgs parses options from the command line.
//
// Returns the name of the user-provided data file
func parseCLIArgs() string {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:  %s <filename>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&debug, "debug", false, "enable debug output")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Need a file to process!\n\n")
		flag.Usage()
		os.Exit(1)
	}

	return flag.Arg(0)
}

// unmarshalLatLonger attempts to unmarshal a JSON encoded
// latlong.LatLonger coordinate.
//
// The coordinate may be a JSON encoded latlong.Coordinate,
// nvector.Coordinate, or utm.Coordinate.
//
// For each of the above coordinate types, unmarshalLatLonger attempts
// to unmarshal the string. It starts with latlong.Coordinate. If it
// successfully unmarshals the string as a latlong.Coordinate, it
// returns it along with a nil error. If it fails, it tries to
// unmarshal it as a nvector.Coordinate. unmarshalLatLonger tries each
// type until one succeeds. If it fails to unmarshal the string to
// **any** of the above coordinate types, it returns a non-nil error.
//
// If unmarshaling is successful, the coordinate is returned as a latlong.LatLonger.
func unmarshalLatLonger(s string) (l latlong.LatLonger, err error) {
	var u utm.Coordinate
	var lt latlong.Coordinate
	var n nvector.Coordinate

	if err := json.Unmarshal([]byte(s), &lt); err == nil {
		return lt, nil
	} else if debug {
		fmt.Println(err)
	}

	if err := json.Unmarshal([]byte(s), &n); err == nil {
		return n, nil
	} else if debug {
		fmt.Println(err)
	}

	if err := json.Unmarshal([]byte(s), &u); err == nil {
		return u, nil
	} else if debug {
		fmt.Println(err)
	}

	return nil, errors.New("Cannot unmarshal coordinate: " + s)
}

// loadTrips loads trip information line-by-line from a file and sends
// results over a channel.
//
// Refer to online documentation for format expectations
//
// Attempts to open a file and read its contents line-by-line. As it
// reads through the file, loadTrips groups coordinates by traveler
// ID. Records are aggregated for each traveler ID. Once all
// coordinates for a traveler have been seen, the trip information is
// sent over the trips channel.
//
// When loadTrips finishes processing all of the lines in the file and
// sends the final trip over the output channel, it closes the output
// channel to signal that nothing is left.
func loadTrips(fname string, trips chan trip) {
	// Try to open the file.
	if file, err := os.Open(fname); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		// Close the file when finished.
		defer file.Close()

		scanner := bufio.NewScanner(file)

        var tmp = trip{0, nil} 

		// Loop through each line.
		for scanner.Scan() {
			line := scanner.Text()
            
			// For storing the id and coord that will be extracted.
			var id int
			var js string

			// Extracting info from the line.
			fmt.Sscanf(line, "%d\t%s", &id, &js)

            l, e := unmarshalLatLonger(js)
            
            if e != nil {
            	fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
            }
            
			// If still on the same id.
			if tmp.id == id {
					tmp.trajectory = append(tmp.trajectory, l)
			} else {
				trips <- tmp
				tmp.id = id
				tmp.trajectory = nil
                tmp.trajectory = append(tmp.trajectory, l)
			}
		}
        trips <- tmp
	}

	close(trips)
}

// computeDistances continually receives trips over a channel and
// computes the total travel distance for each trip, sending the
// totalled results over a channel.
//
// After the distance of the last trip has been calculated and sent
// over the output channel (totals), computeDistances closes the
// channel to indicate that there will be no more results.
func computeDistances(trips chan trip, totals chan total) {
	for t := range trips {
		var dist float64
		dist = 0

		for i := 0; i < len(t.trajectory)-1; i++ {
			dist += latlong.Distance(t.trajectory[i], t.trajectory[i+1])
		}

		totals <- total{t.id, dist}
	}
	close(totals)
}

func main() {
	fname := parseCLIArgs()

	log.SetFlags(0) // Dial back the log output
	if debug {
		log.Printf("Starting program %s", os.Args[0])
	}

	// Initialized necessary channels.
	trips_chan := make(chan trip)
	totals_chan := make(chan total)

	go loadTrips(fname, trips_chan)

	go computeDistances(trips_chan, totals_chan)

	for tot := range totals_chan {
		fmt.Println(tot)
	}

	//fmt.Println(fname)

	//fmt.Println(unmarshalLatLonger("{\"Latitude\":37.924782627013734,\"Longitude\":-91.63306471017802}"))
	//fmt.Println(unmarshalLatLonger("{\"X\":-1.3080905091896413,\"Y\":-45.24470795787057,\"Z\":35.12850197544032}"))
	//fmt.Println(unmarshalLatLonger("{\"Easting\":606669.8132040,\"Northing\":4.21706174608e+06,\"ZoneNumber\":15,\"ZoneLetter\":\"S\"}"))
}
