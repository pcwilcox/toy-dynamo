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

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-test/deep"
)

type TestShard struct {
}

func (s *TestShard) Count() int {
	return 1
}

func (s *TestShard) ContainsShard(in string) bool {
	return true
}

func (s *TestShard) ContainsServer(in string) bool {
	return true
}

func (s *TestShard) FindBob(in string) string {
	return "bob"
}

func (s *TestShard) GetIP() string {
	return "an IP"
}
func (s *TestShard) Remove(in string) bool {
	return true
}

func (s *TestShard) Add(in string) bool {
	return true
}

func (s *TestShard) Random(n int) []string {
	return []string{}
}

func (s *TestShard) PrimaryID() string {
	return ""
}

func (s *TestShard) RandomGlobal(n int) []string {
	return []string{}
}

func (s *TestShard) RandomLocal(n int) []string {
	return []string{}
}

func (s *TestShard) List() []string {
	return []string{}
}

func (s *TestShard) String() string {
	return ""
}

func (s *TestShard) Overwrite(in ShardGlob) {
	// does nothing
}

func (s *TestShard) GetShardGlob() ShardGlob {
	return ShardGlob{}
}
func TestSetTimeSetsTime(t *testing.T) {
	before := time.Now()

	time.Sleep(1 * time.Millisecond)
	setTime()
	time.Sleep(1 * time.Millisecond)

	after := time.Now()

	assert(t, before.Before(now), "SetTime didn't set the time to be after 'before'")
	assert(t, now.Before(after), "SetTime didn't set the time before 'after'")
}

func TestSetTimeSetsGoal(t *testing.T) {
	setTime()
	assert(t, now.Add(5*time.Second) == goalTime, "SetTime set the wrong goal")
}
func TestTimesUpSetsNeedHelp(t *testing.T) {
	setTime()
	time.Sleep(5 * time.Second)
	assert(t, timesUp(), "Times up failed")
	fmt.Println(goalTime)
	assert(t, needHelp, "Times up didn't set needHelp")
}

func TestTimesUpReturnsFalseIfEarly(t *testing.T) {
	setTime()
	assert(t, !timesUp(), "TimesUp returned true early")
}

func TestClockPrunePrunesClocks(t *testing.T) {
	// Mock a KVS
	timeExists := time.Now()
	k := TestKVS{
		dbClock: map[string]int{keyExists: 1},
		dbKey:   keyExists,
		dbTime:  timeExists,
		dbVal:   valExists,
	}

	// Mock a view
	s := TestShard{}

	// Make a gossip
	g := GossipVals{kvs: &k, shardList: &s}

	// Mock an input
	ig := TimeGlob{
		List: map[string]time.Time{
			keyExists:    timeExists,
			keyNotExists: time.Now(),
		},
	}

	og := g.ClockPrune(ig)

	_, pruned := og.List[keyExists]    // This key should have been removed
	_, exists := og.List[keyNotExists] // this one shouldn't have
	assert(t, !pruned, "ClockPrune didn't prune a matching entry")
	assert(t, exists, "ClockPrune pruned a non-matching entry")
}

func TestBuildEntryGlobBuildsGlob(t *testing.T) {
	// Mock a KVS
	timeExists := time.Now()
	k := TestKVS{
		dbClock:   map[string]int{keyExists: 1},
		dbKey:     keyExists,
		dbTime:    timeExists,
		dbVal:     valExists,
		dbVersion: 1,
	}

	// Mock a view
	s := TestShard{}

	// Make a gossip
	g := GossipVals{kvs: &k, shardList: &s}

	// Mock an input
	ig := TimeGlob{
		List: map[string]time.Time{
			keyExists:    timeExists,
			keyNotExists: time.Now(),
		},
	}

	eg := g.BuildEntryGlob(ig)

	te := Entry{
		Version:   1,
		Value:     valExists,
		Timestamp: timeExists,
		Clock:     k.dbClock,
		Tombstone: false,
	}
	teg := EntryGlob{Keys: map[string]Entry{keyExists: te}}

	if diff := deep.Equal(eg, teg); diff != nil {
		t.Error(diff)
	}

}

func TestConflictResolutionKeyNotExist(t *testing.T) {
	// Mock a KVS
	timeExists := time.Now()
	k := TestKVS{
		dbClock: map[string]int{keyExists: 1},
		dbKey:   keyExists,
		dbTime:  timeExists,
		dbVal:   valExists,
	}

	// Mock a view
	s := TestShard{}

	// Make a gossip
	g := GossipVals{kvs: &k, shardList: &s}

	// Mock an input that Bob doesn't have
	keyNotExistsEntry := Entry{
		Version:   1,
		Value:     valNotExists,
		Timestamp: timeExists,
		Clock:     map[string]int{keyNotExists: 1},
		Tombstone: false,
	}

	// Make sure Alice's entry wins
	assert(t, g.ConflictResolution(keyNotExists, &keyNotExistsEntry), "ConflictResolution didn't pick Alice win")
}

