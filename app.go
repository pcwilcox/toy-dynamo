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
	"log"
	"net/http"

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

// Initialize fires up the router and such
func (app *App) Initialize() {

	// Initialize a router
	r := mux.NewRouter()

	// Since all endpoints use the rootURL we just use a subrouter here
	s := r.PathPrefix(rootURL).Subrouter()

	// This is the search handler, which has a different prefix
	s.HandleFunc(search+keySuffix, app.SearchHandler).Methods(http.MethodGet)

	// Each of the request types gets a handler
	s.HandleFunc(keySuffix, app.PutHandler).Methods(http.MethodPut)
	s.HandleFunc(keySuffix, app.GetHandler).Methods(http.MethodGet)
	s.HandleFunc(keySuffix, app.DeleteHandler).Methods(http.MethodDelete)

	log.Println("App initialized")
	// Load up the server through a logger interface
	err := http.ListenAndServe(port, handlers.LoggingHandler(MultiLogOutput, r))
	if err != nil {
		log.Fatalln(err)
	}
}

// PutHandler attempts to put the key:val into the db
func (app *App) PutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling PUT request")

	// Predefine these so they can be used further down
	var err error
	var value string

	// These get written below
	var body []byte
	var status int

	// Same content type for everything
	w.Header().Set("Content-Type", "application/json")

	// Safety check - the system will panic trying to read a nil body
	if r.Body != nil {
		// Parse the form so we can read values
		r.ParseForm()

		if len(r.Form) > 0 {
			value = r.Form["val"][0]

			// Maximum input restrictions
			maxVal := 1048576 // 1 megabyte
			maxKey := 200     // 200 characters

			// This pulls the {subject} out of the URL, that forms the key
			vars := mux.Vars(r)
			key := vars["subject"]

			// Check for valid input
			if len(value) > maxVal {
				// The value is > 1MB so error out
				log.Println("ERROR: Value length too long")

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
					log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
				}
			} else if len(key) > maxKey {
				// The key is more than 200 characters so error out
				log.Println("ERROR: Key length too long")

				// Set the status code
				status = http.StatusUnprocessableEntity // code 422

				// Build the response and shove it into a JSON-[]byte
				resp := map[string]interface{}{
					"msg":   "Error",
					"error": "Key not valid",
				}
				body, err = json.Marshal(resp)
				if err != nil {
					log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
				}
			} else {
				// key/val are valid inputs, let's insert into the db
				log.Println("Key and value lengths ok")

				// Check to see if the db already contains the key
				if app.db.Contains(key) {
					// It does so we'll update it
					log.Println("Key already exists in DB, overwriting...")
					app.db.Put(key, value)

					log.Printf("Inserted key-value pair")
					// Set status
					status = http.StatusOK // code 200

					// Build the response body
					resp := map[string]interface{}{
						"replaced": true,
						"msg":      "Updated successfully",
					}
					body, err = json.Marshal(resp)
					if err != nil {
						log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
					}
				} else {
					log.Println("Key does not exist in DB, inserting...")
					log.Printf("Inserted key-value pair")
					// It's a new entry so it gets a different status code
					status = http.StatusCreated // code 201
					app.db.Put(key, value)

					// And a slightly different response body
					resp := map[string]interface{}{
						"replaced": false,
						"msg":      "Added successfully",
					}
					body, err = json.Marshal(resp)
					if err != nil {
						log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
					}
				}
			}

		} else {
			log.Println("ERROR: No data sent with request")

			// There's no body in the request
			status = http.StatusNotFound // code 404

			// And a slightly different response body
			resp := map[string]interface{}{
				"msg":   "Error",
				"error": "Value is missing",
			}
			body, err = json.Marshal(resp)
			if err != nil {
				log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
			}
		}
	}
	// We assigned status code and body above, write them here
	w.WriteHeader(status)
	w.Write(body)
}

// GetHandler gets the corresponding val from the db
func (app *App) GetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling GET request")

	// Read the key from the URL
	vars := mux.Vars(r)
	key := vars["subject"]

	// Declare some vars
	var body []byte
	var err error

	// Same content type for everything
	w.Header().Set("Content-Type", "application/json")

	// See if the key exists in the db
	if app.db.Contains(key) {
		// It does
		w.WriteHeader(http.StatusOK) // code 200

		// Get the key out of the db
		val := app.db.Get(key)
		log.Println("Key found in DB")

		// Package it into a map->JSON->[]byte
		resp := map[string]interface{}{
			"msg":   "Success",
			"value": val,
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}
	} else {
		log.Println("Key not found in DB")
		// The key doesn't exist in the db
		w.WriteHeader(http.StatusNotFound) // code 404

		// Error response
		resp := map[string]interface{}{
			"msg":   "Error",
			"error": "Key does not exist",
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}
	}
	w.Write(body)
}

// SearchHandler checks if the db contains the key
func (app *App) SearchHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling SEARCH request")

	// Read the key from the URL
	vars := mux.Vars(r)
	key := vars["subject"]

	// Declare some vars
	var body []byte
	var err error

	// Same content type for everything
	w.Header().Set("Content-Type", "application/json")

	// See if the key exists in the db
	if app.db.Contains(key) {
		log.Println("Key found in DB")
		// It does
		w.WriteHeader(http.StatusOK) // code 200

		// Package it into a map->JSON->[]byte
		resp := map[string]interface{}{
			"msg":     "Success",
			"isExist": "true",
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}
	} else {
		log.Println("Key not found in DB")
		// The key doesn't exist in the db
		w.WriteHeader(http.StatusNotFound) // code 404

		// Error response
		resp := map[string]interface{}{
			"msg":     "Error",
			"isExist": "false",
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}
	}
	w.Write(body)

}

// DeleteHandler deletes k:v pairs from the db
func (app *App) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling DELETE request")

	// Get the key from the URL
	vars := mux.Vars(r)
	key := vars["subject"]

	// Declare here, define below
	var err error
	var body []byte

	w.Header().Set("Content-Type", "application/json")

	// Check to see if we've got the key
	if app.db.Contains(key) {
		log.Println("Key found in DB")
		// We do
		w.WriteHeader(http.StatusOK) // code 200

		// Delete it
		app.db.Delete(key)

		// Successful response
		resp := map[string]interface{}{
			"msg": "Success",
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}
	} else {
		log.Println("Key not found in DB")

		// We don't have the key
		w.WriteHeader(http.StatusNotFound) // code 404

		// Error response
		resp := map[string]interface{}{
			"error": "Key does not exist",
			"msg":   "Error",
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}
	}
	w.Write(body)

}
