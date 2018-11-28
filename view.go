// view.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson  lelawson
// Pete Wilcox      pcwilcox
// Annie Shen       ashen7
// Victoria Tran    vilatran
//
// Defines an interface and struct for maintaining the view of the system.
//

package main

import (
	"sort"
	"strings"
)

// A View maintains a list of IP:Port pairs as its view of the system configuration and implements methods for modifying it
type View interface {
	// Return the number of items in the view
	Count() int

	// Return true if the given string exists as an item in the view
	Contains(string) bool

	// Removes an item from the view, returns true if successful
	Remove(string) bool

	// Adds an item to the view, returns true if successful
	Add(string) bool

	// Given N, returns a slice of length <= N with items randomly selected from the view
	Random(int) []string

	// Returns the item associated with this server
	Primary() string

	// Overwrite is used by gossiping to overwrite the view
	Overwrite([]string)

	// List just gives a []string of our views to make it easy to gossip
	List() []string

	// String prints it
	String() string
}

// A viewList is a struct which implements the View interface and holds the view of the server configs
type viewList struct {
	views   map[string]string // This is a map because it gives O(1) lookups
	primary string            // This is the server we're actually on
}

// List spits out a byte slice
func (v *viewList) List() []string {
	if v != nil {
		var s []string
		for k := range v.views {
			s = append(s, k)
		}
		return s
	}
	return nil
}

// Overwrite simply replaces the view list
func (v *viewList) Overwrite(n []string) {
	if v != nil && v.views != nil {
		diff := false
		if len(n) == len(v.views) {
			for _, val := range n {
				if !v.Contains(val) {
					diff = true
				}
			}
		} else {
			diff = true
		}
		if diff {
			for k := range v.views {
				delete(v.views, k)
			}
			for _, k := range n {
				v.views[k] = k
			}
			viewChange = true
		}
	}
}

// Count returns the number of elements in the view list
func (v *viewList) Count() int {
	if v != nil {
		return len(v.views)
	}
	return 0
}

// Contains returns true if the viewList contains a particular item
func (v *viewList) Contains(item string) bool {
	if v != nil {
		_, ok := v.views[item]
		return ok
	}
	return false
}

// Remove deletes an item from the view
func (v *viewList) Remove(item string) bool {
	if v != nil {
		delete(v.views, item)
		viewChange = true
		return true
	}
	return false
}

// Add inserts an item into the view
func (v *viewList) Add(item string) bool {
	if v != nil {
		v.views[item] = item
		viewChange = true
		return true
	}
	return false
}

// Random picks up to N random elements and returns them as a slice (up to because it'll max out at the number of items available)
func (v *viewList) Random(n int) []string {
	if v != nil {
		var m int
		// The limit here is len()-1 because we don't want to return the primary
		if len(v.views)-1 > n {
			m = n
		} else {
			m = len(v.views) - 1
		}
		var items []string

		for _, k := range v.views {
			if k != v.primary {
				items = append(items, k)
			}
		}
		return items[0:m]
	}
	return nil
}

// Primary returns the actual server we're on
func (v *viewList) Primary() string {
	if v != nil {
		return v.primary
	}
	return ""
}

// String converts the view into a comma-separated string
func (v *viewList) String() string {
	if v != nil {
		var items []string
		for _, k := range v.views {
			items = append(items, k)
		}
		sort.Strings(items)
		str := ""
		i := len(items)
		j := 0
		for ; j < i-1; j++ {
			str = str + items[j] + ","
		}
		str = str + items[j]
		return str
	}
	return ""
}

// NewView creates a viewlist object and initializes it with the input string
func NewView(main string, input string) *viewList {
	// Make a new map
	v := make(map[string]string)

	// Convert the input string into a slice
	slice := strings.Split(input, ",")

	// Insert each element of the slice into the map
	for _, s := range slice {
		v[s] = s
	}

	list := viewList{
		views:   v,
		primary: main,
	}
	return &list
}
