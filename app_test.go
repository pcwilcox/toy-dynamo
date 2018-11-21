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
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/gorilla/mux"
)

// Global server and router variables used in tests
var testServer http.Server
var testRouter http.Handler
var testKVS TestKVS
var testApp App

// This struct is a stub for the dbAccess object required by the app
type TestKVS struct {
	dbKey   string
	dbVal   string
	dbClock map[string]int
	dbTime  time.Time
}

func (kvs *TestKVS) GetTimestamp(key string) time.Time {
	if key == kvs.dbKey {
		return kvs.dbTime
	}
	return time.Time{}
}

// This stub returns true for the key which exists and false for the one which doesn't
func (kvs *TestKVS) Contains(key string) (bool, int) {
	if key == kvs.dbKey {
		return true, 1
	}
	return false, 0
}

func (kvs *TestKVS) GetClock(key string) map[string]int {
	if key == kvs.dbKey {
		return kvs.dbClock
	}
	return map[string]int{}
}

// This stub returns the valExistsue associated with the key which exists, and returns nil for the key which doesn't //
func (kvs *TestKVS) Get(key string, clock map[string]int) (string, map[string]int) {
	if key == kvs.dbKey {
		return kvs.dbVal, kvs.dbClock
	}
	return "", nil
}

// This stub returns true for the key which exists and false for the one which doesn't
func (kvs *TestKVS) Delete(key string, timestamp time.Time, payload map[string]int) bool {
	if key == kvs.dbKey {
		return true
	}
	return false
}

// idk lets try this
func (kvs *TestKVS) Put(key, valExists string, time time.Time, payload map[string]int) bool {
	for k := range kvs.dbClock {
		kvs.dbClock[k] = 0
	}
	for k, v := range payload {
		kvs.dbClock[k] = v
	}
	return true
}

func (kvs *TestKVS) OverwriteEntry(key string, entry KeyEntry) {
	// goes nowhere does nothing
}

func (kvs *TestKVS) GetTimeGlob() timeGlob {
	m := make(map[string]time.Time)
	m[kvs.dbKey] = kvs.dbTime
	g := timeGlob{List: m}
	return g
}

func (kvs *TestKVS) GetEntryGlob(g timeGlob) entryGlob {
	m := make(map[string]Entry)
	e := Entry{
		Version:   1,
		Clock:     kvs.dbClock,
		Timestamp: kvs.dbTime,
		Value:     kvs.dbVal,
		Tombstone: false,
	}
	m[kvs.dbKey] = e
	j := entryGlob{Keys: m}
	return j
}

// Trying to reduce code repetition
func setup(key string, val string) (string, *mux.Router) {
	clock := map[string]int{key: 1}
	testKVS = TestKVS{dbKey: key, dbVal: val, dbClock: clock}

	// This should probably be converted to a mock instance
	v := NewView(testMain, testView)

	// Stub the app
	testApp := App{&testKVS, *v}

	l, err := net.Listen("tcp", "")
	if err != nil {
		panic(err)
	}

	// Get a test router and add handlers to it
	testRouter := mux.NewRouter()

	testRouter.HandleFunc(rootURL+keySuffix, testApp.PutHandler).Methods(http.MethodPut)
	testRouter.HandleFunc(rootURL+keySuffix, testApp.GetHandler).Methods(http.MethodGet)
	testRouter.HandleFunc(rootURL+keySuffix, testApp.DeleteHandler).Methods(http.MethodDelete)
	testRouter.HandleFunc(rootURL+search+keySuffix, testApp.SearchHandler).Methods(http.MethodGet)
	testRouter.HandleFunc(view, testApp.ViewPutHandler).Methods(http.MethodPut)
	testRouter.HandleFunc(view, testApp.ViewGetHandler).Methods(http.MethodGet)
	testRouter.HandleFunc(view, testApp.ViewDeleteHandler).Methods(http.MethodDelete)

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
	serverURL, router := setup(keyExists, valExists)

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

	expectedStatus := http.StatusCreated // code 201
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedPayload := map[string]interface{}{keyExists: float64(2)}
	expectedBody := map[string]interface{}{
		"msg":      "Updated successfully",
		"replaced": true,
		"payload":  expectedPayload,
	}

	equals(t, expectedBody, gotBody)

	teardown()
}

