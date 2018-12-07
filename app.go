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
	db    dbAccess
	shard Shard
}

// Initialize takes a Listener, assigns a Router to it, and then attaches HTTP handler
// functions in order to implement the RESTful API. Each Serve() event is handled in a
// concurrent goroutine. This function should not return while the system is running,
// thus it panics if there is an error.
func (app *App) Initialize(l net.Listener) {

	// Initialize a router
	r := mux.NewRouter()

	// Many endpoints use the rootURL so we'll save space and make a subrouter
	s := r.PathPrefix(rootURL).Subrouter()

	// This is the search handler, which has a different prefix
	s.HandleFunc(search+keySuffix, app.SearchHandler).Methods(http.MethodGet)

	// These handlers implement the /view endpoint and handle GET, PUT, DELETE
	r.HandleFunc(view, app.ViewPutHandler).Methods(http.MethodPut)
	r.HandleFunc(view, app.ViewGetHandler).Methods(http.MethodGet)
	r.HandleFunc(view, app.ViewDeleteHandler).Methods(http.MethodDelete)

	// These handlers implement the KVS API and handle GET, PUT, DELETE
	s.HandleFunc(keySuffix, app.PutHandler).Methods(http.MethodPut)
	s.HandleFunc(keySuffix, app.GetHandler).Methods(http.MethodGet)
	s.HandleFunc(keySuffix, app.DeleteHandler).Methods(http.MethodDelete)

	// LoggingHandler allows us to log all router activity to our predefined log
	Logger := handlers.LoggingHandler(MultiLogOutput, r)

	// We define a server here and attach the log-enabled router to it
	v := &http.Server{
		Handler: Logger,
	}

	log.Println("REST API initialized.")

	// Start the server. The server will return two types of errors:
	//   cmux.ErrListenerClosed - This error occurs when the connection
	//          from the Listener closes, and it isn't even really an
	//          error as far as we are concerned, so ignore it.
	//   Anything else          - This indicates an actual error with the
	//          connection and since the app doesn't really have any ability
	//          to handle it, we'll just log it and panic.
	if err := v.Serve(l); err != cmux.ErrListenerClosed {
		log.Fatalln(err)
	}
}

