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
func unmarshalLatLonger(s string) (l latlong.LatLonger, err error) {

	// Declarations
	var u utm.Coordinate
	var lt latlong.Coordinate
	var n nvector.Coordinate

	// Check if it is a latlong.Coordinate
	if err := json.Unmarshal([]byte(s), &lt); err == nil {
		return lt, nil
	} else if debug {
		fmt.Println(err)
	}

	// Check if it is a nvector.Coordinate
	if err := json.Unmarshal([]byte(s), &n); err == nil {
		return n, nil
	} else if debug {
		fmt.Println(err)
	}

	// Check if it is a utm.Coordinate
	if err := json.Unmarshal([]byte(s), &u); err == nil {
		return u, nil
	} else if debug {
		fmt.Println(err)
	}

	// Return error if none of the above.
	return nil, errors.New("Cannot unmarshal coordinate: " + s)
}

// loadTrips loads trip information line-by-line from a file and sends
// results over a channel.
func loadTrips(fname string, trips chan trip) {

	// Try to open the 'fname' file.
	if file, err := os.Open(fname); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		// If the file is successfully opened.
		// Close the file when finished.
		defer file.Close()

		// New scanner
		scanner := bufio.NewScanner(file)

		// Temporary trip for sending over channel.
		var tmp = trip{0, nil}

		// Loop through each line of the fname file.
		for scanner.Scan() {

			// Get the current line.
			line := scanner.Text()

			// For storing the id and coord that will be extracted.
			var id int
			var js string

			// Extract the info from the line.
			fmt.Sscanf(line, "%d\t%s", &id, &js)

			// Unpack the latlong.LatLonger and error.
			l, e := unmarshalLatLonger(js)

			// Exit the program if there is an error.
			if e != nil {
				fmt.Fprintln(os.Stderr, e)
				os.Exit(1)
			}

			// If still on the same id,
			// add the latlong.Latlonger to the trip's trajectory.
			if tmp.id == id {
				tmp.trajectory = append(tmp.trajectory, l)
			} else { // Otherwise send off the trip and reset tmp.
				trips <- tmp
				tmp.id = id
				tmp.trajectory = nil
				tmp.trajectory = append(tmp.trajectory, l)
			}
		}
		// Send off the final trip.
		trips <- tmp
	}

	close(trips)
}

// computeDistances continually receives trips over a channel and
// computes the total travel distance for each trip, sending the
// totalled results over a channel.
func computeDistances(trips chan trip, totals chan total) {
	for t := range trips {
		// Reset dist to 0 for each new trip t.
		var dist float64 = 0

		// Add up total distance.
		for i := 0; i < len(t.trajectory)-1; i++ {
			dist += latlong.Distance(t.trajectory[i], t.trajectory[i+1])
		}

		// Send total over the channel when finished.
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

	// Spin up goroutines
	go loadTrips(fname, trips_chan)
	go computeDistances(trips_chan, totals_chan)

	// Output totals until channel is closed.
	for tot := range totals_chan {
		fmt.Println(tot)
	}

}
