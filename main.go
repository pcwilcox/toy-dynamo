package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/hello", NameHandler).
		Methods("GET").
		Queries("name", "{name}")
	r.HandleFunc("/hello", NamelessHandler).
		Methods("GET")
	r.HandleFunc("/check", CheckHandler)

	log.Println(http.ListenAndServe(":8080", r))
}

// NamelessHandler accepts GET requests of the type http://localhost:8080/hello
func NamelessHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello user!\n"))
}

// NameHandler accepts GET requests of the type http://localhost:8080/hello?name={name}
func NameHandler(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["name"]
	w.Write([]byte(fmt.Sprintf("Hello %s!\n", username)))
}

func CheckHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method == "GET" {
		w.Write([]byte("This is a GET request\n"))
		w.WriteHeader(http.StatusOK)
	} else if method == "POST" {
		w.Write([]byte("This is a POST request\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
