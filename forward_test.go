// forward_test.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson           lelawson
// Pete Wilcox               pcwilcox
//
// This file contains unit tests for forward.go
//

package main

import (
	"net/http"
	"testing"
)

const (
	keyExists    = "KEY_EXISTS"
	keyNotExists = "KEY_NOT_EXISTS"
	valExists    = "VAL_EXISTS"
	valNotExists = "VAL_NOT_EXISTS"
)

type testRest struct {
}

func (t *testRest) Initialize() {

}

func (t *testRest) PutHandler(w http.ResponseWriter, r *http.Request) {

}
func (t *testRest) GetHandler(w http.ResponseWriter, r *http.Request) {

}

func (t *testRest) DeleteHandler(w http.ResponseWriter, r *http.Request) {

}

func (t *testRest) ServiceDownHandler(http.ResponseWriter, *http.Request) {

}
func TestForwarderContainsKeyExists(t *testing.T) {

}

func TestForwarderContainsKeyNotExists(t *testing.T) {

}

func TestForwarderContainsServiceDown(t *testing.T) {

}

func TestForwarderPutKeyExists(t *testing.T) {

}

func TestForwarderPutKeyNotExists(t *testing.T) {

}

func TestForwarderPutKeyInvalid(t *testing.T) {

}

func TestForwarderPutValInvalid(t *testing.T) {

}

func TestForwarderPutServiceDown(t *testing.T) {

}
func TestForwarderDeleteKeyExists(t *testing.T) {

}

func TestForwarderDeleteKeyNotExists(t *testing.T) {

}

func TestForwarderDeleteServiceDown(t *testing.T) {

}
func TestForwarderGetKeyExists(t *testing.T) {

}

func TestForwarderGetKeyNotExists(t *testing.T) {

}

func TestForwarderGetServiceDown(t *testing.T) {

}

func TestServiceDownWhenServiceUp(t *testing.T) {

}

func TestServiceDownWhenServiceDown(t *testing.T) {

}
