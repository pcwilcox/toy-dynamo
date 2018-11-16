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

// This tests for a key that does not exist in the db, the KVS should return version -1
func TestKVSContainsCheckIfDoesntExist(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}

	alive, version := k.Contains(keyNotHere)

	assert(t, version == -1, "Key found that does not exist.")
	assert(t, !alive, "Contains returned alive for nonexistent key.")
}
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

func TestKVSGetValDoesntExist(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	returned, _ := k.Get(keyone, nil)
	equals(t, "", returned)
}

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

func TestKVSDeleteKeyDoesntExist(t *testing.T) {
	db := map[string]KeyEntry{}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, !k.Delete(keyNotHere, time.Now()), "Deleted a keyvalue pair not in data store prior")
}

func TestKVSPutNewKeyNewVal(t *testing.T) {
	k := NewKVS()
	assert(t, k.Put(keyone, valone, time.Now(), nil), "New key and value were not added")
}

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

func TestKVSPutInvalidkey(t *testing.T) {
	k := NewKVS()
	assert(t, !k.Put(invalidKey, valtwo, time.Now(), nil), "Invalid key added")

}

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
