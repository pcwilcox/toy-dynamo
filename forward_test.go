// forward_test.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson           lelawson
// Pete Wilcox               pcwilcox
//
// This file contains unit tests for forward.go
//

package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

const (
	keyNotExists = "KEY_NOT_EXISTS"
	valNotExists = "VAL_NOT_EXISTS"
)

type testRest struct {
	server     *httptest.Server
	key        string
	val        string
	keyInvalid bool
	valInvalid bool
}

// Initialize makes a test server for our unit tests to use as a stub
func (t *testRest) Initialize() {

	// We make a local listener and hook a server up to it
	l, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(rootURL+"/{subject}", t.PutHandler).Methods("PUT")
	r.HandleFunc(rootURL+"/{subject}", t.GetHandler).Methods("GET")
	r.HandleFunc(rootURL+"/{subject}", t.DeleteHandler).Methods("DELETE")
	r.HandleFunc("/alive", t.AliveHandler).Methods("GET")

	// Stub the server
	t.server = httptest.NewUnstartedServer(r)
	t.server.Listener.Close()
	t.server.Listener = l
	t.server.Start()
}

// Teardwon simply stops the server we set up earlier
func (t *testRest) Teardown() {
	if t.server != nil {
		t.server.Close()
	}

}

func (t *testRest) PutHandler(w http.ResponseWriter, r *http.Request) {
	if t.keyInvalid || t.valInvalid {
		w.WriteHeader(http.StatusUnprocessableEntity) // code 422
	} else if len(t.key) > 200 || len(t.val) > 1024*1024 {
		w.WriteHeader(http.StatusUnprocessableEntity) // code 422
	} else if strings.Compare(t.key, keyExists) == 0 {
		w.WriteHeader(http.StatusOK) // code 200
	} else {
		w.WriteHeader(http.StatusCreated) // code 201
	}
}

func (t *testRest) GetHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Compare(t.key, keyExists) == 0 {
		w.WriteHeader(http.StatusOK) // code 200
		resp := map[string]interface{}{
			"result": "Success",
			"value":  valExists["val"],
		}
		body, err := json.Marshal(resp)
		if err != nil {
			log.Fatalln("oh no")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)

	} else {
		w.WriteHeader(http.StatusNotFound) // code 404
	}
}

func (t *testRest) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Compare(t.key, keyExists) == 0 {
		w.WriteHeader(http.StatusOK) // code 200
	} else {
		w.WriteHeader(http.StatusNotFound) // code 404
	}
}

func (t *testRest) ServiceDownHandler(http.ResponseWriter, *http.Request) {
	// goes nowhere does nothing
}

func (t *testRest) AliveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// TestForwarderContainsKeyExistsReturnsTrue verifies that the Contains() method returns true if the data store has the key
func TestForwarderContainsKeyExistsReturnsTrue(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's in the db
	s.key = keyExists

	// Here's the actual test
	assert(t, f.Contains(s.key), "Contains() returned false for KeyExists")
}

// TestForwarderContainsKeyNotExistsReturnsFalse verifies that the Contains() method returns false if the data store doesn't have the key
func TestForwarderContainsKeyNotExistsReturnsFalse(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's in the db
	s.key = keyNotExists

	// Here's the actual test
	assert(t, !f.Contains(s.key), "Contains() returned true for keyNotExists")
}

// TestForwarderContainsServiceDownReturnsFalse verifies that the ServiceUp check works
func TestForwarderContainsServiceDownReturnsFalse(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Close the server
	s.Teardown()

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's in the db
	s.key = keyNotExists

	// Here's the actual test
	assert(t, !f.Contains(s.key), "Contains() returned true when service was down")
}

// TestForwarderPutKeyExistsReturnsTrue verifies that the Put() function returns true if it overwrites a key
func TestForwarderPutKeyExistsReturnsTrue(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's in the db
	s.key = keyExists
	s.val = valExists["val"]

	// Here's the actual test
	assert(t, f.Put(s.key, s.val), "Put() keyExists returned false")
}

// TestForwarderPutKeyNotExistsReturnsFalse verifies that the Put() function returns false if it doesn't overwrite a key
func TestForwarderPutKeyNotExistsReturnsFalse(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's in the db
	s.key = keyNotExists
	s.val = valExists["val"]

	// Here's the actual test
	assert(t, !f.Put(s.key, s.val), "Put() keyNotExists returned true")
}

