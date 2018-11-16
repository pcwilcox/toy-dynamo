package main

import (
	"strings"
	"sync"
	"testing"
	"time"
)

type testEntry struct {
	time    time.Time
	clock   map[string]int
	value   string
	version int
}

func (e *testEntry) GetVersion() int {
	return e.version
}

func (e *testEntry) GetValue() string {
	return e.value
}

func (e *testEntry) GetTimestamp() time.Time {
	return e.time
}

func (e *testEntry) GetClock() map[string]int {
	return e.clock
}

func (e *testEntry) Update(time time.Time, clock map[string]int, val string) {
	// goes nowhere does nothing
}

func (e *testEntry) Alive() bool {
	// TODO: figure this out
	return true
}

func (e *testEntry) Delete(time time.Time) {
	// goes nowhere does nothing
}

// This tests for a key that does not exist in the db, the KVS should return version -1 and alive == false
func TestKVSContainsCheckIfDoesntExist(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}

	alive, version := k.Contains(keyNotHere)

	assert(t, version == -1, "Key found that does not exist.")
	assert(t, !alive, "Contains returned alive for nonexistent key.")
}

// If the key does exist, the KVS should return alive == true and its version
func TestKVSContainsCheckIfDoesExist(t *testing.T) {
	entryExists := testEntry{
		time:    time.Now(),
		clock:   nil,
		value:   valExists,
		version: 1,
	}
	db := map[string]KeyEntry{
		keyExists: &entryExists,
	}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	alive, version := k.Contains(keyExists)
	assert(t, alive, "Contains() returned !alive for existing key.")
	assert(t, version == 1, "Contains() returned incorrect version.")

}

// Get() on an existing key/value pair should return the value for the key
func TestKVSGetExistingVal(t *testing.T) {
	entryExists := testEntry{
		time:    time.Now(),
		clock:   nil,
		value:   valExists,
		version: 1,
	}
	db := map[string]KeyEntry{
		keyExists: &entryExists,
	}

	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	returned, _ := k.Get(keyExists, nil)
	equals(t, valExists, returned)
}

// Get on a non-existing key should return an empty string
func TestKVSGetValDoesntExist(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	returned, _ := k.Get(keyone, nil)
	equals(t, "", returned)
}

// Delete on an existing key/value pair should return true and further Gets() should fail
func TestKVSDeleteExistingKeyValPair(t *testing.T) {
	entryExists := testEntry{
		time:    time.Now(),
		clock:   nil,
		value:   valExists,
		version: 1,
	}
	db := map[string]KeyEntry{
		keyExists: &entryExists,
	}

	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, k.Delete(keyExists, time.Now()), "Did not delete Key Val Pair")
}

// Delete on a key that doesn't exist should return false
func TestKVSDeleteKeyDoesntExist(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, !k.Delete(keyNotHere, time.Now()), "Deleted a keyvalue pair not in data store prior")
}

// Put() with a new key should return true
func TestKVSPutNewKeyNewVal(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, k.Put(keyone, valone, time.Now(), nil), "New key and value were not added")
}

// Overwriting a value should return true
func TestKVSPutExistKeyOverwriteVal(t *testing.T) {
	entryExists := testEntry{
		time:    time.Now(),
		clock:   nil,
		value:   valExists,
		version: 1,
	}
	db := map[string]KeyEntry{
		keyExists: &entryExists,
	}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, k.Put(keyone, valtwo, time.Now(), nil), "Did not overwrite existing key's value")

}

// Put() with an invalid key should return failure
func TestKVSPutInvalidkey(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, !k.Put(invalidKey, valtwo, time.Now(), nil), "Invalid key added")
}

// Put() with an invalid value should fail
func TestKVSPutInvalidVal(t *testing.T) {
	k := NewKVS()
	var b strings.Builder
	b.Grow(1048577)
	for i := 0; i < 1048577; i++ {
		b.WriteByte(0)
	}
	invalidVal := b.String()
	assert(t, !k.Put(keyone, invalidVal, time.Now(), nil), "Invalid value added")
}

// GetVersion() should return the version
func TestEntryGetVersion(t *testing.T) {
	e := Entry{
		version:   1,
		timestamp: time.Now(),
		clock: map[string]int{
			keyExists: 1,
		},
		value:     valExists,
		tombstone: false,
	}

	assert(t, e.GetVersion() == 1, "GetVersion() returned wrong value")
}

