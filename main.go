//
// main.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson          lelawson
// Pete Wilcox              pcwilcox
//
// This is the main source file for HW2. It sets up some initialization variables by
// reading the environment, then sets up the two interfaces the application uses. If
// the app is launched as a 'leader', then it will use a kvs object from kvs.go for
// its back end data store. If it is launched as a 'follower' then it will use a
// forwarder object from forward.go as its back end data store. Whichever data store
// is used is passed as an initialization member to an App object from app.go. The
// App object implements the RESTful API front end and communicates with its data
// store in order to satisfy client requests.
//

package main

import (
	"io"
	"log"
	"os"
)

// MultiLogOutput controls logging output to stdout and to a log file
var MultiLogOutput io.Writer

func main() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	MultiLogOutput = io.MultiWriter(os.Stdout, logFile)
	// Set up the logging output to stdout
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetOutput(MultiLogOutput)

	// We'll be using a dbAccess object to interface to the back end
	var k dbAccess

	// Check to see if ${MAINIP} is defined in the environment. If it is, we're a follower.
	envMainIP := os.Getenv("MAINIP")
	log.Println("MAINIP: " + envMainIP)

	if envMainIP == "" {
		// We're the leader, so we need a local key-value store as our dbAccess
		k = NewKVS()
		log.Println("Using local key-value store")
	} else {
		// We're a follower, so we need to set up a forwarder as our dbAccess
		prefix := "http://"
		URL := prefix + envMainIP

		log.Println("Implementing forwarder to address " + URL)
		k = &Forwarder{mainIP: URL}
	}

	// The App object is the front end
	a := App{db: k}

	log.Println("Starting server...")
	// Initialize starts the server
	a.Initialize()
}
