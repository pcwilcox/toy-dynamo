package main

import (
	"testing"
)

const (
	keyone    = "keyone"
	valone    = "valone"
	deletekey = "IDONTEXIST"
)

func TestKVSContains(t *testing.T) {
	k := NewKVS()
	if k.Contains(keyone) != false {
		t.Fatalf("WTF")
	}
}

func TestKVSPut(t *testing.T) {
	k := NewKVS()
	if k.Put(keyone, valone) != true {
		t.Fatalf("WTF")
	}
}

func TestKVSDelete(t *testing.T) {
	k := NewKVS()
	if k.Delete(deletekey) != false {
		t.Fatalf("WTF")
	}
}

func TestKVSGet(t *testing.T) {
	k := NewKVS()
	k.Put(keyone, valone)

	if k.Get(keyone) != valone {
		t.Fatalf("WTF")
	}
}

func TestKVSServiceUp(t *testing.T) {
	k := NewKVS()
	if k.ServiceUp() != true {
		t.Fatalf("WTF")
	}
}
