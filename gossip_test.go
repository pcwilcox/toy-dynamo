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
