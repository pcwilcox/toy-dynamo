/* app.go
 *
 * CMPS 128 Fall 2018
 *
 * Lawrence Lawson     lelawson
 * Pete Wilcox         pcwilcox
 *
 * This is the source file defining the front end of the RESTful API.
 */

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

/* Set this externally */
var port string

// The data access object
var db dbAccess

// rootURL is the path prefix for the kvs as in: http://localhost:{port}/ROOT_URL/foo
const rootURL = "/keyValue-store"

// Initialize fires up the router and such
func Initialize(k *dbAccess) {
	/* Initialize a router */
	r := mux.NewRouter()

	// access the database hopefully
	db = *k

	/* There's basically two endpoints here so we'll set up a subrouter */
	s := r.PathPrefix(rootURL).Subrouter()

	/* This handles COUNT calls */
	r.HandleFunc(rootURL, CountHandler).Methods("COUNT")

	/* Methods for specific items */
	s.HandleFunc("/{subject}", PutHandler).Methods("PUT")
	s.HandleFunc("/{subject}", GetHandler).Methods("GET")
	s.HandleFunc("/{subject}", DeleteHandler).Methods("Delete")

	/* Load up the server through a logger interface */
	err := http.ListenAndServe(":"+port, handlers.LoggingHandler(os.Stdout, r))
	if err != nil {
		log.Fatalln(err)
	}
}

// CountHandler returns the db's count function result
func CountHandler(w http.ResponseWriter, r *http.Request) {
	if ServiceUp() {
		w.WriteHeader(http.StatusOK) // code 200
		w.Header().Set("Content-Type", "application/json")
		count := strconv.Itoa(db.Count())
		resp := map[string]string{"result": "success", "msg": count}
		body, _ := json.Marshal(resp)
		w.Write([]byte(body))
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write(ServiceDownHandler())
	}
}

// PutHandler attempts to put the key:val into the db
func PutHandler(w http.ResponseWriter, r *http.Request) {
	if ServiceUp() {
		// This automatically loads the form values from the request
		r.ParseForm()
		val := r.FormValue("val")

		// These values are fixed per spec
		maxVal := 1048576 // 1 megabyte
		maxKey := 200     // 200 characters

		// This pulls the {subject} out of the URL which forms the key
		vars := mux.Vars(r)
		key := vars["subject"]

		// assume the key is there already
		newKey := false
		replaced := "True"
		msg := "Updated successfully"

		// Init an empty response and overwrite it below
		resp := map[string]string{}

		// Check for valid input
		if len(val) > maxVal {
			w.WriteHeader(http.StatusUnprocessableEntity)
			resp = map[string]string{"result": "Error", "msg": "Object too large. Size limit is 1MB"}
		} else if len(key) > maxKey {
			w.WriteHeader(http.StatusUnprocessableEntity)
			resp = map[string]string{"msg": "Key not valid", "result": "Error"}
		} else {
			// key/val are valid inputs, let's insert into the db
			if db.Put(key, val) {
				newKey = true // we made a new entry
				replaced = "False"
				msg = "Added successfully"
			}
			resp = map[string]string{"replaced": replaced, "msg": msg}

			if newKey == true {
				w.WriteHeader(http.StatusCreated) // code 201
			} else {
				w.WriteHeader(http.StatusOK) // code 200
			}
		}
		body, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write(ServiceDownHandler())
	}
}

// GetHandler gets the corresponding val from the db
func GetHandler(w http.ResponseWriter, r *http.Request) {
	if ServiceUp() {
		vars := mux.Vars(r)
		key := vars["subject"]
		resp := map[string]string{}
		if db.Contains(key) {
			w.WriteHeader(http.StatusOK)
			val, _ := db.Get(key)
			resp = map[string]string{"result": "Success", "value": val}
		} else {
			w.WriteHeader(http.StatusNotFound)
			resp = map[string]string{"result": "Error", "value": "Not Found"}
		}
		w.Header().Set("Content-Type", "application/json")
		body, _ := json.Marshal(resp)
		w.Write(body)
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not implemented yet"))
}

// DeleteHandler deletes k:v pairs from the db
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if ServiceUp() {
		vars := mux.Vars(r)
		key := vars["subject"]
		resp := map[string]string{}
		if db.Contains(key) {
			w.WriteHeader(http.StatusOK)
			db.Delete(key)
			resp = map[string]string{"result": "Success"}

		} else {
			w.WriteHeader(http.StatusNotFound)
			resp = map[string]string{"result": "Error", "msg": "Status code 404"}
		}
		w.Header().Set("Content-Type", "application/json")
		body, _ := json.Marshal(resp)
		w.Write(body)
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not implemented yet"))
}

// ServiceDownHandler spits out the required error in JSON format
func ServiceDownHandler() []byte {
	responseMap := map[string]string{"result": "Error", "msg": "Server unavilable"}
	js, _ := json.Marshal(responseMap)
	return js
}

// ServiceUp checks to see if the leader is up
func ServiceUp() bool {
	return true
}
