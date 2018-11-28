// view_test.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson   lelawson
// Pete Wilcox       pcwilcox
// Annie Shen        ashen7
// Victoria Tran     vilatran
//
// Unit tests for the viewlist struct

package main

import (
	"fmt"
	"strings"
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

// List should return a byte slice of the view
func TestListReturnsExistingViews(t *testing.T) {
	v := NewView(testMain, testView)
	r := strings.Split(testView, ",")
	s := v.List()

	match := false
	for _, u := range r {
		for _, w := range s {
			if u == w {
				match = true
			}
		}
		assert(t, match, "Element in input not returned in output")
	}
}

// Overwrite should completely overwrite the view stored
func TestOverwriteWorks(t *testing.T) {
	v := NewView(testMain, testView)
	newTestView := []string{"172.132.164.20:8081", "172.132.164.20:8082", "172.132.164.20:8083"}
	m := make(map[string]string)
	for _, s := range newTestView {
		m[s] = s
	}
	v.Overwrite(newTestView)
	assert(t, viewChange, "Overwrite did not set viewChange")
	equals(t, m, v.views)

	// Test that the 'diff' check works
	newTestView = []string{"172.132.164.20:8081", "172.132.164.20:8082"}
	n := make(map[string]string)
	for _, s := range newTestView {
		n[s] = s
	}

	v.Overwrite(newTestView)
	assert(t, viewChange, "Overwrite did not set viewChange")
	equals(t, n, v.views)
}

// Make sure we're handling a nil object
func TestListViewNilDoesntExplode(t *testing.T) {
	var v *viewList
	var a []string
	equals(t, a, v.List())
}

func TestOverwriteViewNilDoesntExplode(t *testing.T) {
	var v *viewList
	v.Overwrite([]string{"172.132.164.20:8081"})
	assert(t, true, "Overwrite blew up on nil")
}

func TestCountViewNilDoesntExplode(t *testing.T) {
	var v *viewList
	assert(t, v.Count() == 0, "Count blew up")
}

func TestContainsViewNilDoesntExplode(t *testing.T) {
	var v *viewList
	s := "172.132.164.20:8081"
	assert(t, !v.Contains(s), "Contains was wrong")
}

func TestRemoveViewNilDoesntExplode(t *testing.T) {
	var v *viewList
	s := "172.132.164.20:8081"
	assert(t, !v.Remove(s), "Remove blew up")
}

func TestAddViewNilDoesntExplode(t *testing.T) {
	var v *viewList
	s := "172.132.164.20:8081"
	assert(t, !v.Add(s), "Add blew up")
}

func TestRandomViewNilDoesntExplode(t *testing.T) {
	var v *viewList
	assert(t, v.Random(1) == nil, "Random blew up")
}

func TestPrimaryViewNilReturnsEmpty(t *testing.T) {
	var v *viewList
	assert(t, v.Primary() == "", "Primary blew up")
}

func TestStringViewNilReturnsEmpty(t *testing.T) {
	var v *viewList
	assert(t, v.String() == "", "String blew up")
}
