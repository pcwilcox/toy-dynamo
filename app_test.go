/* app_test.go
 *
 * CMPS 128 Fall 2018
 *
 * Lawrence Lawson     lelawson
 * Pete Wilcox         pcwilcox
 *
 * Unit test definitions for app.go
 */

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// Define some constants. These can be reconfigured as needed.
const (
	DOMAIN   = "http://localhost"
	PORT     = "8080"
	ROOT     = "/keyValue-store"
	HOSTNAME = DOMAIN + ":" + PORT + ROOT
	KEY      = "KEY_EXISTS"
	VAL      = "VAL_EXISTS"
)

type TestKVS struct {
	key   string
	val   string
	count int
}

/* This stub returns true for the key which exists and false for the one which doesn't */
func (t *TestKVS) Contains(key string) bool {
	if key == t.key {
		return true
	}
	return false
}

/* This stub just returns 1 */
func (t *TestKVS) Count() int {
	return t.count
}

/* This stub returns the value associated with the key which exists, and returns nil for the key which doesn't */
func (t *TestKVS) Get(key string) string {
	if key == t.key {
		return t.val
	}
	return ""
}

/* This stub returns true for the key which exists and false for the one which doesn't */
func (t *TestKVS) Delete(key string) bool {
	if key == t.key {
		t.count--
		return true
	}
	return false
}

// idk lets try this
func (t *TestKVS) Put(key, val string) bool {
	return true
}

// Responds to PUT requests on http://localhost:IP/keyValue-Store/{subject} with message body "val={value}"
func TestPutRequest(t *testing.T) {
	// Stub the db
	db := TestKVS{KEY, VAL, 1}

	// Stub the app
	app := App{&db, ":8080"}

	/* Stub the handler */
	handler := http.HandlerFunc(app.PutHandler)

	/* Use a httptest recorder to observe responses */
	recorder := httptest.NewRecorder()

	/*********************************
	 * FIRST TEST:
	 * 'subject' exist in the db
	 *********************************/

	/* This subject exists in the store already */
	subject := KEY

	/* Set up the URL */
	url := HOSTNAME + "/" + subject

	/* Stub a request */
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	ok(t, err)

	/* Finally, make the request to the function being tested. */
	handler.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK
	var expectedBody string
	trythis, _ := json.Marshal(map[string]string{"replaced": "True", "msg": "Updated successfully"})
	err = json.Unmarshal(trythis, expectedBody)

	gotStatus := recorder.Code
	gotBody := recorder.Body.String()

	equals(t, expectedBody, gotBody)

	equals(t, expectedStatus, gotStatus)
}

/* These functions were taken from Ben Johnson's post here:
 * https://medium.com/@benbjohnson/structuring-tests-in-go-46ddee7a25c
 */

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
