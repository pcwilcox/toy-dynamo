// shard.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson  lelawson
// Pete Wilcox      pcwilcox
// Annie Shen       ashen7
// Victoria Tran    vilatran
//
// Defines an interface and struct for the sharding system.
//

package main

import (
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"

	"github.com/go-test/deep"
)

// Shard interface defines the interactions with the shard system
type Shard interface {
	// Returns the number of elemetns in the shard map
	CountShards() int

	// Return number of servers
	CountServers() int

	// Return true if we contain the server
	ContainsServer(string) bool

	// Return true if we contain the shard
	ContainsShard(string) bool

	// Deletes a shard ID from the shard list
	Remove(string) bool

	// Inserts an shard ID into the shard list
	Add(string) bool

	// Returns the actual shard ID I am in
	PrimaryID() string

	// Return our IP
	GetIP() string

	// Converts the shard IDs and servers in the ID into a comma-separated string
	String() string

	// Return a random number of elements from the local view
	RandomLocal(int) []string

	// Return a random number of elements from the global view
	RandomGlobal(int) []string

	// FindBob returns a random element of a particular shard
	FindBob(string) string

	// Overwrite with a new view of the world
	Overwrite(ShardGlob)

	// GetShardGlob returns a Shard object
	GetShardGlob() ShardGlob

	// GetAllShards returns a comma separated string of all shard ids
	GetAllShards() string

	// GetMembers returns a comma separated string of all member servers addresses
	GetMembers(string) string
}

// ShardList is a struct which implements the Shard interface and holds shard ID system of servers
/*
ShardList: {
	ShardStrings: {
        A: "192.168.0.10:8081,192.168.0.10:8082",
        B: "192.168.0.10:8083,192.168.0.10:8084",
    },
    ShardSlices: {
        A: ["192.168.0.10:8081", "192.168.0.10:8082"],
        B: ["192.168.0.10:8083", "192.168.0.10:8084"],
    },
    PrimaryShard: "A",
    PrimaryIP: "192.168.0.10:8081",
}
*/
type ShardList struct {
	ShardString  map[string]string   // This is the map of shard IDs to server names
	ShardSlice   map[string][]string // this is a mapping of shard IDs to slices of server strings
	PrimaryShard string              // This is the shard ID I belong in
	PrimaryIP    string              // this is my IP
	Tree         *RBTree             // This is our red-black tree holding the shard positions on the ring
	Size         int                 // total number of servers
	NumShards    int                 // total number of shards
	Mutex        *sync.RWMutex       // lock for the whole thing
}

// GetSuccessor asks the tree for the successor
func (s *ShardList) GetSuccessor(k int) string {
	if s != nil {
		log.Println("Finding successor of position ", k)
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		return s.Tree.successor(k)
	}
	return ""
}

// GetAllShards returns a comma-separated list of shards
func (s *ShardList) GetAllShards() string {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		var sl []string
		for k := range s.ShardString {
			sl = append(sl, k)
		}
		st := strings.Join(sl, ", ")
		return st
	}
	return ""
}

// GetMembers returns the members of one shard
func (s *ShardList) GetMembers(shard string) string {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		return s.ShardString[shard]
	}
	return ""
}

// FindBob returns a random element of the chosen shard
func (s *ShardList) FindBob(shard string) string {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		log.Println("Finding a Bob in shard ", shard)
		r := rand.Int()
		l := s.ShardSlice[shard]
		i := r % len(l)
		bob := l[i]
		log.Println("Found Bob: ", bob)
		return bob
	}
	return ""
}

// GetShardGlob returns a ShardGlob
func (s *ShardList) GetShardGlob() ShardGlob {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		g := ShardGlob{ShardList: s.ShardSlice}
		return g
	}
	return ShardGlob{}
}

// Overwrite overwrites our view of the world with another
func (s *ShardList) Overwrite(sg ShardGlob) {
	if s != nil {
		log.Println("attempting to overwrite")
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		s.overwrite(sg)
	}
}

