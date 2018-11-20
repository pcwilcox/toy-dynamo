// app.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson     lelawson
// Pete Wilcox         pcwilcox
// Annie Shen          ashen7
// Victoria Tran       vilatran
//
// This is the source file defining the front end of the RESTful API.
//

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/soheilhy/cmux"
)

// App is a struct to hold the state for the REST API
//
// Initialized with: App app := App{db: {dbAccess}} where db is an object implementing
// the dbAccess interface (see dbAccess.go).
//

// App is a struct representing the externally-accessible state of the data store
type App struct {
	db   dbAccess
	view viewList
}

// Initialize fires up the router and such
func (app *App) Initialize(l net.Listener) {

	// Initialize a router
	r := mux.NewRouter()

	// Since all endpoints use the rootURL we just use a subrouter here
	s := r.PathPrefix(rootURL).Subrouter()

	// This is the search handler, which has a different prefix
	s.HandleFunc(search+keySuffix, app.SearchHandler).Methods(http.MethodGet)

	// This is the view handler, has its own GET, PUT and DELETE
	r.HandleFunc(view, app.ViewPutHandler).Methods(http.MethodPut)
	r.HandleFunc(view, app.ViewGetHandler).Methods(http.MethodGet)
	r.HandleFunc(view, app.ViewDeleteHandler).Methods(http.MethodDelete)

	// Each of the request types gets a handler
	s.HandleFunc(keySuffix, app.PutHandler).Methods(http.MethodPut)
	s.HandleFunc(keySuffix, app.GetHandler).Methods(http.MethodGet)
	s.HandleFunc(keySuffix, app.DeleteHandler).Methods(http.MethodDelete)

	// Make logger
	Logger := handlers.LoggingHandler(MultiLogOutput, r)

	v := &http.Server{
		Handler: Logger,
	}
	log.Println("App initialized")
	// Load up the server through a logger interface
	if err := v.Serve(l); err != cmux.ErrListenerClosed {
		log.Fatalln(err)
	}
}

