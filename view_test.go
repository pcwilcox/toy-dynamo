// view_test.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson   lelawson
// Pete Wilcox       pcwilcox
// Annie Shen        ashen7
// Victoria Tran
//
// Unit tests for the viewlist struct

package main

import (
	"fmt"
	"testing"
)

// NewView should initialize a new view list
func TestNewViewSetsInitialView(t *testing.T) {
	v := NewView(testMain, testView)
	equals(t, 3, v.Count())
	equals(t, true, v.Contains("176.32.164.10:8082"))
	equals(t, true, v.Contains("176.32.164.10:8083"))
}

// String should print a comma-separated list
func TestViewListString(t *testing.T) {
	v := NewView(testMain, testView)
	equals(t, testView, fmt.Sprint(v))
}

// Primary should return the main IP:Port
func TestPrimaryReturnsIP(t *testing.T) {
	v := NewView(testMain, testView)
	equals(t, testMain, v.Primary())
}

// Random should return no more than N elements
func TestRandomReturnsNItems(t *testing.T) {
	v := NewView(testMain, testView)
	u := v.Random(1)
	equals(t, 1, len(u))
}

// Random should return no more than the number of items in the view - 1
func TestRandomReturnsMaxItems(t *testing.T) {
	v := NewView(testMain, testView)
	u := v.Random(5)
	equals(t, 2, len(u))
}

// Random should not return the primary
func TestRandomDoesntReturnPrimary(t *testing.T) {
	v := NewView(testMain, testView)
	u := v.Random(5)
	for _, k := range u {
		assert(t, k != testMain, "Primary returned by Random()")
	}
}

// Add should insert another element
func TestAddInsertsAnotherElement(t *testing.T) {
	v := NewView(testMain, testView)
	assert(t, v.Add("172.32.164.10:8085"), "Add returned false")
	assert(t, v.Count() > 3, "Add didn't increase the size of the map")
}

// Remove should remove an element
func TestRemoveDeletesElement(t *testing.T) {
	v := NewView(testMain, testView)
	assert(t, v.Remove("172.32.164.10:8082"), "Delete returned false")
	assert(t, !v.Contains("172.32.164.10:8082"), "Delete didn't remove item from the map")
}

// Contains should return true if the item is there
func TestContainsItemExists(t *testing.T) {
	v := NewView(testMain, testView)
	assert(t, v.Contains(testMain), "Contains returned false")
}

// Contains should return false if the item isn't there
func TestContainsItemNotExists(t *testing.T) {
	v := NewView(testMain, testView)
	assert(t, !v.Contains("172.32.164.10:8090"), "Contains returned true")
}

// Count should return the number of items
func TestCountReturnsNum(t *testing.T) {
	v := NewView(testMain, testView)
	assert(t, v.Count() == 3, "Count was wrong")
}
