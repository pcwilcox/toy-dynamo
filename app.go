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
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

/* Set this externally */
var port string

// rootURL is the path prefix for the kvs as in: http://localhost:{port}/ROOT_URL/foo
const rootURL = "/keyValue-store"

// Initialize fires up the router and such
func Initialize() {
	/* Initialize a router */
	r := mux.NewRouter()

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
		r.ParseForm()
		newKey := false // assume the key is there already
		for k, v := range r.Form {
			if db.Put(k, strings.Join(v, " ")) {
				newKey = true // we made a new entry
			}
		}
		if newKey == true {
			w.WriteHeader(http.StatusCreated) // code 201
		} else {
			w.WriteHeader(http.StatusOK) // code 200
		}
		w.Header().Set("Content-Type", "application/json")

	} else {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write(ServiceDownHandler())
	}
}

// GetHandler gets the corresponding val from the db
func GetHandler(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not implemented yet"))
}

// DeleteHandler deletes k:v pairs from the db
func DeleteHandler(w http.ResponseWriter, r *http.Request) {

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
