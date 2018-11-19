// gossip.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson     lelawson
// Pete Wilcox         pcwilcox
// Annie Shen          ashen7
// Victoria Tran       vilatran
//
// This source file defines the full functionalities for doing synchronization
// across multiple servers
//

package main

import (
	"log"
	"time"
)

// GossipVals is a struct which implements the Gossip
type GossipVals struct {
	view View
	kvs  dbAccess
	// tcp?
}

// Global variable for easier time tracking
var now time.Time
var goalTime time.Time

// setTime updates the time variables when needed
func setTime() {
	// Set the time right now
	now = time.Now()
	// Set the time goal for 5 seconds after timeNow
	goalTime = now.Add(5 * time.Second)
}

// timesUp purely checks if 5 seconds has past
func timesUp() bool {
	if time.Now() == goalTime {
		return true
	}
	return false
}

// GossipHeartbeat contains a forever loop that will check for need of Gossip every 1 second
func (g *GossipVals) GossipHeartbeat() {
	log.Println("Gossip heart starts...")
	setTime()
	// Implement TCP for the server to listen for connection
	go server()

	for {
		if wakeGossip == true || timesUp() == true {
			log.Println("Gossip initiated. Ringing TCP")

			gossipee := g.view.Random(2)

			for _, bob := range gossipee {
				// Get timeglob
				t := g.kvs.GetTimeGlob()
				//Send our timeglob to gossipee and return back their pruned timeglob
				rt, err := sendTimeGlob(bob, t)
				if err != nil {
					log.Println("Error sending timeglob: ", err)
					continue
				}
				// turn the pruned timeglob into and entry glob for gossipee
				re := g.kvs.GetEntryGlob(*rt)
				//send the entryglob needed to update gosipee kvs
				err = sendEntryGlob(bob, re)
				if err != nil {
					log.Println("Error sending entryglob: ", err)
					continue
				}
			}
			wakeGossip = false
			setTime()
		}
		// Sleep for half a second before restarting
		time.Sleep(500 * time.Millisecond)
	}
}

// ClockPrune returns a pruned map that only contains the keys that the gossipee needs updating
func (g *GossipVals) ClockPrune(input timeGlob) timeGlob {
	own := g.kvs.GetTimeGlob() // getTimeGlob() is in glob branch

	// Find and delete duplicates between two maps
	for k, _ := range own.List {
		// Prune key off of input timeGlob if input's k is as new as own's k
		if input.List[k] == own.List[k] {
			delete(input.List, k)
		}
		// Does NOT prune even if input[k] < own[k], because further checks in causal
		// history is needed to make sure which version of key to keep.
	}
	// return the editted map containing only keys than the gossipee wants
	return input
}

// BuildEntryGlob takes timeGlob and turn it into entryGlob
func (g *GossipVals) BuildEntryGlob(inglob timeGlob) entryGlob {
	// TODO: understand getEntryGlob() and how to use it
	outglob := g.kvs.GetEntryGlob(inglob)
	return outglob
}

// UpdateKVS takes entryGlob and update its own KVS. End of Gossip protocol
func (g *GossipVals) UpdateKVS(inglob entryGlob) {
	// Loop through all keys, check for conflicts, and update KVS when necessary.
	for key, aliceEntry := range inglob.Keys {
		if g.ConflictResolution(key, aliceEntry) {
			g.kvs.OverwriteEntry(key, aliceEntry)
		}
	}
}

// A timeGlob is a map of keys to timestamps and lets the gossip module figure out which ones need to be updated
func (g *GossipVals) ConflictResolution(key string, aliceEntry KeyEntry) bool {
	log.Println("Resolving a conflict")
	isSmaller := false
	isLarger := false

	aMap := aliceEntry.GetClock()
	bMap := g.kvs.GetClock(key)

	// if bob does NOT have the key, we definitely update w/ Alice's stuff
	if len(bMap) == 0 {
		return true // Bob can't possibly beat Alice's key with no corresponding key of it's own
	} else {
		// else if Bob DOES have the key, we compare causal history & timestamps
		for k, v := range bMap {
			if aMap[k] < v {
				isSmaller = true
			} else if aMap[k] > v {
				isLarger = true
			}
		}

		if (isSmaller && isLarger) || (!isSmaller && !isLarger) {
			// incomparable or identical clocks, later timestamp wins
			if aliceEntry.GetTimestamp().Sub(g.kvs.GetTimestamp(key)) > 0 {
				return true // alice wins
			} else {
				return false // bob wins
			}
		} else if isSmaller == false && isLarger == true {
			return true // alice wins
		} else {
			return false // bob wins
		}
	}
}