// TestPutRequestKeyDoesntExist verifies the response given when adding a new key
func TestPutRequestKeyDoesntExist(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	expectedPayload := map[string]interface{}{keyNotExists: float64(1)}
	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"msg":      "Added successfully",
		"replaced": false,
		"payload":  expectedPayload,
	}

	if diff := deep.Equal(expectedBody, gotBody); diff != nil {
		t.Error(diff)
	}
	equals(t, expectedBody, gotBody)

	teardown()
}

func TestPutRequestEmptyBody(t *testing.T) {
	serverURL, router := setup(keyExists, valExists)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyNotExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Stub a request
	method := http.MethodPut
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		"msg":     "Error",
		"error":   "Value is missing",
		"payload": map[string]interface{}{},
	}

	if diff := deep.Equal(expectedBody, gotBody); diff != nil {
		t.Error(diff)
	}
	equals(t, expectedBody, gotBody)

	teardown()

}

// TestPutRequestInvalidKey makes a key with length == 201 and verifies that it fails
func TestPutRequestInvalidKey(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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
		"msg":     "Error",
		"error":   "Key not valid",
		"payload": map[string]interface{}{},
	}

	equals(t, expectedBody, gotBody)

	teardown()
}

// TestPutRequestInvalidValue tests for values that are too large
func TestPutRequestInvalidValue(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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
		"result":  "Error",
		"msg":     "Object too large. Size limit is 1MB",
		"payload": map[string]interface{}{},
	}

	equals(t, expectedBody, gotBody)

	teardown()
}

// TestGetRequestKeyExists should return success with the "VAL_EXISTS" string
func TestGetRequestKeyExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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

	expectedPayload := map[string]interface{}{keyExists: float64(1)}
	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"value":   valExists,
		"result":  "Success",
		"payload": expectedPayload,
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestGetRequestKeyNotExists should return that the key has been replaced/updated successfully
func TestGetRequestKeyNotExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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
		"error":   "Key does not exist",
		"result":  "Error",
		"payload": map[string]interface{}{},
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

func TestGetRequestReturnsPayload(t *testing.T) {

	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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

	expectedPayload := map[string]interface{}{keyExists: float64(1)}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"value":   valExists,
		"result":  "Success",
		"payload": expectedPayload,
	}

	equals(t, expectedBody, gotBody)

	teardown()
}

// TestDeleteKeyExists should return success
func TestDeleteKeyExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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
		"msg":     "Key deleted",
		"result":  "Success",
		"payload": map[string]interface{}{},
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestDeleteKeyNotExists should not return success
func TestDeleteKeyNotExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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
		"result":  "Error",
		"msg":     "Key does not exist",
		"payload": map[string]interface{}{},
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestSearchRequestKeyExists verifies the response given when updating a key
func TestSearchRequestKeyExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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
		"isExists": true,
		"result":   "Success",
		"payload":  map[string]interface{}{},
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestSearchRequestKeyDoesntExist verifies the response given when adding a new key
func TestSearchRequestKeyDoesntExist(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"isExists": false,
		"result":   "Success",
		"payload":  map[string]interface{}{},
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

// TestSearchRequestInvalidKey makes a key with length == 201 and verifies that it fails
func TestSearchRequestInvalidKey(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

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

	expectedStatus := http.StatusOK // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"isExists": false,
		"result":   "Success",
		"payload":  map[string]interface{}{},
	}

	equals(t, expectedBody, gotBody)

	teardown()

}

func TestViewPutRequestViewExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// Set up the URL
	url := serverURL + view
	log.Println("*** this is test view: " + testMain)
	// Stub a request
	method := http.MethodPut
	reqBody := strings.NewReader("ip_port=" + testMain)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		"result": "Error",
		"msg":    testMain + " is already in view",
	}

	equals(t, expectedBody, gotBody)

	teardown()
}

func TestViewPutRequestViewNotExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// Set up the URL
	url := serverURL + view
	log.Println("*** this is test view: " + viewNotExist)
	// Stub a request
	method := http.MethodPut
	reqBody := strings.NewReader("ip_port=" + viewNotExist)
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
		"result": "Success",
		"msg":    "Successfully added " + viewNotExist + " to view",
	}
	equals(t, expectedBody, gotBody)

}

func TestViewGetRequestViewExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// Set up the URL
	url := serverURL + view

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
	// Hard coded View example for testing
	expectedBody := map[string]interface{}{
		"view": testView,
	}

	equals(t, expectedBody, gotBody)
	teardown()
}

func TestViewDeleteRequestViewExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// Set up the URL
	url := serverURL + view

	// Stub a request
	method := http.MethodDelete
	reqBody := strings.NewReader("ip_port=" + viewExist)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	log.Println(reqBody)

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
	// Hard coded View example for testing
	expectedBody := map[string]interface{}{
		"result": "Success",
		"msg":    "Successfully removed " + viewExist + " from view",
	}

	equals(t, expectedBody, gotBody)
	teardown()
}

func TestViewDeleteRequestViewNotExists(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// Set up the URL
	url := serverURL + view

	// Stub a request
	method := http.MethodDelete
	reqBody := strings.NewReader("ip_port=" + viewNotExist)
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	log.Println(reqBody)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	// Hard coded View example for testing
	expectedBody := map[string]interface{}{
		"result": "Error",
		"msg":    viewNotExist + " is not in current view",
	}

	equals(t, expectedBody, gotBody)

}

func TestPutHandlerStoresPayload(t *testing.T) {
	// Setup the test
	serverURL, router := setup(keyExists, valExists)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	// Start with a payload map
	testPayload := map[string]int{
		keyExists:    1,
		keyNotExists: 2,
	}
	testPayloadByte, err := json.Marshal(testPayload)
	ok(t, err)
	testPayloadString := string(testPayloadByte[:])

	testBody := "val=" + valExists + "&payload=" + testPayloadString

	// Convert it to a []byte
	reqBody := strings.NewReader(testBody)

	// Stub a request
	method := http.MethodPut
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)
	testPayload[keyExists] = 2

	expectedStatus := http.StatusCreated // code 201
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	_, err = ioutil.ReadAll(recorder.Body)
	ok(t, err)

	equals(t, testPayload, testKVS.dbClock)

	teardown()
}

func TestGetHandlerStalePayload(t *testing.T) {

	// Setup the test
	serverURL, router := setup(keyExists, valExists)

	// Use a httptest recorder to observe responses
	recorder := httptest.NewRecorder()

	// This subject exists in the store already
	subject := keyExists

	// Set up the URL
	url := serverURL + rootURL + "/" + subject

	testPayload := map[string]int{keyExists: 2}
	testPayloadByte, err := json.Marshal(testPayload)
	ok(t, err)
	testPayloadString := string(testPayloadByte[:])

	testBody := "payload=" + testPayloadString

	// Convert it to a []byte
	reqBody := strings.NewReader(testBody)
	// Stub a request
	method := http.MethodGet
	req, err := http.NewRequest(method, url, reqBody)
	ok(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Finally, make the request to the function being tested.
	router.ServeHTTP(recorder, req)

	expectedStatus := http.StatusBadRequest // code 200
	gotStatus := recorder.Code
	equals(t, expectedStatus, gotStatus)
	body, err := ioutil.ReadAll(recorder.Body)
	ok(t, err)

	var gotBody map[string]interface{}
	expectedPayload := map[string]interface{}{keyExists: float64(2)}

	err = json.Unmarshal(body, &gotBody)
	ok(t, err)
	expectedBody := map[string]interface{}{
		"result":  "Error",
		"msg":     "Payload out of date",
		"payload": expectedPayload,
	}
	if diff := deep.Equal(expectedBody, gotBody); diff != nil {
		t.Error(diff)
	}
	equals(t, expectedBody, gotBody)

	teardown()
}

// These functions were taken from Ben Johnson's post here: https://medium.com/@benbjohnson/structuring-tests-in-go-46ddee7a25c

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033%s:%d: "+msg+"\033\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033%s:%d: unexpected error: %s\033\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
