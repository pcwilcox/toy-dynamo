// KVS.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson            lelawson
// Pete Wilcox                pcwilcox
//
// This source file defines the KVS object used for a local data store. The struct
// implements a map to store key-value pairs, and implements the dbAccess interface
// defined in dbAccess.go in order to serve as a db object for the REST API.
//

package main

import (
	"log"
	"sync"
)

// KVS represents a key-value store and implements the dbAccess interface
type KVS struct {
	db    map[string]string
	mutex *sync.RWMutex
}

// NewKVS initializes a KVS object and returns a pointer to it function
func NewKVS() *KVS {
	var k KVS
	k.db = make(map[string]string)
	var m sync.RWMutex
	k.mutex = &m
	return &k
}

// Contains returns true if the dbAccess object contains an object with key equal to the input
func (k *KVS) Contains(key string) bool {
	log.Println("Checking to see if db contains key " + key)
	// Grab a read lock
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	// Once the read lock has been obtained, call the non-locking contains() method
	ok := k.contains(key)
	return ok
}

// contains is the unexported version of Contains() and does not hold a read lock
func (k *KVS) contains(key string) bool {
	_, ok := k.db[key]
	return ok
}

// Get returns the value associated with a particular key. If the key does not exist it returns ""
func (k *KVS) Get(key string) string {
	log.Println("Getting value associated with key " + key)
	// Grab a read lock
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	// Call the non-locking contains() method
	if k.contains(key) {
		log.Println("Value found: " + k.db[key])
		return k.db[key]
	}
	log.Println("Value not found")
	return ""
}

// Delete removes a key-value pair from the object. If the key does not exist it returns false.
func (k *KVS) Delete(key string) bool {
	log.Println("Attempting to delete key " + key)
	// Grab a write lock
	k.mutex.Lock()
	defer k.mutex.Unlock()

	// Call the nonlocking contains method
	if k.contains(key) {
		log.Println("Key found, deleting key-value pair")
		delete(k.db, key)
		return true
	}
	log.Println("Key not found")
	return false

}

// Put adds a key-value pair to the DB. If the key already exists, then it overwrites the existing value. If the key does not exist then it is added.
func (k *KVS) Put(key string, val string) bool {
	maxVal := 1048576 // 1 megabyte
	maxKey := 200     // 200 characters
	keyLen := len(key)
	valLen := len(val)

	log.Println("Attempting to insert key " + key + " with value " + val)

	if keyLen <= maxKey && valLen <= maxVal {
		log.Println("Key and value OK, inserting to DB")
		// Grab a write lock
		k.mutex.Lock()
		defer k.mutex.Unlock()
		if k.contains(key) {
			k.db[key] = val
			log.Println("Overwriting existing key")
			return true
		}
		log.Println("Inserting new key")
		k.db[key] = val
		return true
	}
	log.Println("Invalid entry for key or value")
	return false
}

// ServiceUp returns true if the interface is able to communicate with the DB
func (k *KVS) ServiceUp() bool {
	return true
}
