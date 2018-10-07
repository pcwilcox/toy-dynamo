/*-
 * restful.go
 *
 * Pete Wilcox
 * CruzID: pcwilcox
 * CMPS 128, Fall 2018
 *
 * This is the main source code to satisfy HW1. It implements a Gorilla/Mux router and handles HTTP requests of the following form:
 *
 */
package restful

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	/* Initialize a router */
	r := mux.NewRouter()

	/* Assign handlers for request endpoints */
	r.HandleFunc("/hello", HelloHandler)

	/* The test endpoint may or may not include a query */
	r.HandleFunc("/test", TestHandler)
	r.HandleFunc("/test", TestHandler).
		Queries("msg", "{msg}")
	log.Println(http.ListenAndServe(":8080", r))
}

// HelloHandler accepts GET requests of the type http://localhost:8080/hello
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == "GET" {
		w.Write([]byte("Hello world!"))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// TestHandler handles the test endpoint
func TestHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == "GET" {
		w.Write([]byte("GET request received"))
		w.WriteHeader(http.StatusOK)
	} else if method == "POST" {
		message := r.FormValue("msg")
		w.Write([]byte(fmt.Sprintf("POST message received: %s", message)))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