// TestForwarderPutKeyInvalidReturnsFalse verifies that sending an invalid key returns false
func TestForwarderPutKeyInvalidReturnsFalse(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's in the db
	s.key = "a"
	s.val = valExists["val"]

	for i := 1; i < 202; i = i * 2 {
		s.key = s.key + s.key
	}

	// Here's the actual test
	assert(t, !f.Put(s.key, s.val), "Put() invalid key returned true")
}

//TestForwarderPutValInvalidReturnsFalse verifies that sending an invalid value returns false
func TestForwarderPutValInvalidReturnsFalse(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's in the db
	s.key = keyExists
	s.val = "a"

	for i := 1; i < (1024*1024)+2; i *= 2 {
		s.val = s.val + s.val
	}

	// Here's the actual test
	assert(t, !f.Put(s.key, s.val), "Put() invalid value returned true")
}

// TestForwarderPutServiceDownReturnsFalse verifies that attempting a Put() when the service is down fails
func TestForwarderPutServiceDownReturnsFalse(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Close the server
	s.Teardown()

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's in the db
	s.key = keyExists
	s.val = valExists["val"]

	// Here's the actual test
	assert(t, !f.Put(s.key, s.val), "Put() returned true when the service was down")
}

// TestForwarderDeleteKeyExistsReturnsTrue verifies that deleting a key which exists returns true
func TestForwarderDeleteKeyExists(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's in the db
	s.key = keyExists

	// Here's the actual test
	assert(t, f.Delete(s.key), "Delete() returned false for keyExists")
}

// TestForwarderDeleteKeyNotExists checks that Delete() returns false if the key doesn't exist
func TestForwarderDeleteKeyNotExists(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's not in the db
	s.key = keyNotExists

	// Here's the actual test
	assert(t, !f.Delete(s.key), "Delete() returned true for keyNotExists")
}

// TestForwarderDeleteServiceDownReturnsFalse ... returns false if the service is down
func TestForwarderDeleteServiceDownReturnsFalse(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Close the server
	s.Teardown()

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's not in the db
	s.key = keyNotExists

	// Here's the actual test
	assert(t, !f.Delete(s.key), "Delete() returned true when service was down")
}

// TestForwarderGetKeyExistsReturnsVal checks that Get returns the value when given a key that exists
func TestForwarderGetKeyExistsReturnsVal(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's not in the db
	s.key = keyExists
	s.val = valExists["val"]

	// Here's the actual test
	assert(t, strings.Compare(f.Get(s.key), s.val) == 0, "Get() did not return matching string valExists")
}

// TestForwarderGetKeyNotExistsReturnsEmpty verifies that Get returns an empty string if given a key that doesn't exist
func TestForwarderGetKeyNotExistsReturnsEmpty(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's not in the db
	s.key = keyNotExists
	s.val = ""

	// Here's the actual test
	assert(t, strings.Compare(f.Get(s.key), s.val) == 0, "Get() did not return empty string for valNotExists")
}

// TestForwarderGetServiceDownReturnsEmpty checks that Get returns an empty string if the service is down
func TestForwarderGetServiceDownReturnsEmpty(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Close the server
	s.Teardown()

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Use the key that's not in the db
	s.key = keyExists
	s.val = ""

	// Here's the actual test
	assert(t, strings.Compare(f.Get(s.key), s.val) == 0, "Get() did not return empty string when service was down")
}

// TestServiceDownWhenServiceUpReturnsTrue ... it should return true if the service is up
func TestServiceUpWhenServiceUpReturnsTrue(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Here's the actual test
	assert(t, f.ServiceUp(), "ServiceUp() returned false when the service was up")
}

// TestServiceUpWhenServiceDownReturnsFalse ... should return false when service is down
func TestServiceUpWhenServiceDownReturnsFalse(t *testing.T) {
	// Create a stub server
	s := testRest{}
	s.Initialize()
	defer s.Teardown()

	// Get its IP
	IP := s.server.URL

	// Close the server
	s.Teardown()

	// Create the test object
	f := Forwarder{mainIP: IP}

	// Here's the actual test
	assert(t, !f.ServiceUp(), "ServiceUp() returned true when the service was down")
}
