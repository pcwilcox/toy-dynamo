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

import "net/http"

// Restful is an interface containing methods for a REST API for interacting with a key-value data store
type Restful interface {
	Initialize()
	PutHandler(http.ResponseWriter, *http.Request)
	GetHandler(http.ResponseWriter, *http.Request)
	DeleteHandler(http.ResponseWriter, *http.Request)
	ServiceDownHandler(http.ResponseWriter, *http.Request)
}
