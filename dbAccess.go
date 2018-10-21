/* dbAccess.go
 *
 * CMPS 128 Fall 2018
 *
 * Lawrence Lawson     lelawson
 * Pete Wilcox         pcwilcox
 *
 * Defines a DB interface which is used to interact between the REST API front end and the KVS back end.
 */

package main

// dbAccess interface defines methods for interactions between the REST API front end and the key-value store
type dbAccess interface {

	// Contains returns true if the dbAccess object contains an object with key equal to the input
	Contains(string) bool

	// Get returns the value associated with a particular key. If the key does not exist it returns ""
	Get(string) string

	// Delete removes a key-value pair from the object. If the key does not exist it returns false.
	Delete(string) bool

	// Put adds a key-value pair to the DB. If the key already exists, then it overwrites the existing value. If the key does not exist then it is added.
	Put(string, string) bool

	// ServiceUp returns true if the interface is able to communicate with the DB
	ServiceUp() bool
}
