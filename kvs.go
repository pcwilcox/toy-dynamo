// KVS.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson            lelawson
// Pete Wilcox                pcwilcox
// Annie Shen                 ashen7
// Victoria Tran              vilatran
//
// This source file defines the KVS object used for a local data store. The struct
// implements a map to store key-value pairs, and implements the dbAccess interface
// defined in dbAccess.go in order to serve as a db object for the REST API.
//

package main

import (
	"log"
	"strings"
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
	Update(string, time.Time, map[string]int, string)

	// Write tombstone
	Delete(string, time.Time, map[string]int)

	// Returns true if it hasn't been tombstone
	Alive() bool

	// Set the version
	SetVersion(int)
}

// Entry is the thing in the KVS and implements all the methods
type Entry struct {
	Version   int            // Monotonically increasing version numbers starting at 1
	Timestamp time.Time      // this is set on writes
	Clock     map[string]int // This is captured from the client payload on write
	Value     string         // This is the actual value
	Tombstone bool           // Tombstone value showing that it was deleted
}

// SetVersion the version
func (e *Entry) SetVersion(v int) {
	if e != nil {
		e.Version = v
	}
}

// NewEntry creates a new entry
func NewEntry(time time.Time, clock map[string]int, val string, version int) *Entry {

	// Make a new entry
	var e Entry

	// All keys start at version 1
	e.Version = version

	// The timestamp is set at creation
	e.Timestamp = time

	// There might be no causal history initially but we still need to create the map
	e.Clock = make(map[string]int)
	if clock != nil && len(clock) > 0 {
		for k, v := range clock {
			e.Clock[k] = v
		}
	}

	// Tombstone obviously should be false
	e.Tombstone = false

	// Finally, set the value
	e.Value = val

	// Return a pointer to the entry
	return &e
}

// GetVersion just returns the version
func (e *Entry) GetVersion() int {
	if e != nil {
		return e.Version
	}
	// The version of a key which doesn't exist is -1
	return 0
}

// GetTimestamp returns the timestamp from the entry
func (e *Entry) GetTimestamp() time.Time {
	if e != nil {
		return e.Timestamp
	}
	// The timestamp of a key which doesn't exist is the empty struct Time{} which evaluates to a ton of 0's
	return time.Time{}
}

// GetClock returns the map representing the causal history
func (e *Entry) GetClock() map[string]int {
	if e != nil {
		return e.Clock
	}
	// The clock of a non-existing key is an empty map
	empty := make(map[string]int)
	return empty
}

// GetValue returns the string stored in the entry
func (e *Entry) GetValue() string {
	if e != nil {
		return e.Value
	}
	// The value of a non-existing key is an empty string
	return ""
}

// Update writes a new value for the entry and updates the clock and version info
func (e *Entry) Update(key string, newTime time.Time, newClock map[string]int, newVal string) {
	log.Println("Updating entry - old version: ", e)
	e.Timestamp = newTime
	e.Value = newVal
	e.Clock = newClock
	e.Tombstone = false
	e.Version++
	e.Clock[key] = e.Version
	log.Println("Updated entry: ", e)

	wakeGossip = true
}

// Delete sets a tombstone that the key has been tombstone
func (e *Entry) Delete(key string, newTime time.Time, payload map[string]int) {
	log.Println("Deleting entry: ", e)
	e.Timestamp = newTime
	e.Value = ""
	e.Clock = payload
	e.Tombstone = true
	e.Version++
	e.Clock[key] = e.Version

	wakeGossip = true
}

