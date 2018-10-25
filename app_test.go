// app_test.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson     lelawson
// Pete Wilcox         pcwilcox
//
// Unit test definitions for app.go
//

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

// Define some constants. These can be reconfigured as needed.
const (
	domain       = "http://localhost"
	root         = "/keyValue-store"
	hostname     = domain + ":" + port + root
	keyExists    = "KEY_EXISTS"
	KeyNotExists = "KEY_DOESN'T_EXIST"
)

var valExists = map[string]string{
	"val": "VALUE EXISTS",
}

// This struct is a stub for the dbAccess object required by the app
type TestKVS struct {
	key       string
	valExists map[string]string
	service   bool
}

// This stub returns true for the key which exists and false for the one which doesn't
func (t *TestKVS) Contains(key string) bool {
	if strings.Compare(key, t.key) == 0 {
		return true
	}
	return false
}

// This stub returns the valExistsue associated with the key which exists, and returns nil for the key which doesn't //
func (t *TestKVS) Get(key string) string {
	if key == t.key {
		return t.valExists["val"]
	}
	return ""
}

// This stub returns true for the key which exists and false for the one which doesn't
func (t *TestKVS) Delete(key string) bool {
	if key == t.key {
		return true
	}
	return false
}

// Returns the value of the service bool
func (t *TestKVS) ServiceUp() bool {
	return t.service
}

// idk lets try this
func (t *TestKVS) Put(key, valExists string) bool {
	return true
}

// TestPutRequestKeyExists verifies the response given when updating a key
func TestPutRequestKeyExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "PUT"
	marsh, err := json.Marshal(valExists)
	ok(t, err)
	reqBody := bytes.NewReader(marsh)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"replaced": "True",
		"msg":      "Updated successfully",
	}

	equals(t, expectedBody, gotBody)
}

// TestPutRequestKeyDoesntExist verifies the response given when adding a new key
func TestPutRequestKeyDoesntExist(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := KeyNotExists

	// The value needs to be > 1MB
	submission := make(map[string]string)
	submission["val"] = valNotExists
	b, err := json.Marshal(submission)
	ok(t, err)
	reqBody := bytes.NewReader(b)
	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "PUT"
	//reqBody := strings.NewReader(valExists)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusCreated // code 201
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"replaced": false,
		"msg":      "Added successfully",
	}

	equals(t, expectedBody, gotBody)
}

// TestPutRequestInvalidKey makes a key with length == 201 and verifies that it fails
func TestPutRequestInvalidKey(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject needs to be very long
	subject := ""
	for i := 0; i < 201; i++ {
		subject = subject + "a"
	}

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "PUT"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusUnprocessableEntity // code 422
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":    "Key not valid",
		"result": "Error",
	}

	equals(t, expectedBody, gotBody)
}

// TestPutRequestInvalidValue tests for values that are too large
func TestPutRequestInvalidValue(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject doesn't really matter
	subject := keyExists

	// The value needs to be > 1MB
	submission := make(map[string]string)
	big := "a"
	for i := 1; i < 1048577; i *= 2 {
		big = big + big
	}
	submission["val"] = big
	b, err := json.Marshal(submission)
	ok(t, err)
	reqBody := bytes.NewReader(b)

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "PUT"
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusUnprocessableEntity // code 422
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	var gotBody map[string]interface{}
	err = json.Unmarshal([]byte(recorder.Body.String()), &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":    "Object too large. Size limit is 1MB",
		"result": "Error",
	}

	equals(t, expectedBody, gotBody)
}

// TestPutRequestServiceDown should return that the service is down
func TestPutRequestServiceDown(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, false}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "PUT"
	m, err := json.Marshal(valExists)
	reqBody := bytes.NewReader(m)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotImplemented // code 501
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result": "Error",
		"msg":    "Server unavailable",
	}

	equals(t, expectedBody, gotBody)
}

