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

// Global server and router variables used in tests
var testServer http.Server
var testRouter http.Handler

// This struct is a stub for the dbAccess object required by the app
type TestKVS struct {
	dbKey     string
	dbVal     string
	dbService bool
}

// This stub returns true for the key which exists and false for the one which doesn't
func (t *TestKVS) Contains(key string) bool {
	if key == t.dbKey {
		return true
	}
	return false
}

// This stub returns the valExistsue associated with the key which exists, and returns nil for the key which doesn't //
func (t *TestKVS) Get(key string) string {
	if key == t.dbKey {
		return t.dbVal
	}
	return ""
}

// This stub returns true for the key which exists and false for the one which doesn't
func (t *TestKVS) Delete(key string) bool {
	if key == t.dbKey {
		return true
	}
	return false
}

// Returns the value of the service bool
func (t *TestKVS) ServiceUp() bool {
	return t.dbService
}

// idk lets try this
func (t *TestKVS) Put(key, valExists string) bool {
	return true
}

// Trying to reduce code repetition
func setup(key string, val string, service bool) (string, *mux.Router) {
	k := TestKVS{dbKey: key, dbVal: val, dbService: service}

	// Stub the app
	app := App{&k}

	l, err := net.Listen("tcp", "")
	if err != nil {
		panic(err)
	}

	// Get a test router and add handlers to it
	testRouter := mux.NewRouter()

	testRouter.HandleFunc(rootURL+keySuffix, app.PutHandler).Methods(http.MethodPut)
	testRouter.HandleFunc(rootURL+keySuffix, app.GetHandler).Methods(http.MethodGet)
	testRouter.HandleFunc(rootURL+keySuffix, app.DeleteHandler).Methods(http.MethodDelete)
	testRouter.HandleFunc(rootURL+search+keySuffix, app.SearchHandler).Methods(http.MethodGet)
	testRouter.HandleFunc(rootURL+alive, app.AliveHandler).Methods(http.MethodGet)

	// Stub the server
	testServer := httptest.NewUnstartedServer(testRouter)
	testServer.Listener.Close()
	testServer.Listener = l
	testServer.Start()

	return testServer.URL, testRouter
}

// Close down the server after each test
func teardown() {
	testServer.Close()
}

// TestPutRequestKeyExists verifies the response given when updating a key
func TestPutRequestKeyExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodPut
	reqBody := strings.NewReader("val=" + valExists)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":      "Updated successfully",
		"replaced": true,
	}

	equals(t, expectedBody, gotBody)

	teardown()
}

// TestPutRequestKeyDoesntExist verifies the response given when adding a new key
func TestPutRequestKeyDoesntExist(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyNotExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodPut
	reqBody := strings.NewReader("val=" + valNotExists)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusCreated // code 201
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":      "Added successfully",
		"replaced": false,
	}

	equals(t, expectedBody, gotBody)

	teardown()
}

// TestPutRequestInvalidKey makes a key with length == 201 and verifies that it fails
func TestPutRequestInvalidKey(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	bigKey := "a"
	for i := 1; i < 260; i *= 2 {
		bigKey = bigKey + bigKey
	}

	// This subject exists in the store already
	subject := bigKey

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodPut
	reqBody := strings.NewReader("val=" + valNotExists)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusUnprocessableEntity // code 422
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":   "Error",
		"error": "Key not valid",
	}

	equals(t, expectedBody, gotBody)

	teardown()
}

// TestPutRequestInvalidValue tests for values that are too large
func TestPutRequestInvalidValue(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	bigVal := "a"
	for i := 1; i < 1<<21; i *= 2 {
		bigVal = bigVal + bigVal
	}

	// This subject exists in the store already
	subject := keyNotExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodPut
	reqBody := strings.NewReader("val=" + bigVal)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusUnprocessableEntity // code 422
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result": "Error",
		"msg":    "Object too large. Size limit is 1MB",
	}

	equals(t, expectedBody, gotBody)

	teardown()
}

// TestPutRequestServiceDown should return that the service is down
func TestPutRequestServiceDown(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, false)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyNotExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodPut
	reqBody := strings.NewReader("val=" + valNotExists)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

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

	teardown()
}

// TestGetRequestKeyExists should return success with the "VAL_EXISTS" string
func TestGetRequestKeyExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodGet
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"value": valExists,
		"msg":   "Success",
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestGetRequestServiceDown should return server unavailable
func TestGetRequestServiceDown(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, false)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodGet
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

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

	teardown()

}

// TestGetRequestKeyNotExists should return that the key has been replaced/updated successfully
func TestGetRequestKeyNotExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyNotExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodGet
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotFound // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"error": "Key does not exist",
		"msg":   "Error",
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestDeleteServerDown should return server unavailable
func TestDeleteServerDown(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, false)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyNotExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodDelete
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotImplemented // code 501
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":    "Server unavailable",
		"result": "Error",
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestDeleteKeyExists should return success
func TestDeleteKeyExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodDelete
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 200
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

	teardown()

}

// TestDeleteKeyNotExists should not return success
func TestDeleteKeyNotExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyNotExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodDelete
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotFound // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":   "Error",
		"error": "Key does not exist",
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestSearchRequestKeyExists verifies the response given when updating a key
func TestSearchRequestKeyExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := serverURL + rootURL + search + "/" + subject

	// Stub a request
	method := http.MethodGet
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"isExist": "true",
		"msg":     "Success",
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestSearchRequestKeyDoesntExist verifies the response given when adding a new key
func TestSearchRequestKeyDoesntExist(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyNotExists

	// Set up the URL
	url := serverURL + rootURL + search + "/" + subject

	// Stub a request
	method := http.MethodGet
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotFound // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"isExist": "false",
		"msg":     "Error",
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestSearchRequestInvalidKey makes a key with length == 201 and verifies that it fails
func TestSearchRequestInvalidKey(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, true)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	bigKey := "a"
	for i := 1; i < 201; i *= 2 {
		bigKey = bigKey + bigKey
	}

	// This subject exists in the store already
	subject := bigKey

	// Set up the URL
	url := serverURL + rootURL + search + "/" + subject

	// Stub a request
	method := http.MethodGet
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotFound // code 404
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"isExist": "false",
		"msg":     "Error",
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestSearchRequestServiceDown should return that the service is down
func TestSearchRequestServiceDown(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists, false)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := serverURL + rootURL + search + "/" + subject

	// Stub a request
	method := http.MethodGet
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusNotImplemented // code 501
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":    "Server unavailable",
		"result": "Error",
	}

	equals(t, expectedBody, gotBody)

	teardown()

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
