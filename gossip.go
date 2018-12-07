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
	kvs       dbAccess
	shardList Shard
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
	if goalTime.Before(time.Now()) {
		needHelp = true
		return true
	}
	return false
}

// GossipHeartbeat contains a forever loop that will check for need of Gossip every 1 second
func (g *GossipVals) GossipHeartbeat() {
	log.Println("Gossip heart starts...")
	setTime()

	for {
		if wakeGossip || shardChange || timesUp() {
			log.Println("Gossip initiated. Ringing TCP")

			gossipee := g.shardList.RandomLocal(2)

			if needHelp {
				for _, bob := range gossipee {
					askForHelp(bob)
				}
				needHelp = false
			} else {
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

			}
			if shardChange {
				gossipee = g.shardList.RandomGlobal(2)
				// Propagate views
				for _, bob := range gossipee {
					sendShardGob(bob, g.shardList.GetShardGlob())
				}
			}
			wakeGossip = false
			setTime()
		}
		// Sleep for half a second before restarting
		time.Sleep(50 * time.Millisecond)
	}
}

// ClockPrune returns a pruned map that only contains the keys that the gossipee needs updating
func (g *GossipVals) ClockPrune(input TimeGlob) TimeGlob {
	own := g.kvs.GetTimeGlob() // getTimeGlob() is in glob branch

	// Find and delete duplicates between two maps
	for k := range own.List {
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
func (g *GossipVals) BuildEntryGlob(inglob TimeGlob) EntryGlob {
	// TODO: understand getEntryGlob() and how to use it
	outglob := g.kvs.GetEntryGlob(inglob)
	return outglob
}

// UpdateKVS takes entryGlob and update its own KVS. End of Gossip protocol
func (g *GossipVals) UpdateKVS(inglob EntryGlob) {
	// Loop through all keys, check for conflicts, and update KVS when necessary.
	for key, aliceEntry := range inglob.Keys {
		if g.ConflictResolution(key, &aliceEntry) {
			g.kvs.OverwriteEntry(key, &aliceEntry)
		}
	}
}

// ConflictResolution returns true if Bob should update with Alice's key
func (g *GossipVals) ConflictResolution(key string, aliceEntry KeyEntry) bool {
	log.Println("Resolving a conflict")
	isSmaller := false
	isLarger := false
	incomparable := false

	log.Printf("Comparing Alice's version '%#v'\n", aliceEntry)
	aMap := aliceEntry.GetClock()
	bMap := g.kvs.GetClock(key)
	log.Println("aMap: ", aMap)
	log.Println("bMap: ", bMap)

	// if bob does NOT have the key, we definitely update w/ Alice's stuff
	if len(bMap) == 0 {
		log.Println("Bob doesn't have the entry: ", key)
		return true // Bob can't possibly beat Alice's key with no corresponding key of it's own
	}
	// else if Bob DOES have the key, we compare causal history & timestamps
	for k, v := range bMap {
		if aMap[k] < v {
			isSmaller = true
		} else if aMap[k] > v {
			isLarger = true
		}
	}
	for k := range aMap {
		if _, exist := bMap[k]; !exist {
			incomparable = true
		}
	}

	if (isSmaller && isLarger) || (!isSmaller && !isLarger) || incomparable {
		// incomparable or identical clocks, later timestamp wins
		if aliceEntry.GetTimestamp().After(g.kvs.GetTimestamp(key)) {
			log.Println("Alice wins with the later timestamp")
			return true // alice wins
		}
		log.Println("Bob wins with a later timestamp")
		return false // bob wins
	} else if isSmaller == false && isLarger == true {
		log.Println("Alice wins with a larger clock")
		return true // alice wins
	}
	log.Println("Bob wins with a larger clock")
	return false // bob wins
}

// UpdateShardList overwrites our view of the shard list
func (g *GossipVals) UpdateShardList(s ShardGlob) {
	g.shardList.Overwrite(s)
}
