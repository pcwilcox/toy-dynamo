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
	"log"
	"os"
)

func main() {
	// Set up the logging output to stdout
	log.SetOutput(os.Stdout)

	// TODO: Change this to check the environment later
	leader := true

	// This object represents the interface between the REST API and the data store
	var k dbAccess

	if leader == true {
		// We're the leader so we'll set up a local data store
		k = NewKVS()
	}
	// TODO: else we'll set up a remote

	// The App object is the front end
	a := App{db: k}

	// Initialize starts the server
	a.Initialize()
}