func (s *ShardList) overwrite(sg ShardGlob) {
	if s != nil {
		if diff := deep.Equal(sg.ShardList, s.ShardSlice); diff != nil {
			log.Println(diff)
			// Remove our old view of the world
			for k := range s.ShardSlice {
				log.Println("deleting id ", k)
				delete(s.ShardSlice, k)
				delete(s.ShardString, k)
				for _, i := range getVirtualNodePositions(k) {
					s.Tree.delete(i)
				}
			}

			// Write the new one
			for k, v := range sg.ShardList {
				// Directly transfer the slices over
				s.ShardSlice[k] = v
				log.Println("shard id: ", k, " servers: ", s.ShardSlice[k])

				// Join the slices to form the string
				s.ShardString[k] = strings.Join(v, ",")
				log.Println("shard id: ", k, " servers: ", s.ShardString[k])

				// Check which shard we're in
				for i := range v {
					if v[i] == s.PrimaryIP {
						s.PrimaryShard = k
						log.Println("This server's new ID is ", k)
					}
				}

				// rebuild the tree
				for _, i := range getVirtualNodePositions(k) {
					s.Tree.put(i, k)
				}
			}

			shardChange = true
		}
	}

}

// RandomGlobal returns a random selection of other servers from any shard
func (s *ShardList) RandomGlobal(n int) []string {
	if s != nil {
		log.Println("Finding ", n, " random global servers")
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		var t []string

		if n > s.Size-1 {
			n = s.Size - 1
		}

		for _, v := range s.ShardSlice {
			r := rand.Int() % len(v)
			if v[r] == s.PrimaryIP {
				continue
			}
			t = append(t, v[r])
			if len(t) >= n {
				break
			}
		}

		log.Println("found servers: ", t)
		return t
	}
	return []string{}
}

// RandomLocal returns a random selection of other servers from within our own shard
func (s *ShardList) RandomLocal(n int) []string {
	if s != nil {
		log.Println("Finding ", n, " random local servers")
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		var t []string

		l := s.ShardSlice[s.PrimaryID()]
		if n > len(l)-1 {
			n = len(l) - 1
		}

		for len(t) < n {
			r := rand.Int() % len(l)
			if l[r] == s.PrimaryIP {
				continue
			}
			t = append(t, l[r])
			if len(t) >= n {
				break
			}
		}

		log.Println("found servers: ", t)
		return t
	}
	return []string{}
}

// CountServers returns the number of servers in the shard map
func (s *ShardList) CountServers() int {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		return s.Size
	}
	return 0
}

// CountShards returns the number of shards
func (s *ShardList) CountShards() int {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		return s.NumShards
	}
	return 0
}

// ContainsShard returns true if the ShardList contains a given shardID
func (s *ShardList) ContainsShard(shardID string) bool {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		sp := strings.TrimSpace(shardID)
		log.Println("checking for shard ", sp)
		for k := range s.ShardSlice {
			match := k == sp
			if match {
				log.Println(k, " is a match for input", sp)
			}
		}
		_, ok := s.ShardSlice[sp]
		return ok
	}
	return false
}

// ContainsServer checks to see if the server exists
func (s *ShardList) ContainsServer(ip string) bool {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		for _, v := range s.ShardSlice {
			for _, i := range v {
				if i == ip {
					return true
				}
			}
		}
	}
	return false
}

// Remove deletes a shard ID from the shard list
func (s *ShardList) Remove(shardID string) bool {
	if s != nil {
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		delete(s.ShardString, shardID)
		delete(s.ShardSlice, shardID)
		s.NumShards--
		shardChange = true
		return true
	}
	return false
}

// Add inserts an shard ID into the my shard list
func (s *ShardList) Add(newShardID string) bool {
	if s != nil {
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		// QUESTION: is here where I choose the random name, or the caller?
		// Insert newShardID into both maps
		s.ShardString[newShardID] = ""
		s.ShardSlice[newShardID] = append(s.ShardSlice[newShardID], "")
		s.NumShards++
		shardChange = true
		return true
	}
	return false
}

// AddServer adds a server to the view
func (s *ShardList) AddServer(newServer string) bool {
	if s != nil {
		s.Mutex.Lock()
		defer s.Mutex.Unlock()

		st := s.string()
		sp := strings.Split(st, ",")
		sp = append(sp, newServer)
		sort.Strings(sp)
		sg := MakeGlob(sp, s.NumShards)
		s.overwrite(sg)
		return true
	}
	return false
}

// RemoveServer removes a server from the view
func (s *ShardList) RemoveServer(server string) bool {
	if s != nil {
		s.Mutex.Lock()
		defer s.Mutex.Unlock()

		st := s.string()
		sp := strings.Split(st, ",")
		newSlice := []string{}
		for _, v := range sp {
			if v == server {
				continue
			}
			newSlice = append(newSlice, v)
		}
		sort.Strings(newSlice)
		sg := MakeGlob(newSlice, s.NumShards)
		s.overwrite(sg)
		return true
	}
	return false
}