// PutHandler attempts to put the key:val into the db
func (app *App) PutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling PUT request")

	// Predefine these so they can be used further down
	var err error
	var value string
	var payloadString string
	var payloadInt map[string]int
	var payloadMap map[string]interface{}
	var status int
	var body []byte

	// Safety check - the system will panic trying to read a nil body
	if r.Body != nil {
		// Parse the form so we can read values
		r.ParseForm()

		// Create an intermediate map to parse the payload into
		payloadMap = make(map[string]interface{})

		if len(r.Form) > 0 {
			// Read the values from the request body
			value = r.Form["val"][0]
			if r.Form["payload"] != nil {
				payloadString = r.Form["payload"][0]

				if payloadString != "" {
					// Read the payload into the intermediate map
					err = json.Unmarshal([]byte(payloadString), &payloadMap)
					if err != nil {
						log.Fatalln(err)
					}
				}
			}
		}

		// Convert the intermediate map into map[string]int as needed by KVS
		payloadInt = make(map[string]int)
		for k, v := range payloadMap {
			payloadInt[k] = int(v.(float64))
		}

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
				"result":  "Error",
				"msg":     "Object too large. Size limit is 1MB",
				"payload": payloadInt,
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
				"msg":     "Error",
				"error":   "Key not valid",
				"payload": payloadInt,
			}
			body, err = json.Marshal(resp)
			if err != nil {
				log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
			}
		} else {
			// key/val are valid inputs, let's insert into the db
			log.Println("Key and value lengths ok")

			// Check to see if the db already contains the key
			alive, version := app.db.Contains(key)
			if alive && payloadInt[key] <= version {
				// It does so we'll update it
				log.Println("Key already exists in DB, overwriting...")
				time := time.Now()

				// Add this key/version to its payload
				payloadInt[key] = version + 1

				// Put it in the db
				app.db.Put(key, value, time, payloadInt)

				log.Printf("Inserted key-value pair")
				// Set status
				status = http.StatusCreated // code 201

				// Build the response body
				resp := map[string]interface{}{
					"replaced": true,
					"msg":      "Updated successfully",
					"payload":  payloadInt,
				}
				body, err = json.Marshal(resp)
				if err != nil {
					log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
				}
			} else {
				log.Println("Key does not exist in DB, inserting...")
				// It's a new entry so it gets a different status code
				// This status code is bullshit but the spec got reversed
				status = http.StatusOK // code 200
				time := time.Now()

				// Add this key to its own payload, since the key might be dead we just increment the version number
				payloadInt[key] = version + 1
				app.db.Put(key, value, time, payloadInt)

				// And a slightly different response body
				resp := map[string]interface{}{
					"replaced": false,
					"msg":      "Added successfully",
					"payload":  payloadInt,
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

	var payloadString string

	var err error

	if r.Body != nil {
		log.Println("Body not nil")
		// Read the message body
		s, _ := ioutil.ReadAll(r.Body)
		log.Println(string(s))
		sBody, _ := url.QueryUnescape(string(s))
		if len(sBody) > 0 {
			payloadString = strings.Split(sBody, "=")[1]
			log.Println(payloadString)
		}
	}

	// Create an intermediate map to parse the payload into
	payloadMap := make(map[string]interface{})

	if payloadString != "" {
		// Read the payload into the intermediate map
		err = json.Unmarshal([]byte(payloadString), &payloadMap)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Convert the intermediate map into map[string]int as needed by KVS
	payloadInt := make(map[string]int)
	for k, v := range payloadMap {
		payloadInt[k] = int(v.(float64))
	}
	log.Println(payloadInt)
	// Declare some vars
	var body []byte

	// Same content type for everything
	w.Header().Set("Content-Type", "application/json")

	// See if the key exists in the db
	alive, version := app.db.Contains(key)
	log.Println("Alive: ", alive)
	log.Println("Version: ", version)
	log.Println("Client version: ", payloadInt[key])

	// Check to see if payload of the key is old
	if version < payloadInt[key] {
		w.WriteHeader(http.StatusBadRequest) // Code 400

		log.Println("Key requested is out of date")

		resp := map[string]interface{}{
			"result":  "Error",
			"msg":     "Payload out of date",
			"payload": payloadInt,
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL Error: Failed to marshal JSON response")
		}
	} else if alive {
		// It does
		w.WriteHeader(http.StatusOK) // code 200

		// Get the key out of the db
		val, payload := app.db.Get(key, payloadInt)
		log.Println("Key found in DB")

		// Package it into a map->JSON->[]byte
		resp := map[string]interface{}{
			"result":  "Success",
			"value":   val,
			"payload": payload,
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
			"result":  "Error",
			"error":   "Key does not exist",
			"payload": payloadInt,
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

	// Read the payload out of the message body
	r.ParseForm()

	var payloadString string

	// Create an intermediate map to parse the payload into
	payloadMap := make(map[string]interface{})

	if len(r.Form) > 0 {
		// Read the values from the request body
		if r.Form["payload"] != nil {
			payloadString = r.Form["payload"][0]

			if payloadString != "" {
				// Read the payload into the intermediate map
				err = json.Unmarshal([]byte(payloadString), &payloadMap)
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	}

	// Convert the intermediate map into map[string]int as needed by KVS
	payloadInt := make(map[string]int)
	for k, v := range payloadMap {
		payloadInt[k] = v.(int)
	}
	log.Println("SEARCH with payload ", payloadInt)

	// See if the key exists in the db
	alive, version := app.db.Contains(key)
	if version < payloadInt[key] {
		log.Println("Payload out of date error")
		w.WriteHeader(http.StatusBadRequest) // code 400

		resp := map[string]interface{}{
			"result":  "Error",
			"msg":     "Payload out of date",
			"payload": payloadInt,
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL Error: Failed to marshal JSON response")
		}

	} else if alive {
		log.Println("Key found in DB")
		// It does
		w.WriteHeader(http.StatusOK) // code 200

		// Package it into a map->JSON->[]byte
		resp := map[string]interface{}{
			"result":   "Success",
			"isExists": true,
			"payload":  payloadInt,
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}

	} else {
		log.Println("Key not found in DB")
		// The key doesn't exist in the db
		w.WriteHeader(http.StatusOK) // code 200

		// Error response
		resp := map[string]interface{}{
			"result":   "Success",
			"isExists": false,
			"payload":  payloadInt,
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

	// Read the payload out of the message body
	r.ParseForm()

	var payloadString string

	// Create an intermediate map to parse the payload into
	payloadMap := make(map[string]interface{})

	if len(r.Form) > 0 {
		// Read the values from the request body
		if r.Form["payload"] != nil {
			payloadString = r.Form["payload"][0]

			if payloadString != "" {
				// Read the payload into the intermediate map
				err = json.Unmarshal([]byte(payloadString), &payloadMap)
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	}

	// Convert the intermediate map into map[string]int as needed by KVS
	payloadInt := make(map[string]int)
	for k, v := range payloadMap {
		payloadInt[k] = v.(int)
	}

	// Check to see if we've got the key
	alive, _ := app.db.Contains(key)
	if alive {
		log.Println("Key found in DB")
		// We do
		w.WriteHeader(http.StatusOK) // code 200

		// Delete it
		time := time.Now()
		app.db.Delete(key, time, payloadInt)

		// Successful response
		resp := map[string]interface{}{
			"result":  "Success",
			"msg":     "Key deleted",
			"payload": payloadInt,
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
			"msg":     "Key does not exist",
			"result":  "Error",
			"payload": payloadInt,
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}
	}

	w.Write(body)
}

// ViewPutHandler inititate a view change.
// All containers in the system should add to their view
func (app *App) ViewPutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /view PUT request")

	// Read the payload out of the message body
	r.ParseForm()

	// Declare some vars
	var body []byte
	var err error
	var newPort string

	if len(r.Form) > 0 {
		// Read the values from the request body
		if r.Form["ip_port"] != nil {
			newPort = r.Form["ip_port"][0]
		}
	}
	log.Println(newPort)

	// Same content type for everything
	w.Header().Set("Content-Type", "application/json")

	// Check if the port you want to add new to our view
	if !app.view.Contains(newPort) {
		log.Println("Port to be added is brand new to view: " + newPort)

		// We do
		w.WriteHeader(http.StatusOK) // code 200

		// Add it
		app.view.Add(newPort)

		// Successful response
		resp := map[string]interface{}{
			"result": "Success",
			"msg":    "Successfully added " + newPort + " to view",
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}

	} else {
		log.Println("Port to be added is already in view")

		// We already have the port
		w.WriteHeader(http.StatusNotFound) // code 404

		// Error response
		resp := map[string]interface{}{
			"result": "Error",
			"msg":    newPort + " is already in view",
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}
	}
	w.Write(body)
}

// ViewGetHandler returns the view slice of the system
func (app *App) ViewGetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /view GET request")

	// Declare some vars
	var body []byte
	var err error
	var str string

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // code 200

	// Turn envView into string for JSON response
	str = app.view.String()
	log.Println("My view: " + str)

	// Package it into a map->JSON->[]byte
	resp := map[string]interface{}{
		"view": str,
	}
	body, err = json.Marshal(resp)
	if err != nil {
		log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
	}
	w.Write(body)

}

// ViewDeleteHandler inititate a view change. All containers' system view should change.
func (app *App) ViewDeleteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /view DELETE request")

	// Declare some vars
	var body []byte
	var err error

	// Read the message body
	s, _ := ioutil.ReadAll(r.Body)
	log.Println(string(s))
	sBody, _ := url.QueryUnescape(string(s))
	log.Println(sBody)
	deletePort := strings.Split(sBody, "=")[1]

	// Same content type for everything
	w.Header().Set("Content-Type", "application/json")

	// Check if the port you want to delete is in view
	if app.view.Contains(deletePort) {
		log.Println("Port to be deleted found in view")

		// We do
		w.WriteHeader(http.StatusOK) // code 200

		// Delete it
		app.view.Remove(deletePort)

		// Successful response
		resp := map[string]interface{}{
			"result": "Success",
			"msg":    "Successfully removed " + deletePort + " from view",
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}

	} else {
		log.Println("Port to be deleted not found in view")

		// We don't have the port
		w.WriteHeader(http.StatusNotFound) // code 404

		// Error response
		resp := map[string]interface{}{
			"result": "Error",
			"msg":    deletePort + " is not in current view",
		}
		body, err = json.Marshal(resp)
		if err != nil {
			log.Fatalln("FATAL ERROR: Failed to marshal JSON response")
		}
	}

	w.Write(body)
}
