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
	domain        = "http://localhost"
	rootURL       = "/keyValue-store"
	port          = ":8080"
	hostname      = domain + port + rootURL
	keyExists     = "KEY_EXISTS"
	keyNotExists  = "KEY_DOESN'T_EXIST"
	valExists     = "VALUE_EXISTS"
	valNotExists  = "VALUE_DOESN'T_EXIST"
	listenAddress = "127.0.0.1:8080"
	keySuffix     = "/{subject}"
	alive         = "/alive"
	search        = "/search"
	view          = "/view"
	keyone        = "Key One"
	valone        = "Value One"
	keyNotHere    = "Key Not Here"
	valtwo        = "Value Two"
	deletekey     = "I DONT EXIST"
	invalidVal    = ""
	testView      = "176.32.164.10:8082,176.32.164.10:8083,176.32.164.10:8084"
	testMain      = "176.32.164.10:8082"
	viewExist     = "176.32.164.10:8083"
	viewNotExist  = "176.32.164.10:8085"
	invalidKey    = `Lorem ipsum dolor sit amet, 
			consectetuer adipiscing elit. Aenean commodo 
			ligula eget dolor. Aenean massa. Cum sociis natoque 
			penatibus et magnis dis parturient montes, 
			nascetur ridiculus mus. Donec qu`
)

var myIP string
var wakeGossip bool