// MakeGlob takes a slice of servers and makes a glob
func MakeGlob(servers []string, shards int) ShardGlob {
	sort.Strings(servers)
	// We'll make a new map for them
	newMap := make(map[string][]string)

	// iterate over the servers
	for i := 0; i < len(servers); i++ {
		// index them into the map, mod the number of shards
		shardIndex := i % shards

		// the shard id is the index into the name list
		name := shardNames[shardIndex]

		// append them to the list
		newMap[name] = append(newMap[name], servers[i])
	}
	sg := ShardGlob{newMap}
	return sg
}

// PrimaryID returns the actual shard ID I am in
func (s *ShardList) PrimaryID() string {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		return s.PrimaryShard
	}
	return ""
}

// GetIP returns my IP
func (s *ShardList) GetIP() string {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		return s.PrimaryIP
	}
	return ""
}

// NumLeftoverServers returns the number of leftover servers after an uneven spread
func (s *ShardList) NumLeftoverServers() int {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()
		return s.Size % s.NumShards
	}
	return -1
}

// String returns a comma-separated string of servers
func (s *ShardList) String() string {
	if s != nil {
		s.Mutex.RLock()
		defer s.Mutex.RUnlock()

		return s.string()
	}
	return ""
}

func (s *ShardList) string() string {
	if s != nil {
		str := []string{}

		for _, v := range s.ShardSlice {
			for _, t := range v {
				str = append(str, t)
			}
		}
		sort.Strings(str)
		j := strings.Join(str, ",")
		return j
	}
	return ""
}

// NumServerPerShard returns number of servers per shard (equally) after reshuffle
func (s *ShardList) NumServerPerShard() int {
	if s != nil {
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		i := s.Size / s.NumShards
		if i >= 2 {
			return i
		}
	}
	// The caller function needs to send response to client. Insufficent shard number!!
	return -1
}

// NewShard creates a shardlist object and initializes it with the input string
func NewShard(primaryIP string, globalView string, numShards int) *ShardList {
	log.Println("Making new ShardList")
	log.Println("primaryIP : ", primaryIP)
	log.Println("globalView: ", globalView)
	log.Println("numShards : ", numShards)
	// init fields
	shardSlice := make(map[string][]string)
	shardString := make(map[string]string)
	var r RBTree
	var m sync.RWMutex
	s := ShardList{
		ShardSlice:  shardSlice,
		ShardString: shardString,
		Tree:        &r,
		NumShards:   numShards,
		PrimaryIP:   primaryIP,
		Mutex:       &m,
	}

	// take the view and split it into individual server IPs
	sp := strings.Split(globalView, ",")

	// assign size
	s.Size = len(sp)

	// sort them
	sort.Strings(sp)

	// take our list of shard names and sort it
	sort.Strings(shardNames)

	// iterate over the servers
	for i := 0; i < len(sp); i++ {
		// index them into the map, mod the number of shards
		shardIndex := i % numShards

		// the shard id is the index into the name list
		name := shardNames[shardIndex]

		// append them to the list
		s.ShardSlice[name] = append(s.ShardSlice[name], sp[i])

		// check if this particular server is us and assign our shard id
		if sp[i] == s.PrimaryIP {
			s.PrimaryShard = name
		}
	}

	// now insert them into the joined version
	for k, v := range s.ShardSlice {
		s.ShardString[k] = strings.Join(v, ",")
	}

	for k := range s.ShardSlice {
		for _, i := range getVirtualNodePositions(k) {
			s.Tree.put(i, k)
		}
	}

	log.Println("ShardString: ", s.ShardString)
	log.Println("RBTree: ", s.Tree)
	log.Println("Tree root: ", s.Tree.Root)
	log.Println("Size: ", s.Size)

	return &s
}

// ChangeShardNumber is called by the REST API
// Returns true if the change is legal, false otherwise
func (s *ShardList) ChangeShardNumber(n int) bool {
	if s != nil {
		log.Println("Attempting to change shard count to ", n)
		s.Mutex.Lock()
		defer s.Mutex.Unlock()
		if s.Size/n < 2 {
			log.Println("Not enough nodes")
			return false
		}

		// Get our list of servers
		str := s.string()
		log.Println("Current servers: ", str)
		sl := strings.Split(str, ",")
		log.Println("as a slice: ", sl)
		sort.Strings(sl)
		log.Println("sorted: ", sl)

		sg := MakeGlob(sl, n)

		s.overwrite(sg)

		return true
	}
	return false
}
