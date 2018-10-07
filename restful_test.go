/*-
 * restful_test.go
 *
 * Pete Wilcox
 * CruzID: pcwilcox
 * CMPS 128, Fall 2018
 *
 * This is the unit test suite for restful.go
 *
 */
package restful

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

/* Define some constants. These can be reconfigured as needed. */
const DOMAIN = "http://localhost"
const PORT = "8080"
const HOSTNAME = DOMAIN + ":" + PORT

func TestHelloHandler(t *testing.T) {
	/* Stub the handler */
	handler := http.HandlerFunc(HelloHandler)

	/* Use a httptest recorder to observe responses */
	recorder := httptest.NewRecorder()

	/*********************************
	 * FIRST TEST:
	 * Fixed response to GET request
	 *********************************/

	/* Set up the URL */
	url := HOSTNAME + "/hello"

	/* Stub a request */
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Error(err)
	}

	/* Finally, make the request to the function being tested. */
	handler.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK
	expectedBody := "Hello world!"

	gotStatus := recorder.Code
	gotBody := recorder.Body.String()

	if expectedStatus != gotStatus {
		t.Errorf("Expected status '%v', got status '%v'", expectedStatus, gotStatus)
	}

	if expectedBody != gotBody {
		t.Errorf("Expected body '%v', got body '%v'", expectedBody, gotBody)
	}

	/*************************************************
	 * SECOND TEST
	 * Any other request should get METHOD FORBIDDEN
	 *************************************************/

	/* test method types */
	methods := []string{
		"POST",
		"PUT",
		"DELETE",
	}

	/* separate test for each message */
	for _, method := range methods {

		/* set up the URL */
		url = HOSTNAME + "/hello"

		/* stub the request */
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			t.Error(err)
		}

		/* New recorder needed for each request */
		recorder = httptest.NewRecorder()

		/* Make the request and check the response */
		handler.ServeHTTP(recorder, req)

		expectedStatus = http.StatusMethodNotAllowed

		gotStatus = recorder.Code

		if expectedStatus != gotStatus {
			t.Errorf("Expected status '%v', got status '%v'", expectedStatus, gotStatus)
		}
	}
}

func TestTestHandler(t *testing.T) {
	/* Stub the handler */
	handler := http.HandlerFunc(TestHandler)

	/* Use a httptest recorder to observe responses */
	recorder := httptest.NewRecorder()

	/*********************************
	 * FIRST TEST:
	 * Fixed response to GET request
	 *********************************/

	/* Set up the URL */
	url := HOSTNAME + "/test"

	/* Stub a request */
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Error(err)
	}

	/* Finally, make the request to the function being tested. */
	handler.ServeHTTP(recorder, req)

	expectedStatus := http.StatusOK
	expectedBody := "GET request received"

	gotStatus := recorder.Code
	gotBody := recorder.Body.String()

	if expectedStatus != gotStatus {
		t.Errorf("Expected status '%v', got status '%v'", expectedStatus, gotStatus)
	}

	if expectedBody != gotBody {
		t.Errorf("Expected body '%v', got body '%v'", expectedBody, gotBody)
	}

	/*************************************************
	 * SECOND TEST:
	 * Response to POST request returns query message
	 *************************************************/

	/* test query strings */
	messages := []string{
		"'ACoolMessage'",
		"'Fight the power!'",
		"'Unit testing is fun'",
	}

	/* separate test for each message */
	for _, message := range messages {

		/* set up the URL */
		url = HOSTNAME + "/test?msg=" + message

		/* stub the request */
		method = "POST"
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			t.Error(err)
		}

		/* New recorder needed for each request */
		recorder = httptest.NewRecorder()

		/* Make the request and check the response */
		handler.ServeHTTP(recorder, req)

		expectedStatus = http.StatusOK
		expectedBody = "POST message received: " + message

		gotStatus = recorder.Code
		gotBody = recorder.Body.String()

		if expectedStatus != gotStatus {
			t.Errorf("Expected status '%v', got status '%v'", expectedStatus, gotStatus)
		}

		if expectedBody != gotBody {
			t.Errorf("Expected body '%v', got body '%v'", expectedBody, gotBody)
		}
	}

	/*************************************************
	 * THIRD TEST
	 * Any other request should get METHOD FORBIDDEN
	 *************************************************/

	/* test method types */
	methods := []string{
		"PUT",
		"DELETE",
	}

	/* separate test for each message */
	for _, method := range methods {

		/* set up the URL */
		url = HOSTNAME + "/test"

		/* stub the request */
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			t.Error(err)
		}

		/* New recorder needed for each request */
		recorder = httptest.NewRecorder()

		/* Make the request and check the response */
		handler.ServeHTTP(recorder, req)

		expectedStatus = http.StatusMethodNotAllowed

		gotStatus = recorder.Code

		if expectedStatus != gotStatus {
			t.Errorf("Expected status '%v', got status '%v'", expectedStatus, gotStatus)
		}
	}

	/*************************************************
	 * THIRD TEST
	 * Any other request should get METHOD FORBIDDEN
	 *************************************************/

}
