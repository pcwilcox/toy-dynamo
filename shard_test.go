// shard_test.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson   lelawson
// Pete Wilcox       pcwilcox
// Annie Shen        ashen7
// Victoria Tran     vilatran
//
// Unit tests for the shardring struct

package main

import "testing"

/*

Pause the test_shard.go
Take care of all shard.go and the NewShard first then come back.

*/

// func TestShardInitialStart(t *testing.T) {
// 	s := NewShard(testMain, testView)
// 	equals(t, 3, v.Count())
// 	equals(t, true, v.Contains("176.32.164.10:8082"))
// 	equals(t, true, v.Contains("176.32.164.10:8083"))
// }

func TestShardInitalStart(t *testing.T) {
	s := NewShard(testMain, testView, 1)
	equals(t)
}

// funct TestViewListCount(t *testing.T) {
// 	v := NewShard(testMain, testView, test)
// }