// TestGetRequestKeyExists should return success with the "VAL_EXISTS" string
func TestGetRequestKeyExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.GetHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}
	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":   "Success",
		"value": valExists["val"],
	}

	equals(t, expectedBody, gotBody)
}

// TestGetRequestServiceDown should return server unavailable
func TestGetRequestServiceDown(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, false}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.GetHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotImplemented // code 501
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}
	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result": "Error",
		"msg":    "Server unavailable",
	}

	equals(t, expectedBody, gotBody)
}

// TestGetRequestKeyNotExists should return that the key has been replaced/updated successfully
func TestGetRequestKeyNotExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.GetHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := KeyNotExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotFound // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	var gotBody map[string]interface{}
	err = json.Unmarshal([]byte(recorder.Body.String()), &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"error": "Key does not exist",
		"msg":   "Error",
	}

	equals(t, expectedBody, gotBody)
}

// TestDeleteServerDown should return server unavailable
func TestDeleteServerDown(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, false}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.DeleteHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "DELETE"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotImplemented // code 501
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}
	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result": "Error",
		"msg":    "Server unavailable",
	}

	equals(t, expectedBody, gotBody)
}

// TestDeleteKeyExists should return success
func TestDeleteKeyExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.DeleteHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "DELETE"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}
	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg": "Success",
	}

	equals(t, expectedBody, gotBody)
}

// TestDeleteKeyNotExists should return success
func TestDeleteKeyNotExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/{subject}", app.DeleteHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := KeyNotExists

	// Set up the URL
	url := ts.URL + root + "/" + subject

	// Stub a request
	method := "DELETE"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotFound // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)

	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}
	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result": "Error",
		"msg":    "Status code 404",
	}

	equals(t, expectedBody, gotBody)
}

// TestSearchRequestKeyExists verifies the response given when updating a key
func TestSearchRequestKeyExists(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/search/{subject}", app.SearchHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/search/" + subject

	// Stub a request
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":     "Key does exist",
		"isExist": "true",
	}

	equals(t, expectedBody, gotBody)
}

// TestSearchRequestKeyDoesntExist verifies the response given when adding a new key
func TestSearchRequestKeyDoesntExist(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/search/{subject}", app.SearchHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := KeyNotExists

	// Set up the URL
	url := ts.URL + root + "/search/" + subject

	// Stub a request
	method := "GET"
	//reqBody := strings.NewReader(valExists)
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotFound // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":     "Key does not exist",
		"isExist": "false",
	}

	equals(t, expectedBody, gotBody)
}

// TestSearchRequestInvalidKey makes a key with length == 201 and verifies that it fails
func TestSearchRequestInvalidKey(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, true}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/search/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject needs to be very long
	subject := ""
	for i := 0; i < 201; i++ {
		subject = subject + "a"
	}

	// Set up the URL
	url := ts.URL + root + "/search/" + subject

	// Stub a request
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusUnprocessableEntity // code 422
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":    "Key not valid",
		"result": "Error",
	}

	equals(t, expectedBody, gotBody)
}

// TestSearchRequestServiceDown should return that the service is down
func TestSearchRequestServiceDown(t *testing.T) {
	// Stub the db
	db := TestKVS{keyExists, valExists, false}

	// Stub the app
	app := App{&db}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	ok(t, err)

	// Create a router
	r := mux.NewRouter()
	r.HandleFunc(root+"/search/{subject}", app.PutHandler)
	// Stub the server
	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := ts.URL + root + "/search/" + subject

	// Stub a request
	method := "GET"
	m, err := json.Marshal(valExists)
	ok(t, err)
	reqBody := bytes.NewReader(m)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	// Finally, make the request to the function being tested.
	r.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotImplemented // code 501
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result": "Error",
		"msg":    "Server unavailable",
	}

	equals(t, expectedBody, gotBody)
}

// These functions were taken from Ben Johnson's post here: https://medium.com/@benbjohnson/structuring-tests-in-go-46ddee7a25c

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
