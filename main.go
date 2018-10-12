/*-
 * main.go
 *
 * Pete Wilcox
 * CruzID: pcwilcox
 * CMPS 128, Fall 2018
 *
 * This is the main source code to satisfy HW1. It implements a Gorilla/Mux router and handles HTTP requests of the following form:
 *		- http://localhost:8080/hello			returns 'Hello world!'
 *		- http://localhost:8080/test			returns 'GET request received' if a GET request
 *												else returns 'POST message received: <msg>' if POST
 * All other requests return 405.
 */
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	/* Set logger to stdout */
	log.SetOutput(os.Stdout)

	/* Initialize a router */
	r := mux.NewRouter()

	/* Assign handlers for request endpoints */
	r.HandleFunc("/hello", HelloHandler)

	/* The test endpoint may or may not include a query */
	r.HandleFunc("/test", TestHandler)
	r.HandleFunc("/test", TestHandler).
		Queries("msg", "{msg}")

	/* Load up the server through a logger interface */
	err := http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, r))
	if err != nil {
		log.Fatalln(err)
	}
}

// HelloHandler accepts GET requests of the type http://localhost:8080/hello
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == "GET" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world!"))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// TestHandler handles the test endpoint
func TestHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == "GET" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET request received"))
	} else if method == "POST" {
		w.WriteHeader(http.StatusOK)
		message := r.FormValue("msg")
		w.Write([]byte(fmt.Sprintf("POST message received: %s", message)))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
