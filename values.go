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
	rootURL   = "/keyValue-store" // We hang the router off this
	port      = ":8080"           // This is used for the TCP module
	search    = "/search"
	view      = "/view"
	keySuffix = "/{subject}"

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
	testMain     = "176.32.164.10:8082"
	viewExist    = "176.32.164.10:8083"
	viewNotExist = "176.32.164.10:8085"
	invalidKey   = `Lorem ipsum dolor sit amet, 
			consectetuer adipiscing elit. Aenean commodo 
			ligula eget dolor. Aenean massa. Cum sociis natoque 
			penatibus et magnis dis parturient montes, 
			nascetur ridiculus mus. Donec qu`
)

// These values are used throughout the app and are initially set in main
var myIP string     // set as environment variable IP_PORT
var wakeGossip bool // If true, we wake up during the heartbeat loop
var needHelp bool   // If this is true, we haven't heard anything in a while
var viewChange bool // If this is true, we need to communicate a view change
