// app.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson     lelawson
// Pete Wilcox         pcwilcox
//
// This is the source file defining the front end of the RESTful API.
//

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// App is a struct to hold the state for the REST API
//
// Initialized with: App app := App{db: {dbAccess}} where db is an object implementing
// the dbAccess interface (see dbAccess.go).
//

// App is a struct representing the externally-accessible state of the data store
type App struct {
	db dbAccess
}

// rootURL is the path prefix for the kvs as in: http://localhost:8080/ROOT_URL/foo
const (
	rootURL = "/keyValue-store"
	port    = ":8080"
)

// Initialize fires up the router and such
func (app *App) Initialize() {

	// Initialize a router
	r := mux.NewRouter()

	// Since all endpoints use the rootURL we just use a subrouter here
	s := r.PathPrefix(rootURL).Subrouter()

	// Each of the request types gets a handler
	s.HandleFunc("/{subject}", app.PutHandler).Methods("PUT")
	s.HandleFunc("/{subject}", app.GetHandler).Methods("GET")
	s.HandleFunc("/{subject}", app.DeleteHandler).Methods("DELETE")

	// Load up the server through a logger interface
	err := http.ListenAndServe(port, handlers.LoggingHandler(os.Stdout, r))
	if err != nil {
		log.Fatalln(err)
	}
}

// PutHandler attempts to put the key:val into the db
func (app *App) PutHandler(w http.ResponseWriter, r *http.Request) {
	// Check to see if the service is up
	if app.db.ServiceUp() {
		// Predefine these so they can be used further down
		var err error
		var val string

		// Safety check - the system will panic trying to read a nil body
		if r.Body != nil {
			// Reads the request body into a []byte
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}

			// Converts the []byte to a string
			s := string(reqBody[:])

			// Remove the prefix
			prefix := "val="
			val = strings.TrimPrefix(s, prefix)
		}

		// Maximum input restrictions
		maxVal := 1048576 // 1 megabyte
		maxKey := 200     // 200 characters

		// This pulls the {subject} out of the URL, that forms the key
		vars := mux.Vars(r)
		key := vars["subject"]

		// Same content type for everything
		w.Header().Set("Content-Type", "application/json")

		// These get written below
		var body []byte
		var status int

		// Check for valid input
		if len(val) > maxVal {
			// The value is > 1MB so error out

			// Set the status code
			status = http.StatusUnprocessableEntity // code 422

			// Form the response into something that JSON can handle
			resp := map[string]interface{}{
				"result": "Error",
				"msg":    "Object too large. Size limit is 1MB",
			}

			// Convert it from a map into a []byte
			body, err = json.Marshal(resp)
			if err != nil {
				// Could try and make this a recoverable error maybe
				log.Fatalln("oh no")
			}
		} else if len(key) > maxKey {
			// The key is more than 200 characters so error out

			// Set the status code
			status = http.StatusUnprocessableEntity // code 422

			// Build the response and shove it into a JSON-[]byte
			resp := map[string]interface{}{
				"msg":    "Key not valid",
				"result": "Error",
			}
			body, err = json.Marshal(resp)
			if err != nil {
				log.Fatalln("oh no")
			}
		} else {
			// key/val are valid inputs, let's insert into the db

			// Check to see if the db already contains the key
			if app.db.Contains(key) {
				// It does so we'll update it
				app.db.Put(key, val)

				// Set status
				status = http.StatusOK // code 200

				// Build the response body
				resp := map[string]interface{}{
					"replaced": "True",
					"msg":      "Updated successfully",
				}
				body, err = json.Marshal(resp)
				if err != nil {
					log.Fatalln("oh no")
				}
			} else {
				// It's a new entry so it gets a different status code
				status = http.StatusCreated // code 201
				app.db.Put(key, val)

				// And a slightly different response body
				resp := map[string]interface{}{
					"replaced": "False",
					"msg":      "Added successfully",
				}
				body, err = json.Marshal(resp)
				if err != nil {
					log.Fatalln("oh no")
				}
			}
		}
		// We assigned status code and body above, write them here
		w.WriteHeader(status)
		w.Write(body)
	} else {
		// The service is down so that gets an entirely different response
		ServiceDownHandler(w, r)
	}
}

// GetHandler gets the corresponding val from the db
func (app *App) GetHandler(w http.ResponseWriter, r *http.Request) {
	// Is the service up?
	if app.db.ServiceUp() {
		// Read the key from the URL
		vars := mux.Vars(r)
		key := vars["subject"]

		// Declare some vars
		var body []byte
		var err error

		// See if the key exists in the db
		if app.db.Contains(key) {
			// It does
			w.WriteHeader(http.StatusOK) // code 200

			// Get the key out of the db
			val := app.db.Get(key)

			// Package it into a map->JSON->[]byte
			resp := map[string]interface{}{
				"result": "Success",
				"value":  val,
			}
			body, err = json.Marshal(resp)
			if err != nil {
				log.Fatalln("oh no")
			}
		} else {
			// The key doesn't exist in the db
			w.WriteHeader(http.StatusNotFound) // code 404

			// Error response
			resp := map[string]interface{}{
				"result": "Error",
				"value":  "Not Found",
			}
			body, err = json.Marshal(resp)
			if err != nil {
				log.Fatalln("oh no")
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	} else {
		// oh no it's down
		ServiceDownHandler(w, r)
	}
}

// DeleteHandler deletes k:v pairs from the db
func (app *App) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	// Is the server up?
	if app.db.ServiceUp() {

		// Get the key from the URL
		vars := mux.Vars(r)
		key := vars["subject"]

		// Declare here, define below
		var err error
		var body []byte

		// Check to see if we've got the key
		if app.db.Contains(key) {
			// We do
			w.WriteHeader(http.StatusOK) // code 200

			// Delete it
			app.db.Delete(key)

			// Successful response
			resp := map[string]interface{}{
				"result": "Success",
			}
			body, err = json.Marshal(resp)
			if err != nil {
				log.Fatalln("oh no")
			}
		} else {

			// We don't have the key
			w.WriteHeader(http.StatusNotFound) // code 404

			// Error response
			resp := map[string]interface{}{
				"result": "Error",
				"msg":    "Status code 404",
			}
			body, err = json.Marshal(resp)
			if err != nil {
				log.Fatalln("oh no")
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	} else {
		// oh no what happened to the server?
		ServiceDownHandler(w, r)
	}
}

// ServiceDownHandler writes to the responseWriter the service down message
func ServiceDownHandler(w http.ResponseWriter, r *http.Request) {
	// This is a weird error code
	w.WriteHeader(http.StatusNotImplemented) // code 501

	// Package up the error response
	resp := map[string]interface{}{
		"result": "Error",
		"msg":    "Server unavailable",
	}
	js, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
