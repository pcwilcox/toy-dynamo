// KVS.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson            lelawson
// Pete Wilcox                pcwilcox
// Annie Shen                 ashen7
// Victoria Tran              vilatran
//
// This file defines structures implemented within kvs.go and tcp.go
// in order to communicate REST requests between servers

package main

import "time"

// ShardGlob contains the only really important information we need to transmit
type ShardGlob struct {
	ShardList map[string][]string
}

// TimeGlob is a map of keys to timestamps and lets the gossip module figure out which ones need to be updated
type TimeGlob struct {
	List map[string]time.Time
}

// EntryGlob is a map of keys to entries which allowes the gossip module to enter into conflict resolution and update the required keys
type EntryGlob struct {
	Keys map[string]Entry
}

// GetRequest is sent from Alice to Bob when Alice receives a request
// for a key belonging to Bob's shard.
type GetRequest struct {
	Key     string
	Payload map[string]int
}

// PutRequest is sent from Alice to Bob when Alice receives a write for
// a key belonging to Bob
type PutRequest struct {
	Key       string
	Value     string
	Payload   map[string]int
	Timestamp time.Time
}

// GetResponse is sent from Bob back to Alice
type GetResponse struct {
	Value   string
	Payload map[string]int
}

// ContainsResponse is sent back and contains the information returned by Contains.
type ContainsResponse struct {
	Alive   bool
	Version int
}
