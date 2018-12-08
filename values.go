// values.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson         lelawson
// Pete Wilcox             pcwilcox
// Annie Shen              ashen7
// Victoria Tran           vilatran
//
// Contains constant values used elsewhere in the program
//

package main

const (
	// These control the REST API
	rootURL    = "/keyValue-store" // We hang the router off this
	port       = ":8080"           // This is used for the TCP module
	searchURL  = "/search"
	viewURL    = "/view"
	shardURL   = "/shard"
	membersURL = "/members"
	countURL   = "/count"

	changeNum   = "/changeShardNumber"
	keySuffix   = "/{subject}"
	shardSuffix = "/{shard-id}"
	myID        = "/my_id"
	allID       = "/all_ids"

	// Maximum input restrictions
	maxVal = 1048576 // 1 megabyte
	maxKey = 200     // 200 characters

	// These are for unit tests
	keyExists    = "KEY_EXISTS"
	keyNotExists = "KEY_DOESN'T_EXIST"
	valExists    = "VALUE_EXISTS"
	valNotExists = "VALUE_DOESN'T_EXIST"
	keyone       = "Key One"
	valone       = "Value One"
	keyNotHere   = "Key Not Here"
	valtwo       = "Value Two"
	deletekey    = "I DONT EXIST"
	invalidVal   = ""
	testView     = "176.32.164.10:8082,176.32.164.10:8083,176.32.164.10:8084"
	testView2    = "176.32.164.10:8080,176.32.164.10:8081,176.32.164.10:8082,176.32.164.10:8083,176.32.164.10:8084,176.32.164.10:8085"
	testMain     = "176.32.164.10:8082"
	// testShard	 =
	viewExist    = "176.32.164.10:8083"
	viewNotExist = "176.32.164.10:8085"
	invalidKey   = `Lorem ipsum dolor sit amet, 
			consectetuer adipiscing elit. Aenean commodo 
			ligula eget dolor. Aenean massa. Cum sociis natoque 
			penatibus et magnis dis parturient montes, 
			nascetur ridiculus mus. Donec qu`
)

// These values are used throughout the app and are initially set in main
var myIP string           // set as environment variable IP_PORT
var wakeGossip bool       // If true, we wake up during the heartbeat loop
var needHelp bool         // If this is true, we haven't heard anything in a while
var ringSize = 1000       // The number of positions on the ring
var numVirtualNodes = 10  // The number of virtual nodes per shard
var shardChange bool      // If this si true, we need to communicate a shard transition
var distributeKeys bool   // If this is true we need to check and distribute any keys
var defaultShardCount = 1 // if it isn't specified

// These are shard names
var shardNames = []string{"Mariano", "Vada", "Sunshine", "Iris", "Vicki", "Clelia", "Lizette", "Isaura", "Marry", "Dorthy", "Emelina", "Karyn", "Coletta", "Denese", "Mose", "Vaughn", "Spring", "Tyron", "Anthony", "September", "Cecile", "Adelina", "Lacie", "Yoshiko", "Joshua", "Theron", "Lorenzo", "Mathilda", "Dudley", "Shanell", "Etta", "Elfrieda", "Precious", "Mitzie", "Benito", "Starla", "Raphael", "Marianne", "Corina", "Abbey", "Athena", "Cathey", "Zulema", "Cathleen", "Retta", "Rena", "Hope", "Alana", "Keva", "Lacy"}
