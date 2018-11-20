// dbAccess.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson     lelawson
// Pete Wilcox         pcwilcox
//
// Defines a data store access interface which is used to interact between the REST API front end and the KVS back end.
//

package main

import "time"

// dbAccess interface defines methods for interactions between the REST API front end and the key-value store
type dbAccess interface {

	// Contains returns true if the data store contains an object with key equal to the input
	Contains(string) (bool, int)

	// Get returns the value associated with a particular key. If the key does not exist it returns ""
	Get(string, map[string]int) (string, map[string]int)

	// Delete removes a key-value pair from the object. If the key does not exist it returns false.
	Delete(string, time.Time, map[string]int) bool

	// Put adds a key-value pair to the data store. If the key already exists, then it overwrites the existing value. If the key does not exist then it is added.
	Put(string, string, time.Time, map[string]int) bool

	// Returns an entry's vector clock
	GetClock(string) map[string]int

	// Returns an entry's timestamp
	GetTimestamp(string) time.Time

	// Overwrite the existing entry for this key with the one provided
	OverwriteEntry(string, KeyEntry)

	// Returns a timeGlob struct of all of the keys in the db
	GetTimeGlob() timeGlob

	// Returns an entryGlob struct of all of the keys in the given timeGlob
	GetEntryGlob(timeGlob) entryGlob
}
