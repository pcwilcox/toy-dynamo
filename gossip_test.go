// gossip_test.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson   lelawson
// Pete Wilcox       pcwilcox
// Annie Shen        ashen7
// Victoria Tran     vilatran
//
// Unit tests for the gossip struct

package main

/*
func TestConflictResolutionAliceWins(t *testing.T) {
	g := GossipVals{}
	alice := Entry{
		version:   1,
		timestamp: time.Now(),
		clock: map[string]int{
			"keyA": 1,
			"keyB": 0,
			"keyC": 1,
		},
		value:     valExists,
		tombstone: false,
	}

	bob := Entry{
		version:   1,
		timestamp: time.Now(),
		clock: map[string]int{
			"keyA": 1,
			"keyB": 1,
			"keyC": 1,
		},
		value:     valExists,
		tombstone: false,
	}
	// bob should win, func should return 1
	//equals(t, bob.ConflictResolution(alice.getClock(), alice.getTimestamp()), 1)
}

*/

// TODO: Tests needed
//
// Test setTime sets now
//
// Test setTime sets goalTime = now + 5
//
// Test timesUp works to end loop
//
// Test timesUp sets needHelp
//
// Test timesUp returns false correctly
//
// Test GossipHeartbeat
//   - mock a viewlist to return a specific set of Bobs
//   - mock a KVS to return a specific timeglob and entryglob
//   - mock a TCP listener or something so we can test SendTimeGlob and SendEntryGlob are called correctly
//   - this function does a lot of stuff so we might want to refactor it
//
// Test ClockPrune:
//   - mock a GetTimeGlob function to return a specific glob
//   - Mock an input timeglob
//   - check that it prunes the input correctly
//
// Test BuildEntryGlob:
//   - Mock an input timeglob
//   - mock a kvs to return an output glob
//   - test the output glob
//
// Test UpdateKVS:
//   - mock an input entryglob
//   - mock a KVS
//   - check that correct entries get updated
//
// Test ConflictResolution:
//   - Mock a kvs
//   - mock an input entry
//   - make sure the right one wins
//   - repeat for each possible scenario
//
// Test UpdateViews
//   - Mock a viewlist
//   - mock an input
//   - make sure the viewlist gets updated
//
// Test nil stuff
//   - we should not have any panics because of nil entries or objects
//   - log.fatalln instead or failthrough or something
