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
	"log"
)

var primes []int

// Given a key, hashes it to the ring position
func getKeyPosition(key string) int {
	log.Println("Hashing key ", key)
	i := int(crc32.ChecksumIEEE([]byte(key))) % ringSize
	log.Println("Hashes to ", i)
	return i
}

// Given a shard ID and a number of nodes, hashes it to a number of ring positions
func getVirtualNodePositions(shard string) []int {
	log.Println("Getting node positions for shard ", shard)

	if len(primes) < numVirtualNodes {
		makePrimes()
	}

	positions := make([]int, numVirtualNodes)
	for i := 0; i < numVirtualNodes; i++ {
		positions[i] = (int(crc32.ChecksumIEEE([]byte(shard))) * primes[i]) % ringSize
	}
	log.Println("positions computed: ", positions)
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
