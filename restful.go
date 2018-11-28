// restful.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson         lewlawson
// Pete Wilcox             pcwilcox
// Annie Shen      		   ashen7
// Victoria Tran  		   vilatran
//
// This file defines the Restful interface implemented by the RESTful app.
//

package main

import "net/http"

// Restful is an interface containing methods for a REST API for interacting with a key-value data store
type Restful interface {
	// Initialize starts the service
	Initialize()

	// PutHandler responds to PUT requests by adding valid key-value pairs to the data store
	PutHandler(http.ResponseWriter, *http.Request)

	// GetHandler responds to GET requests by reading for a key and returning the value
	GetHandler(http.ResponseWriter, *http.Request)

	// DeleteHandler responds to DELETE requests by removing matching key-value pairs from the data store
	DeleteHandler(http.ResponseWriter, *http.Request)

	// ViewHandlers responds to /view view change requests of Put, Get, and Delete
	ViewPutHandler(http.ResponseWriter, *http.Request)
	ViewGetHandler(http.ResponseWriter, *http.Request)
	ViewDeleteHandler(http.ResponseWriter, *http.Request)
}
