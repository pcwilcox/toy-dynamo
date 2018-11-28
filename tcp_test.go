// tcp_test.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson   lelawson
// Pete Wilcox       pcwilcox
// Annie Shen        ashen7
// Victoria Tran     vilatran
//
// Unit tests for the TCP communication module

package main

import (
	"bufio"
	"reflect"
	"testing"

	"github.com/go-test/deep"
)

func testHandlerFunc(rw *bufio.ReadWriter) {}

func TestNewEndpointMakesEndpoint(t *testing.T) {
	e := NewEndpoint()
	assert(t, e != nil, "Endpoint not created")
	assert(t, e.handler != nil, "Endpoint handler map not created")
}

func TestAddHandleFuncAddsFunction(t *testing.T) {
	e := NewEndpoint()
	e.AddHandleFunc("test", testHandlerFunc)
	k, v := e.handler["test"]
	var f = testHandlerFunc
	ak := reflect.ValueOf(&k).Elem()
	af := reflect.ValueOf(&f).Elem()
	assert(t, v, "Handler key not input correctly")
	if diff := deep.Equal(ak, af); diff != nil {
		t.Error(diff)
	}
}