// PutHandler responds to PUT requests on the /keyValue-store/{key} endpoint.
// It processes the payload attached with the request in order to store it with
// the key. It checks for valid inputs and attempts not to crash if it sees them.
func (app *App) PutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling PUT request")

	// Each of these variables is declared here and then defined further down in
	// the function, depending on how the control structures shake out.
	var err error                         // Error value if any
	var value string                      // Value to be stored
	var payloadString string              // The payload sent by the client
	var payloadInt map[string]int         // The payload we store
	var payloadMap map[string]interface{} // Intermediate map
	var status int                        // Status code returned
	var body []byte                       // Body of response

	// Safety check - the system will panic if it tries to read a nil body
	if r.Body != nil {
		// Parse the form so we can read values in the request
		r.ParseForm()

		// Create an intermediate map to translate the payload
		payloadMap = make(map[string]interface{})

		// It's possible to send an empty form
		if len(r.Form) > 0 {
			// Read the values from the request body
			value = r.Form["val"][0]

			// Check to see if the client sends a payload
			if r.Form["payload"] != nil {
				// If they do, it starts as a string value
				payloadString = r.Form["payload"][0]

				// And that string might be empty
				if payloadString != "" {
					// Decode the payload into the empty map from above
					err = json.Unmarshal([]byte(payloadString), &payloadMap)
					if err != nil {
						log.Fatalln(err)
					}
				}
			}
		}

		// Convert the intermediate map into map[string]int as stored by KVS by
		// iterating over all the keys in the client's payload. Start by creating
		// the intermediate map.
		payloadInt = make(map[string]int)
		for k, v := range payloadMap {
			// Numbers decoded by JSON are stored as
			// float64, so we start with a type assertion
			// To register them that way. We then cast it
			// to an int and store it in the map.
			payloadInt[k] = int(v.(float64))
		}

		// This pulls the {subject} out of the URL, that forms the key. It's called
		// {subject} instead of {key} because the spec for HW2 called it that and it
		// stuck. We could change it I guess. This utilizes Gorilla Mux's URL parsing.
		vars := mux.Vars(r)
		key := vars["subject"]

		// Check for valid input
		if len(value) > maxVal {
			// The value is > 1MB so error out
			log.Println("ERROR: Value length too long")

			// Set the status code
			status = http.StatusUnprocessableEntity // code 422

			// Form the response into something that JSON can handle. The payload here
			// is the one sent by the client with the request.
			resp := map[string]interface{}{
				"result":  "Error",
				"msg":     "Object too large. Size limit is 1MB",
				"payload": payloadInt,
			}

			// Encode the response from a map into a byte slice
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

			// Check to see if the db already contains the key. The type of response
			// the client receives here depends on their causul history. If there is
			// a constraint such that they shouldn't be shown our version of the key,
			// that's the same as if the key doesn't exist.
			alive, version := app.db.Contains(key)
			log.Println("Alive: ", alive)
			log.Println("Version: ", version)

			// The key hasn't been deleted, and it's recent enough to show to the,
			// client so we can give them the 'overwrite' response.
			if alive && payloadInt[key] <= version {
				log.Println("Key already exists in DB, overwriting...")

				// Set the timestamp for the new version of the key.
				time := time.Now()

				// Create the payload to be inserted into the db, starting with this key
				newPayload := map[string]int{key: version + 1}

				// Add each of the client's payload elements
				for k, v := range payloadInt {
					if k != key {
						newPayload[k] = v
					}
				}

				// Put it in the db
				app.db.Put(key, value, time, newPayload)

				// Set status
				status = http.StatusCreated // code 201

				// Build the response body, including the payload. It's got the adjustment
				// above, which isn't really strictly to spec, but it should be ok.
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
				// Either the key is too old to show to the client, or it's new enough but it's been deleted.
				// In either case, from the client's perspective, it doesn't exist.
				log.Println("Key does not exist in DB, inserting...")
				status = http.StatusOK // code 200
				time := time.Now()

				// Create the payload to be inserted into the db, starting with this key
				newPayload := map[string]int{key: version + 1}

				// Add each of the client's payload elements
				for k, v := range payloadInt {
					if k != key {
						newPayload[k] = v
					}
				}

				// Put it in the db
				app.db.Put(key, value, time, newPayload)

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
		// We only get here in a weird state where the body didn't happen or something.
		log.Println("ERROR: No data sent with request")

		// There's no body in the request
		status = http.StatusNotFound // code 404

		// And a slightly different response body - we can only send an empty payload.
		resp := map[string]interface{}{
			"msg":     "Error",
			"error":   "Value is missing",
			"payload": map[string]interface{}{},
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

// GetHandler responds to GET requests to /keyValue-store/{subject} endpoint. It
// uses slightly different techniques to read the request because GET requests
// don't normally carry form data. It contains logic for checking the client's
// causal history in order to assure no constraints are violated.
func (app *App) GetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling GET request")

	// Read the key from the URL using the Gorilla Mux URL parsing.
	vars := mux.Vars(r)
	key := vars["subject"]

	// These two variables are declared here and assigned further down.
	var payloadMap map[string]interface{} // Intermediate map for decoding
	var payloadString string              // Payload sent by the client
	var err error                         // Error value
	var body []byte                       // Response body

	// Check that the body isn't nil
	if r.Body != nil {
		// Read the message body into a string
		s, _ := ioutil.ReadAll(r.Body)
		log.Println(string(s))

		// Python packs the input in Unicode for some reason so we need to convert it
		sBody, _ := url.QueryUnescape(string(s))
		if len(sBody) > 0 {
			// The actual payload we care about comes after the equals sign. This splits the
			// input into a slice and takes the second element of that slice for the payload.
			payloadString = strings.Split(sBody, "=")[1]
			log.Println(payloadString)
		}
	}

	// Create an intermediate map to parse the payload into
	payloadMap = make(map[string]interface{})

	if payloadString != "" {
		// Read the payload into the intermediate map
		err = json.Unmarshal([]byte(payloadString), &payloadMap)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Convert the intermediate map into map[string]int as needed by KVS. As with PutHandler
	// we use a type assertion to set the value to float64 in the payloadMap and then typecast
	// it as an integer to store it in the payloadInt map.
	payloadInt := make(map[string]int)
	for k, v := range payloadMap {
		payloadInt[k] = int(v.(float64))
	}
	log.Println(payloadInt)

	// Same content type for everything
	w.Header().Set("Content-Type", "application/json")

	// Here we'll check to see if the requested key exists and get its version.
	alive, version := app.db.Contains(key)
	log.Println("Alive: ", alive)
	log.Println("Version: ", version)
	log.Println("Client version: ", payloadInt[key])

	// If the version of the key stored in the DB is older than the value in the client's payload,
	// then it would violate causality to show the key to the client. In this case we return an error
	// message per the spec.
	if version < payloadInt[key] {
		w.WriteHeader(http.StatusBadRequest) // Code 400

		log.Println("Key requested is out of date")

		// Form the response body, starting with a map of values. We give the client their own payload back.
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
		// The version is recent enough to show to the client, and the key has not been deleted, so we can
		// return the values normally.
		w.WriteHeader(http.StatusOK) // code 200

		// Get the key and its stored payload from the DB. Get() returns the supremum of the client's and key's
		// payloads using the function mergeClocks(), and that's what is returned below.
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
		// If we get here it's because the key has been deleted, and that deletion is recent enough that it doesn't
		// violate causality to tell the client about it.
		log.Println("Key not found in DB")
		w.WriteHeader(http.StatusNotFound) // code 404

		// Error response, and we just return the payload sent with the request.
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

// SearchHandler implements the /keyValue-store/search/{subject} endpoint and otherwise contains very similar
// logic to the GetHandler.
func (app *App) SearchHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling SEARCH request")

	// Read the key from the URL using Gorilla Mux URL parsing.
	vars := mux.Vars(r)
	key := vars["subject"]

	// Declare some variables here and define them below.
	var body []byte          // Response body
	var err error            // Error value
	var payloadString string // Payload sent by the client
	var payloadMap map[string]interface{}

	// Same content type for everything
	w.Header().Set("Content-Type", "application/json")

	// Check that the body isn't nil
	if r.Body != nil {
		// Read the message body into a string
		s, _ := ioutil.ReadAll(r.Body)
		log.Println(string(s))

		// Python packs the input in Unicode for some reason so we need to convert it
		sBody, _ := url.QueryUnescape(string(s))
		if len(sBody) > 0 {
			// The actual payload we care about comes after the equals sign. This splits the
			// input into a slice and takes the second element of that slice for the payload.
			payloadString = strings.Split(sBody, "=")[1]
			log.Println(payloadString)
		}
	}

	// Create an intermediate map to parse the payload into
	payloadMap = make(map[string]interface{})

	if payloadString != "" {
		// Read the payload into the intermediate map
		err = json.Unmarshal([]byte(payloadString), &payloadMap)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Convert the intermediate map into map[string]int as needed by KVS. As with PutHandler
	// we use a type assertion to set the value to float64 in the payloadMap and then typecast
	// it as an integer to store it in the payloadInt map.
	payloadInt := make(map[string]int)
	for k, v := range payloadMap {
		payloadInt[k] = int(v.(float64))
	}
	log.Println(payloadInt)

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

	// These two variables are declared here and assigned further down.
	var payloadMap map[string]interface{} // Intermediate map for decoding
	var payloadString string              // Payload sent by the client
	var err error                         // Error value
	var body []byte                       // Response body

	w.Header().Set("Content-Type", "application/json")

	// Check that the body isn't nil
	if r.Body != nil {
		// Read the message body into a string
		s, _ := ioutil.ReadAll(r.Body)
		log.Println(string(s))

		// Python packs the input in Unicode for some reason so we need to convert it
		sBody, _ := url.QueryUnescape(string(s))
		if len(sBody) > 0 {
			// The actual payload we care about comes after the equals sign. This splits the
			// input into a slice and takes the second element of that slice for the payload.
			payloadString = strings.Split(sBody, "=")[1]
			log.Println(payloadString)
		}
	}

	// Create an intermediate map to parse the payload into
	payloadMap = make(map[string]interface{})

	if payloadString != "" {
		// Read the payload into the intermediate map
		err = json.Unmarshal([]byte(payloadString), &payloadMap)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Convert the intermediate map into map[string]int as needed by KVS. As with PutHandler
	// we use a type assertion to set the value to float64 in the payloadMap and then typecast
	// it as an integer to store it in the payloadInt map.
	payloadInt := make(map[string]int)
	for k, v := range payloadMap {
		payloadInt[k] = int(v.(float64))
	}

	// Here we'll check to see if the requested key exists and get its version.
	alive, version := app.db.Contains(key)

	// If the version of the key stored in the DB is older than the value in the client's payload,
	// then it would violate causality to show the key to the client. In this case we return an error
	// message per the spec.
	if version < payloadInt[key] {
		w.WriteHeader(http.StatusBadRequest) // Code 400

		log.Println("Key requested is out of date")

		// Form the response body, starting with a map of values. We give the client their own payload back.
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
		// The version is recent enough to show to the client, and the key has not been deleted, so we can
		// return the values normally.
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
	if !app.shard.ContainsServer(newPort) {
		log.Println("Port to be added is brand new to view: " + newPort)

		// We do
		w.WriteHeader(http.StatusOK) // code 200

		// Add it
		app.shard.Add(newPort)

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
	str = app.shard.String()
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
	if app.shard.ContainsServer(deletePort) {
		log.Println("Port to be deleted found in view")

		// We do
		w.WriteHeader(http.StatusOK) // code 200

		// Delete it
		app.shard.Remove(deletePort)

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
