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
	ShardList Shard
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
			log.Println("Heartbeat")

			gossipee := g.ShardList.RandomLocal(2)

			if needHelp {
				for _, bob := range gossipee {
					log.Println("Asking for help from ", bob)
					askForHelp(bob)
				}
				needHelp = false
			} else {
				for _, bob := range gossipee {
					log.Println("sending timeglob to ", bob)
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
				gossipee = g.ShardList.EveryoneElse()
				// Propagate views
				for _, bob := range gossipee {
					log.Println("sending shardlist update to ", bob)
					go sendShardGob(bob, g.ShardList.GetShardGlob())
					shardChange = false
				}

				distributeKeys = true
				go g.kvs.ShuffleKeys()
			}
			wakeGossip = false
			setTime()
		}
		// Sleep
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
		g.kvs.ConflictResolution(key, &aliceEntry)
	}
}

// UpdateShardList overwrites our view of the shard list
func (g *GossipVals) UpdateShardList(s ShardGlob) {
	g.ShardList.Overwrite(s)
}
