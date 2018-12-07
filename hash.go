// hash.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson   lelawson
// Pete Wilcox       pcwilcox
// Annie Shen        ashen7
// Victoria Tran     vilatran
//
// Module to hash servers and keys

package main

import (
	"hash/crc32"
)

var primes []int

// Given a key, hashes it to the ring position
func getKeyPosition(key string) int {
	return int(crc32.ChecksumIEEE([]byte(key))) % ringSize
}

// Given a shard ID and a number of nodes, hashes it to a number of ring positions
func getVirtualNodePositions(shard string) []int {
	s := string(shard)

	if len(primes) < numVirtualNodes {
		makePrimes()
	}

	positions := make([]int, numVirtualNodes)
	for i := 0; i < numVirtualNodes; i++ {
		positions[i] = (int(crc32.ChecksumIEEE([]byte(s))) * primes[i]) % ringSize
	}
	return positions
}

// Generates seeds for the position hash
func makePrimes() {
	N := 1000
	numbers := make([]bool, N)

	for i := range numbers {
		numbers[i] = true
	}
	numbers[0] = false
	numbers[1] = false

	for x := 2; x < N; x++ {
		if numbers[x] {
			for i := x * x; i < N; i += x {
				numbers[i] = false
			}
		}
	}

	for i := range numbers {
		if numbers[i] {
			primes = append(primes, i)
		}
	}
}