// GetVersion of an existing entry should return the version
func TestEntryGetVersionKeyExists(t *testing.T) {
	e := Entry{
		version:   1,
		timestamp: time.Now(),
		clock: map[string]int{
			keyExists: 1,
		},
		value:     valExists,
		tombstone: false,
	}

	assert(t, e.GetVersion() == 1, "GetVersion() returned wrong value")
}

// GetVersion of a non-existing key should return 0
func TestEntryGetVersionKeyNotExists(t *testing.T) {
	var e Entry
	assert(t, e.GetVersion() == 0, "GetVersion() returned wrong value")
}

// GetTimestamp should return the timestamp of the key if it exists
func TestEntryGetTimestampKeyExists(t *testing.T) {
	timestamp := time.Now()
	e := Entry{
		version:   1,
		timestamp: timestamp,
		clock: map[string]int{
			keyExists: 1,
		},
		value:     valExists,
		tombstone: false,
	}
	assert(t, e.GetTimestamp() == timestamp, "GetTimestamp() returned wrong value")
}

// GetTimestamp should return the zero-value timestamp if the key doesn't exist
func TestEntryGetTimestampKeyNotExists(t *testing.T) {
	var e Entry
	assert(t, e.GetTimestamp() == time.Time{}, "GetTimestamp() returned wrong value")
}

// GetClock should return the key's clock map if it exists
func TestEntryGetClockKeyExists(t *testing.T) {
	expectedClock := map[string]int{
		keyExists: 1,
	}
	e := Entry{
		version:   1,
		timestamp: time.Now(),
		clock:     expectedClock,
		value:     valExists,
		tombstone: false,
	}
	equals(t, e.GetClock(), expectedClock)

}

// GetClock should return an empty map for a key which doesn't exist
func TestEntryGetClockKeyNotExists(t *testing.T) {
	e := Entry{}
	assert(t, len(e.GetClock()) == 0, "GetClock() returned non-empty map for nonexisting key")
}

// GetValue should return the value for a key which exists
func TestEntryGetValueKeyExists(t *testing.T) {
	e := Entry{
		version:   1,
		timestamp: time.Now(),
		clock:     map[string]int{keyExists: 1},
		value:     valExists,
		tombstone: false,
	}
	equals(t, e.GetValue(), valExists)
}

// GetValue should return an empty string for a key which doesn't exist
func TestEntryGetValueKeyNotExists(t *testing.T) {
	e := Entry{}
	equals(t, e.GetValue(), "")
}

// Update should change the timestamp
func TestUpdateChangesTimestamp(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		version:   1,
		timestamp: start,
		clock:     map[string]int{keyExists: 1},
		value:     valExists,
		tombstone: false,
	}

	e.Update(finish, map[string]int{}, valExists)
	equals(t, e.GetTimestamp(), finish)
}

// Update should increment the version
func TestUpdateIncrementsVersion(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		version:   1,
		timestamp: start,
		clock:     map[string]int{keyExists: 1},
		value:     valExists,
		tombstone: false,
	}

	e.Update(finish, map[string]int{}, valExists)
	equals(t, e.GetVersion(), 2)
}

// Update should reset the tombstone
func TestUpdateClearsTomstone(t *testing.T) {
	start := time.Now()
	time.Sleep(1)
	finish := time.Now()

	e := Entry{
		version:   1,
		timestamp: start,
		clock:     map[string]int{keyExists: 1},
		value:     valExists,
		tombstone: true,
	}

	e.Update(finish, map[string]int{}, valExists)
	equals(t, e.Alive(), true)

}

// These functions were taken from Ben Johnson's post here: https://medium.com/@benbjohnson/structuring-tests-in-go-46ddee7a25c

// // assert fails the test if the condition is false.
// func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
// 	if !condition {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
// 		tb.FailNow()
// 	}
// }

// // ok fails the test if an err is not nil.
// func ok(tb testing.TB, err error) {
// 	if err != nil {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
// 		tb.FailNow()
// 	}
// }

// // equals fails the test if exp is not equal to act.
// func equals(tb testing.TB, exp, act interface{}) {
// 	if !reflect.DeepEqual(exp, act) {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
// 		tb.FailNow()
// 	}
// }
