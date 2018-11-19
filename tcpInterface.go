// restful.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson         lewlawson
// Pete Wilcox             pcwilcox
//
// This file defines the Restful interface implemented by the RESTful app.
//

package main

// Restful is an interface containing methods for a REST API for interacting with a key-value data store
type tcpInterface interface {
	init()

	Listen() error

	sendEntryGlob(ip string, eg entryGlob) error

	sendTimeGlob(ip string, tg timeGlob) (*timeGlob, error)

	server() error
}
