package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/hello", HelloHandler)
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
	} else if method == "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		w.WriteHeader(http.StatusNotFound)
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
		w.WriteHeader(http.StatusNotFound)
	}
}
