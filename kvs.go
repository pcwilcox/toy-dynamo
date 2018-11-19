// KVS.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson            lelawson
// Pete Wilcox                pcwilcox
// Annie Shen                 ashen7
//
// This source file defines the KVS object used for a local data store. The struct
// implements a map to store key-value pairs, and implements the dbAccess interface
// defined in dbAccess.go in order to serve as a db object for the REST API.
//

package main

import (
	"log"
	"sync"
	"time"
)

// KVS represents a key-value store and implements the dbAccess interface
type KVS struct {
	db    map[string]KeyEntry
	mutex *sync.RWMutex
}

// KeyEntry interface defines methods to get the info associated with a key, and to update them accordingly
type KeyEntry interface {

	// Return the version number
	GetVersion() int

	// Return the timestamp number
	GetTimestamp() time.Time

	// Return the map representing the causal history
	GetClock() map[string]int

	// Return the actual value
	GetValue() string

	// Write a new value - this needs to update the version and overwrite the clock
	Update(time.Time, map[string]int, string)

	// Write tombstone
	Delete(time.Time)

	// Returns true if it hasn't been tombstone
	Alive() bool
}

// Entry is the thing in the KVS and implements all the methods
type Entry struct {
	version   int            // Monotonically increasing version numbers starting at 1
	timestamp time.Time      // this is set on writes
	clock     map[string]int // This is captured from the client payload on write
	value     string         // This is the actual value
	tombstone bool           // Tombstone value showing that it was deleted
}

// NewEntry creates a new entry
func NewEntry(time time.Time, clock map[string]int, val string) *Entry {

	// Make a new entry
	var e Entry

	// All keys start at version 1
	e.version = 1

	// The timestamp is set at creation
	e.timestamp = time

	// There might be no causal history initially but we still need to create the map
	e.clock = make(map[string]int)
	if clock != nil && len(clock) > 0 {
		for k, v := range clock {
			e.clock[k] = v
		}
	}

	// Tombstone obviously should be false
	e.tombstone = false

	// Finally, set the value
	e.value = val

	// Return a pointer to the entry
	return &e
}

// GetVersion just returns the version
func (e *Entry) GetVersion() int {
	if e != nil {
		return e.version
	}
	// The version of a key which doesn't exist is -1
	return 0
}

// GetTimestamp returns the timestamp from the entry
func (e *Entry) GetTimestamp() time.Time {
	if e != nil {
		return e.timestamp
	}
	// The timestamp of a key which doesn't exist is the empty struct Time{} which evaluates to a ton of 0's
	return time.Time{}
}

// GetClock returns the map representing the causal history
func (e *Entry) GetClock() map[string]int {
	if e != nil {
		return e.clock
	}
	// The clock of a non-existing key is an empty map
	empty := make(map[string]int)
	return empty
}

// GetValue returns the string stored in the entry
func (e *Entry) GetValue() string {
	if e != nil {
		return e.value
	}
	// The value of a non-existing key is an empty string
	return ""
}

// Update writes a new value for the entry and updates the clock and version info
func (e *Entry) Update(newTime time.Time, newClock map[string]int, newVal string) {
	e.timestamp = newTime
	e.value = newVal
	e.clock = newClock
	e.tombstone = false
	e.version++
}

// Delete sets a tombstone that the key has been tombstone
func (e *Entry) Delete(newTime time.Time) {
	e.timestamp = newTime
	e.value = ""
	e.clock = map[string]int{}
	e.tombstone = true
	e.version++
}

// Alive returns true if the key exists and doesn't have a tombstone set
func (e *Entry) Alive() bool {
	if e != nil && e.tombstone != true {
		return true
	}
	return false
}

// NewKVS initializes a KVS object and returns a pointer to it
func NewKVS() *KVS {
	var k KVS
	k.db = make(map[string]KeyEntry)
	var m sync.RWMutex
	k.mutex = &m
	return &k
}

