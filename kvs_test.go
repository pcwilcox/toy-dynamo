package main

import (
	"strings"
	"sync"
	"testing"
)

const (
	keyone     = "Key One"
	valone     = "Value One"
	keyNotHere = "Key Not Here"
	valtwo     = "Value Two"
	deletekey  = "I DONT EXIST"
	invalidKey = `Lorem ipsum dolor sit amet, 
			consectetuer adipiscing elit. Aenean commodo 
			ligula eget dolor. Aenean massa. Cum sociis natoque 
			penatibus et magnis dis parturient montes, 
			nascetur ridiculus mus. Donec qu`
	invalidVal = ""
)

func TestKVSContainsCheckIfDoesntExist(t *testing.T) {
	db := map[string]string{
		keyone: valone,
	}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}

	assert(t, !k.Contains(keyNotHere), "Key found that does not exist.")
}
func TestKVSContainsCheckIfDoesExist(t *testing.T) {
	db := map[string]string{
		keyone: valone,
	}

	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, k.Contains(keyone), "Key not found that does exist.")

}

func TestKVSGetExistingVal(t *testing.T) {
	db := map[string]string{
		keyone: valone,
	}

	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	equals(t, valone, k.Get(keyone))
}

func TestKVSGetValDoesntExist(t *testing.T) {
	db := map[string]string{
		keyone: valone,
	}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	equals(t, "", k.Get(keyNotHere))
}

func TestKVSDeleteExistingKeyValPair(t *testing.T) {
	db := map[string]string{
		keyone: valone,
	}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, k.Delete(keyone), "Did not delete Key Val Pair")
}

func TestKVSDeleteKeyDoesntExist(t *testing.T) {
	db := map[string]string{
		keyone: valone,
	}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, !k.Delete(keyNotHere), "Deleted a keyvalue pair not in data store prior")
}

func TestKVSPutNewKeyNewVal(t *testing.T) {
	k := NewKVS()
	assert(t, k.Put(keyone, valone), "New key and value were not added")
}

func TestKVSPutExistKeyOverwriteVal(t *testing.T) {
	db := map[string]string{
		keyone: valone,
	}
	var m sync.RWMutex
	k := KVS{db: db, mutex: &m}
	assert(t, k.Put(keyone, valtwo), "Did not overwrite existing key's value")

}

func TestKVSPutInvalidkey(t *testing.T) {
	k := NewKVS()
	assert(t, !k.Put(invalidKey, valtwo), "Invalid key added")

}

func TestKVSPutInvalidVal(t *testing.T) {
	k := NewKVS()
	var b strings.Builder
	b.Grow(1048577)
	for i := 0; i < 1048577; i++ {
		b.WriteByte(0)
	}
	invalidVal := b.String()
	assert(t, !k.Put(keyone, invalidVal), "Invalid value added")
}

func TestKVSServiceUp(t *testing.T) {
	k := NewKVS()
	if k.ServiceUp() != true {
		t.Fatalf("WTF")
	}
}

// These functions were taken from Ben Johnson's post here: https://medium.com/@benbjohnson/structuring-tests-in-go-46ddee7a25c

// // assert fails the test if the condition is false.
// func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
// 	if !condition {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
// 		tb.FailNow()
// 	}
// }

// // ok fails the test if an err is not nil.
// func ok(tb testing.TB, err error) {
// 	if err != nil {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
// 		tb.FailNow()
// 	}
// }

// // equals fails the test if exp is not equal to act.
// func equals(tb testing.TB, exp, act interface{}) {
// 	if !reflect.DeepEqual(exp, act) {
// 		_, file, line, _ := runtime.Caller(1)
// 		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
// 		tb.FailNow()
// 	}
// }