// Alive returns true if the key exists and doesn't have a tombstone set
func (e *Entry) Alive() bool {
	if e != nil && e.Tombstone != true {
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
	alive := false
	version := 0
	whoShard := MyShard.GetSuccessor(getKeyPosition(key))
	log.Println("shard of the key is ", whoShard)
	if whoShard == MyShard.PrimaryID() {
		k.mutex.RLock()
		defer k.mutex.RUnlock()

		// Once the read lock has been obtained, call the non-locking contains() method
		alive, version = k.contains(key)
		return alive, version
	}
	log.Println("key not in my shard, requesting id of another shard" + whoShard)
	g := GetRequest{
		Key:     key,
		Payload: nil,
	}
	bobIP, err := MyShard.FindBob(whoShard)
	if err != nil {
		log.Println(err)
		return false, 0
	}
	bobResp, err := sendContainsRequest(bobIP, g)
	if err != nil {
		log.Println(err)
		return false, 0
	}
	alive = bobResp.Alive
	version = bobResp.Version

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
	//Get the key position
	whoShard := MyShard.GetSuccessor(getKeyPosition(key))
	//If it isnt my key then I need to ask the server who it belongs to
	if whoShard != MyShard.PrimaryID() {
		log.Println("key requested to Get not in my shard, requesting id of another shard" + whoShard)
		//Constructing the GetRequest Struct
		GetSend := GetRequest{
			Key:     key,
			Payload: payload,
		}
		//Retrieving bob's IP
		bobIP, err := MyShard.FindBob(whoShard)
		if err != nil {
			log.Println(err)
			return "", map[string]int{}
		}
		//Sending the request and recieving bob's responce
		bobResp, err := sendGetRequest(bobIP, GetSend)
		if err != nil {
			log.Println("Error with sendGetRequest", err)
			panic(err)
		}
		val = bobResp.Value
		clock = bobResp.Payload
		return val, clock
	}
	// Grab a read lock
	log.Println("Checking to see if db contains key ")
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
func (k *KVS) Delete(key string, time time.Time, payload map[string]int) bool {

	log.Println("Attempting to delete key ")
	//Get the key position
	whoShard := MyShard.GetSuccessor(getKeyPosition(key))
	//If it isnt my key then I need to ask the server who it belongs to
	if whoShard != MyShard.PrimaryID() {
		log.Println("Key requested to Delete not in my shard, requesting id of another shard" + whoShard)
		//Constructing the PutRequest Struct
		GetDelete := PutRequest{
			Key:       key,
			Value:     "", //This will not be used
			Timestamp: time,
			Payload:   payload,
		}
		//Retrieving bob's IP
		bobIP, err := MyShard.FindBob(whoShard)
		if err != nil {
			log.Println(err)
			return false
		}
		//Sending the request and recieving bob's responce
		bobResp, err := sendDeleteRequest(bobIP, GetDelete)
		if err != nil {
			log.Println("Error with sendDeleteRequest", err)
			return false
		}
		return bobResp
	}
	// Grab a write lock
	k.mutex.Lock()
	defer k.mutex.Unlock()

	doesExist, _ := k.contains(key)

	// Call the nonlocking contains method
	if doesExist {
		log.Println("Key found, deleting key-value pair")
		k.db[key].Delete(key, time, payload)

		// Initiate Gossip
		wakeGossip = true
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
		//Get the key position
		whoShard := MyShard.GetSuccessor(getKeyPosition(key))
		//If it isnt my key then I need to ask the server who it belongs to
		if whoShard != MyShard.PrimaryID() {
			log.Println("Key requested to Put not in my shard, requesting id of another shard" + whoShard)
			//Constructing the PutRequest Struct
			GetDelete := PutRequest{
				Key:       key,
				Value:     val,
				Timestamp: time,
				Payload:   payload,
			}
			//Retrieving bob's IP
			bobIP, err := MyShard.FindBob(whoShard)
			if err != nil {
				log.Println(err)
				return false
			}
			//Sending the request and recieving bob's responce
			bobResp, err := sendPutRequest(bobIP, GetDelete)
			if err != nil {
				log.Println("Error with sendPutRequest", err)
				return false
			}
			return bobResp
		}
		log.Println("Key and value OK, inserting to DB")
		// Grab a write lock
		k.mutex.Lock()
		defer k.mutex.Unlock()

		doesExist, _ := k.contains(key)

		// Check to see if the key exists
		if doesExist {
			// Update it
			k.db[key].Update(key, time, payload, val)
			log.Println("Overwriting existing key")
			// Initiate Gossip
			wakeGossip = true
			return true
		}
		log.Println("Inserting new key")
		// Use the constructor
		k.db[key] = NewEntry(time, payload, val, 1)
		log.Println("Entered key ", key)
		// Initiate Gossip
		wakeGossip = true
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
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	// Check to see if we have the key
	if _, ok := k.db[key]; ok {
		return k.db[key].GetClock()
	}
	return map[string]int{}
}

// GetTimestamp returns the timestamp associated witha  key, otherwise an empty struct
func (k *KVS) GetTimestamp(key string) time.Time {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	// Check to see if we have the key
	if _, ok := k.db[key]; ok {
		return k.db[key].GetTimestamp()
	}
	return time.Time{}
}

// OverwriteEntry overwrites the entry associated with the given key using the given entry
func (k *KVS) OverwriteEntry(key string, entry KeyEntry) {
	if entry != nil {
		log.Println("Overwriting entry: ", k.db[key])
		k.mutex.Lock()
		defer k.mutex.Unlock()
		k.db[key] = entry
		log.Println("New entry: ", entry)
	}
}

func (k *KVS) overwriteEntry(key string, entry KeyEntry) {
	if entry != nil {
		log.Println("Overwriting entry: ", k.db[key])
		k.db[key] = entry
		log.Println("New entry: ", entry)
	}
}

// GetTimeGlob returns a struct containing a map of keys to their timestamps
func (k *KVS) GetTimeGlob() TimeGlob {
	if k != nil {
		k.mutex.RLock()
		defer k.mutex.RUnlock()
		m := make(map[string]time.Time)
		if k.db != nil {
			for key, v := range k.db {
				m[key] = v.GetTimestamp()
			}
		}
		g := TimeGlob{List: m}

		return g
	}
	return TimeGlob{}
}

// GetEntryGlob returns a struct containing a map of keys to their entries
func (k *KVS) GetEntryGlob(tg TimeGlob) EntryGlob {
	if k != nil {
		k.mutex.RLock()
		defer k.mutex.RUnlock()
		entries := make(map[string]Entry)
		eg := EntryGlob{Keys: entries}
		for n := range tg.List {
			time := k.db[n].GetTimestamp()
			clock := k.db[n].GetClock()
			value := k.db[n].GetValue()
			version := k.db[n].GetVersion()
			tombstone := !k.db[n].Alive()

			log.Println("Building entry for key ", n, " with clock ", clock)
			e := Entry{
				Timestamp: time,
				Clock:     clock,
				Value:     value,
				Version:   version,
				Tombstone: tombstone,
			}
			eg.Keys[n] = e
		}
		log.Println("Built EntryGlob: ", eg)
		return eg
	}
	return EntryGlob{Keys: map[string]Entry{}}
}

// ShuffleKeys checks all the keys, and if it finds some that belong to another server then it sends them over that way
func (k *KVS) ShuffleKeys() {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	shards := strings.Split(MyShard.GetAllShards(), ",")
	var eg EntryGlob
	for _, shard := range shards {
		log.Println("Checking for keys belonging to shard ", shard)
		for key := range k.db {
			eg.Keys = make(map[string]Entry)

			if shard == MyShard.GetSuccessor(getKeyPosition(key)) {
				log.Println("Found key ", key)
				eg.Keys[key] = Entry{
					Version:   k.db[key].GetVersion(),
					Timestamp: k.db[key].GetTimestamp(),
					Clock:     k.db[key].GetClock(),
					Value:     k.db[key].GetValue(),
					Tombstone: k.db[key].Alive(),
				}
				delete(k.db, key)
			}
		}
		if len(eg.Keys) > 0 {
			log.Println("Sending entryglob ", eg, " to bob")
			bob, err := MyShard.FindBob(shard)
			if err != nil {
				log.Println(err)
				continue
			} else {
				log.Println("sending to bob ", bob)
				sendEntryGlob(bob, eg)
			}
		}
	}
	distributeKeys = false
}

// Size returns the number of keys in the DB
func (k *KVS) Size(shardID string) int {
	if k != nil {
		if shardID == MyShard.PrimaryID() {
			return k.mySize()
		}
		bob, err := MyShard.FindBob(shardID)
		if err != nil {
			log.Println(err)
			return 0
		}
		return getBobKeyCount(bob)
	}
	return 0
}

func (k *KVS) mySize() int {
	if k != nil {
		k.mutex.RLock()
		defer k.mutex.RUnlock()
		return len(k.db)
	}
	return 0
}

// ConflictResolution returns true if Bob should update with Alice's key
func (k *KVS) ConflictResolution(key string, aliceEntry KeyEntry) bool {
	if k != nil {
		replace := false
		log.Println("Resolving a conflict")
		k.mutex.Lock()
		defer k.mutex.Unlock()

		isSmaller := false
		isLarger := false
		incomparable := false
		var bMap map[string]int

		var bobEntry Entry

		log.Printf("Comparing Alice's version '%#v'\n", aliceEntry)
		log.Printf("key is ", key)
		aMap := aliceEntry.GetClock()
		_, exists := k.db[key]
		log.Println("key ", key, " exists")
		if exists {
			bMap = k.db[key].GetClock()
			bobEntry.Version = k.db[key].GetVersion()
			bobEntry.Value = k.db[key].GetValue()
			bobEntry.Timestamp = k.db[key].GetTimestamp()
			bobEntry.Tombstone = !k.db[key].Alive()
			bobEntry.Clock = k.db[key].GetClock()
			log.Printf("Bob's version '%#v'\n", bobEntry)
		} else {
			bMap = make(map[string]int)
		}
		log.Println("aMap: ", aMap)
		log.Println("bMap: ", bMap)

		// if bob does NOT have the key, we definitely update w/ Alice's stuff
		if len(bMap) == 0 {
			log.Println("Bob doesn't have the entry: ", key)
			replace = true // Bob can't possibly beat Alice's key with no corresponding key of it's own
		} else {
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
				if aliceEntry.GetTimestamp().After(k.db[key].GetTimestamp()) {
					log.Println("Alice wins with the later timestamp")
					replace = true // alice wins
				}
				log.Println("Bob wins with a later timestamp")
			} else if isSmaller == false && isLarger == true {
				log.Println("Alice wins with a larger clock")
				replace = true // alice wins
			} else {
				log.Println("Bob wins with a larger clock")
			}
		}
		if replace {
			k.overwriteEntry(key, aliceEntry)
		}
	}
	return false
}
