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

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Forwarder is an object which implements the dbAccess interface by wrapping the methods around
// HTTP requests to the server defined in the mainIP field.
type Forwarder struct {
	mainIP string
}

const (
	rootURL = "/keyValue-store"
)

// Contains returns true if the server at mainIP says it has the given key
func (f *Forwarder) Contains(key string) bool {
	if f.ServiceUp() && key != "" {
		// Assemble the URL for the request
		URL := f.mainIP + rootURL + "/" + key

		// Make a GET request to the URL
		resp, err := http.Get(URL)
		if err != nil {
			return false
		}

		// Make sure we close the body
		defer resp.Body.Close()

		// Read the status code
		if resp.StatusCode == http.StatusOK {
			// Code 200 means the key exists
			return true
		}
		// Any other code means it doesn't
		return false
	}
	// The input key is an empty string or the service is down
	return false
}

// Get sends a GET request to the server for the value associated with the given key
func (f *Forwarder) Get(key string) string {
	if f.ServiceUp() && key != "" {
		// Assemble a URL for the request
		URL := f.mainIP + rootURL + "/" + key

		// Make a GET request to the URL
		resp, err := http.Get(URL)
		if err != nil {
			return ""
		}

		// Make sure we close the body
		defer resp.Body.Close()

		// Read the body into a []byte
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return ""
		}

		// Make a map to unload the response into
		var data map[string]string

		// Read the message body and put it into the map
		err = json.Unmarshal(body, &data)
		if err != nil {
			return ""
		}

		// Get the value out of the map
		return data["value"]
	}
	// Service is down or we have invalid input
	return ""
}

// Delete sends a DELETE request to the server with the given key
func (f *Forwarder) Delete(key string) bool {
	if f.ServiceUp() && key != "" {
		// Assemble the URL
		URL := f.mainIP + rootURL + "/" + key

		// Request body needs to be an io.Reader
		var body io.Reader

		// Go's http library doesn't have a handy request method for DELETE
		req, err := http.NewRequest(http.MethodDelete, URL, body)
		if err != nil {
			return false
		}

		// Need a client in order to make the request
		client := http.DefaultClient

		// Now we can actually do the request
		resp, err := client.Do(req)
		if err != nil {
			return false
		}

		// Make sure we close the body
		defer resp.Body.Close()

		// Read the status code
		if resp.StatusCode == http.StatusOK {
			// Code 200 means the key was deleted
			return true
		}
		// Any other code means it wasn't
		return false

	}
	// Service is down or the key was empty
	return false
}

// Put sends a PUT request with the given key-value pair to the server
func (f *Forwarder) Put(key, val string) bool {
	if f.ServiceUp() && key != "" {
		// Make the URL
		URL := f.mainIP + rootURL + "/" + key

		// Request body needs to be an io.Reader
		var body io.Reader
		body = strings.NewReader("val=" + val)

		// Go's http library doesn't have a handy request method for PUT
		req, err := http.NewRequest(http.MethodPut, URL, body)
		if err != nil {
			return false
		}

		// Need a client in order to make the request
		client := http.DefaultClient

		// Now we can actually do the request
		resp, err := client.Do(req)
		if err != nil {
			return false
		}

		// Make sure we close the body
		defer resp.Body.Close()

		// Read the status code
		if resp.StatusCode == http.StatusOK {
			// Code 200 means the value was updated
			return true
		}
		// Any other code means it wasn't
		return false

	}
	// Service was down or the key was empty
	return false
}

// ServiceUp pings the server to see if it's up
func (f *Forwarder) ServiceUp() bool {
	// Make the URL
	URL := f.mainIP + "/alive"

	// Do the request
	resp, err := http.Get(URL)
	if err != nil {
		return false
	}

	// Make sure we close the body
	defer resp.Body.Close()

	// Read the status code
	if resp.StatusCode == http.StatusOK {
		// Code 200 means the service is up
		return true
	}
	// Any other code means it isn't
	return false

}