// Contains returns true if the dbAccess object contains an object with key equal to the input, it checks the input payload to ensure proper version
func (k *KVS) Contains(key string) (bool, int) {
	log.Println("Checking to see if db contains key ")
	// Grab a read lock
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	// Once the read lock has been obtained, call the non-locking contains() method
	alive, version := k.contains(key)
	return alive, version
}

// contains is the unexported version of Contains() and does not hold a read lock
func (k *KVS) contains(key string) (bool, int) {
	t := k.db[key]
	if t != nil {
		return t.Alive(), t.GetVersion()
	}
	return false, 0
}

// Get returns the value associated with a particular key. If the key does not exist it returns ""
func (k *KVS) Get(key string, payload map[string]int) (val string, clock map[string]int) {
	log.Println("Getting value associated with key ")
	// Grab a read lock
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	_, version := k.contains(key)

	// Call the non-locking contains() method, use the version from above with default value 0
	if version != 0 {
		log.Println("Value found")

		// Get the key and clock from the db
		val = k.db[key].GetValue()
		clock = k.db[key].GetClock()

		// Add this key's causal history to the client's payload
		clock = mergeClocks(payload, clock)

		// Return
		return val, clock
	}
	log.Println("Value not found")

	// We don't have the value so just return the empty string with the payload they sent us
	return "", payload
}

// Delete sets the tombstone associated with a particular key, updates its version and timestamp, so it appears dead
func (k *KVS) Delete(key string, time time.Time) bool {
	log.Println("Attempting to delete key ")
	// Grab a write lock
	k.mutex.Lock()
	defer k.mutex.Unlock()

	doesExist, _ := k.contains(key)

	// Call the nonlocking contains method
	if doesExist {
		log.Println("Key found, deleting key-value pair")
		k.db[key].Delete(time)
		return true
	}
	log.Println("Key not found")
	return false
}

// Put adds a key-value pair to the DB. If the key already exists, then it overwrites the existing value. If the key does not exist then it is added.
func (k *KVS) Put(key string, val string, time time.Time, payload map[string]int) bool {
	maxVal := 1048576 // 1 megabyte
	maxKey := 200     // 200 characters
	keyLen := len(key)
	valLen := len(val)

	log.Println("Attempting to insert key-value pair")

	if keyLen <= maxKey && valLen <= maxVal {
		log.Println("Key and value OK, inserting to DB")
		// Grab a write lock
		k.mutex.Lock()
		defer k.mutex.Unlock()

		doesExist, _ := k.contains(key)

		// Check to see if the key exists
		if doesExist {
			// Update it
			k.db[key].Update(time, payload, val)
			log.Println("Overwriting existing key")
			return true
		}
		log.Println("Inserting new key")
		// Use the constructor
		k.db[key] = NewEntry(time, payload, val)
		return true
	}
	log.Println("Invalid entry for key or value")
	return false
}

// Add the server's keys to the clock if they don't already exist
func mergeClocks(client map[string]int, server map[string]int) map[string]int {
	log.Println("Merging clocks. Client: ", client, " Server: ", server)
	if len(server) < 1 {
		return client
	}
	if len(client) < 1 {
		return server
	}
	for k, v := range server {
		if client[k] < v {
			client[k] = v
		}
	}

	return client
}

// GetClock returns the clock associated with a key, it'll return an empty map for one that doesn't exist
func (k *KVS) GetClock(key string) map[string]int {
	// Check to see if we have the key
	if _, ok := k.db[key]; ok {
		return k.db[key].GetClock()
	}
	return map[string]int{}
}

// GetTimestamp returns the timestamp associated witha  key, otherwise an empty struct
func (k *KVS) GetTimestamp(key string) time.Time {
	// Check to see if we have the key
	if _, ok := k.db[key]; ok {
		return k.db[key].GetTimestamp()
	}
	return time.Time{}
}

// OverwriteEntry overwrites the entry associated with the given key using the given entry
func (k *KVS) OverwriteEntry(key string, entry KeyEntry) {
	if entry != nil {
		k.db[key] = entry
	}
}
