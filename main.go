// main.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson          lelawson
// Pete Wilcox              pcwilcox
// Annie Shen				ashen7
// Victoria Tran            vilatran
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
	"strconv"
)

// Versioning info defined via linker flags at compile time
var branch string // Git branch
var hash string   // Shortened commit hash
var build string  // Number of commits in the branch

// MyShard is our shard system
var MyShard *ShardList

// MultiLogOutput controls logging output to stdout and to a log file
var MultiLogOutput io.Writer

func main() {
	// Create a logfile
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	// Create a stream that writes to console and the logfile
	MultiLogOutput = io.MultiWriter(os.Stdout, logFile)
	// Set some logging flags and setup the logger
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetOutput(MultiLogOutput)

	// Print version info to the log
	version := branch + "." + hash + "." + build
	log.Println("Running version " + version)

	// IP_PORT is defined at runtime in the docker command
	myIP = os.Getenv("IP_PORT")

	log.Println("My IP is " + myIP)

	// docker run -p 8082:8080
	// --net=mynetwork
	// --ip=192.168.0.2
	// -e VIEW="192.168.0.2:8080,192.168.0.3:8080,192.168.0.4:8080,192.168.0.5:8080"
	// -e IP_PORT="192.168.0.2:8080" -e S=”2” testing

	// VIEW is defined at runtime in the docker command as a string
	view := os.Getenv("VIEW")
	log.Println("My view is: " + view)

	// Store s as the number of shards from env
	// docker run -p 8082:8080 --net=mynet --ip=10.0.0.2 -e VIEW="10.0.0.2:8080,10.0.0.3:8080,10.0.0.4:8080" -e IP_PORT="10.0.0.2:8080" -e S="3" REPLICA_1
	s := os.Getenv("S")
	var numshards int
	log.Println("There is total of " + s + " shard(s)")
	if s == "" {
		numshards = defaultShardCount
	} else {
		// Convert string to int
		numshards, err = strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
	}
	// Create a ShardList and create the seperation of shard ID to servers
	MyShard = NewShard(myIP, view, numshards)

	log.Println("My shard ID is ", MyShard.PrimaryID())

	// Make a KVS to use as the db
	k := NewKVS()

	// The App object is the front end and has references to the KVS and viewList
	a := App{db: k, shard: *MyShard}

	log.Println("Starting server...")

	// The gossip object controls communicating with other servers and has references to the viewlist and the kvs
	gossip := GossipVals{
		kvs:       k,
		ShardList: MyShard,
	}
	// Start the heartbeat loop
	go gossip.GossipHeartbeat() // goroutines

	// Start the servers with references to the REST app and the gossip module
	server(&a, &gossip)
}