func TestConflictResolutionKeyExistAliceBigger(t *testing.T) {
	// Mock a KVS
	timeExists := time.Now()
	k := TestKVS{
		dbClock: map[string]int{keyExists: 1},
		dbKey:   keyExists,
		dbTime:  timeExists,
		dbVal:   valExists,
	}

	// Mock a view
	s := TestShard{}

	// Make a gossip
	g := GossipVals{kvs: &k, shardList: &s}

	// Mock an input that Bob doesn't have
	newKeyExistsEntry := Entry{
		Version:   2,
		Value:     valNotExists,
		Timestamp: timeExists,
		Clock:     map[string]int{keyExists: 2},
		Tombstone: false,
	}

	// Make sure Alice's entry wins
	assert(t, g.ConflictResolution(keyExists, &newKeyExistsEntry), "ConflictResolution didn't pick Alice win")
}

func TestConflictResolutionKeyExistBobBigger(t *testing.T) {
	// Mock a KVS
	timeExists := time.Now()
	k := TestKVS{
		dbClock: map[string]int{keyExists: 2},
		dbKey:   keyExists,
		dbTime:  timeExists,
		dbVal:   valExists,
	}

	// Mock a view
	s := TestShard{}

	// Make a gossip
	g := GossipVals{kvs: &k, shardList: &s}

	// Mock an input that Bob doesn't have
	newKeyExistsEntry := Entry{
		Version:   1,
		Value:     valNotExists,
		Timestamp: timeExists,
		Clock:     map[string]int{keyExists: 1},
		Tombstone: false,
	}

	// Make sure Bob's entry wins
	assert(t, !g.ConflictResolution(keyExists, &newKeyExistsEntry), "ConflictResolution didn't pick Bob's entry to win")
}

func TestConflictResolutionKeyExistAliceLater(t *testing.T) {
	// Mock a KVS
	timeExists := time.Now()
	k := TestKVS{
		dbClock: map[string]int{keyExists: 2, "someKey": 1},
		dbKey:   keyExists,
		dbTime:  timeExists,
		dbVal:   valExists,
	}

	// Mock a view
	s := TestShard{}

	// Make a gossip
	g := GossipVals{kvs: &k, shardList: &s}
	time.Sleep(1 * time.Second)

	aliceTime := time.Now()

	// Mock an input that Bob doesn't have
	newKeyExistsEntry := Entry{
		Version:   2,
		Value:     valNotExists,
		Timestamp: aliceTime,
		Clock:     map[string]int{keyExists: 2, "someOtherKey": 1},
		Tombstone: false,
	}

	// Make sure Alice's entry wins
	assert(t, g.ConflictResolution(keyExists, &newKeyExistsEntry), "ConflictResolution didn't pick Alice to win")
}

func TestConflictResolutionKeyExistBobLater(t *testing.T) {
	// Mock a KVS
	aliceTime := time.Now()
	time.Sleep(1 * time.Second)
	bobTime := time.Now()
	k := TestKVS{
		dbClock: map[string]int{keyExists: 2, "someKey": 2},
		dbKey:   keyExists,
		dbTime:  bobTime,
		dbVal:   valExists,
	}

	// Mock a view
	s := TestShard{}

	// Make a gossip
	g := GossipVals{kvs: &k, shardList: &s}

	// Mock an input that Bob doesn't have
	newKeyExistsEntry := Entry{
		Version:   2,
		Value:     valNotExists,
		Timestamp: aliceTime,
		Clock:     map[string]int{keyExists: 2, "someOtherKey": 2},
		Tombstone: false,
	}

	// Make sure Bob's entry wins
	assert(t, !g.ConflictResolution(keyExists, &newKeyExistsEntry), "ConflictResolution didn't pick Bob's entry to win")
}

func TestUpdateKVSUpdatesKVS(t *testing.T) {
	// Mock a KVS
	timeExists := time.Now()
	k := TestKVS{
		dbClock: map[string]int{keyExists: 1},
		dbKey:   keyExists,
		dbTime:  timeExists,
		dbVal:   valExists,
	}

	// Mock a view
	s := TestShard{}

	// Make a gossip
	g := GossipVals{kvs: &k, shardList: &s}

	// Mock an input that Bob doesn't have
	newKeyExistsEntry := Entry{
		Version:   2,
		Value:     valNotExists,
		Timestamp: timeExists,
		Clock:     map[string]int{keyExists: 2},
		Tombstone: false,
	}
	teg := EntryGlob{Keys: map[string]Entry{keyExists: newKeyExistsEntry}}

	g.UpdateKVS(teg)

	tg := TimeGlob{List: map[string]time.Time{keyExists: timeExists}}

	eg := g.kvs.GetEntryGlob(tg)

	if diff := deep.Equal(eg, teg); diff != nil {
		t.Error(diff)
	}

}

//
// Test GossipHeartbeat
//   - mock a viewlist to return a specific set of Bobs
//   - mock a KVS to return a specific timeglob and entryglob
//   - mock a TCP listener or something so we can test SendTimeGlob and SendEntryGlob are called correctly
//   - this function does a lot of stuff so we might want to refactor it
//
// Test nil stuff
//   - we should not have any panics because of nil entries or objects
//   - log.fatalln instead or failthrough or something
