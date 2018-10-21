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
	Contains(string) bool
	Count() int
	Get(string) string
	Delete(string) bool
	Put(string, string)
}

/* this is an example of how this interface can be implemented
type kvs struct {
	db map[string]string
}

func (k *kvs) Contains(key string) bool {
	_, exist := k.db[key]
	return exist
}

func (k *kvs) Count() int {
	return len(k.db)
}

func (k *kvs) Get(key string) (string, bool) {
	val, exist := k.db[key]
	return val, exist
}

func (k *kvs) Delete(key string) bool {
	delete(k.db, key)
	return true
}

func (k *kvs) Put(key, val string) bool {
	k.db[key] = val
	return true
}

*/
