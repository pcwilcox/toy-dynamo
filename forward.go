// forward.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson        lelawson
// Pete Wilcox            pcwilcox
//
// This file defines the forwarder struct and the methods that allow it to implement the dbAccess
// interface. This interface allows it to act as the data store for the REST API front end defined
// in app.go. It does this by taking REST requests from the App object and packaging them as new
// requests to an external server defined in the input.
//

package main

// Forwarder is an object which implements the dbAccess interface by wrapping the methods around
// HTTP requests to the server defined in the mainIP field.
type Forwarder struct {
	mainIP string
}

// Contains returns true if the server at mainIP says it has the given key
func (f *Forwarder) Contains(key string) bool {
	return true
}

// Get sends a GET request to the server for the value associated with the given key
func (f *Forwarder) Get(key string) string {
	return "foo"
}

// Delete sends a DELETE request to the server with the given key
func (f *Forwarder) Delete(key string) bool {
	return true
}

// Put sends a PUT request with the given key-value pair to the server
func (f *Forwarder) Put(key, val string) bool {
	return true
}

// ServiceUp pings the server to see if it's up
func (f *Forwarder) ServiceUp() bool {
	return true
}
