// rbtree.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson   lelawson
// Pete Wilcox       pcwilcox
// Annie Shen        ashen7
// Victoria Tran     vilatran
//
// Red-Black tree used for ring position lookups

package main

import (
	"sync"
)

// RBTree is a red-black tree structure which balances itself in order
// to maintain O(logn) insert, delete, search times. It uses integers
// for keys and strings for values.
type RBTree struct {
	Root  *RBNode // the root RBNode
	Mutex sync.RWMutex
}

// Colors for nodes
const (
	red   = true
	black = false
)

// Returns number of keys in the tree
func (r *RBTree) size() int {
	if r != nil {
		r.Mutex.RLock()
		defer r.Mutex.RUnlock()
		return r.Root.size()
	}
	return 0
}

// Returns value associated with a particular key if it exists
func (r *RBTree) get(key int) string {
	if r != nil {
		r.Mutex.RLock()
		defer r.Mutex.RUnlock()
		return r.Root.get(key)
	}
	return ""
}

// Inserts a value associated with a particular key
func (r *RBTree) put(key int, value string) {
	if r != nil {
		r.Mutex.Lock()
		defer r.Mutex.Unlock()
		if r.Root == nil {
			r.Root = newNode(key, value, black)
		} else {
			r.Root = r.Root.put(key, value)
			r.Root.Color = black
		}
	}
}

// Returns the value of the smallest key
func (r *RBTree) min() string {
	if r != nil {
		r.Mutex.RLock()
		defer r.Mutex.RUnlock()
		return r.Root.min().Value
	}
	return ""
}

// Returns key X such that X is the largest key which is less than
// or equal to the input
func (r *RBTree) floor(key int) int {
	if r != nil {
		r.Mutex.RLock()
		defer r.Mutex.RUnlock()
		x := r.Root.floor(key)
		if x != nil {
			return x.Key
		}
	}
	return -1
}

// Returns key X such that X is the k-th smallest key
func (r *RBTree) selection(k int) int {
	if r != nil {
		r.Mutex.RLock()
		defer r.Mutex.RUnlock()
		x := r.Root.selection(k)
		if x != nil {
			return x.Key
		}
	}
	return -1
}

// Returns the number of keys smaller than k
func (r *RBTree) rank(k int) int {
	if r != nil {
		r.Mutex.RLock()
		defer r.Mutex.RUnlock()
		return r.Root.rank(k)
	}
	return -1
}

// Deletes the node with the minimum key
func (r *RBTree) deleteMin() {
	if r != nil {
		r.Mutex.Lock()
		defer r.Mutex.Unlock()
		if !r.Root.Left.isRed() && !r.Root.Right.isRed() {
			r.Root.Color = red
		}
		r.Root = r.Root.deleteMin()
		if r.Root != nil {
			r.Root.Color = black
		}
	}
}

// Deletes the node with key k
func (r *RBTree) delete(k int) {
	if r != nil {
		r.Mutex.Lock()
		defer r.Mutex.Unlock()
		if r.Root != nil && !r.Root.Left.isRed() && !r.Root.Right.isRed() {
			r.Root.Color = red
		}
		r.Root = r.Root.delete(k)
		if r.Root != nil {
			r.Root.Color = black
		}
	}
}

// Returns the value of the maximum key
func (r *RBTree) max() string {
	if r != nil && r.Root != nil {
		r.Mutex.RLock()
		defer r.Mutex.RUnlock()
		x := r.Root.max()
		return x.Value
	}
	return ""
}

// If there is a key k, returns its value. Otherwise, if there is a
// larger key, return that value. Otherwise, return the smallest value.
func (r *RBTree) successor(k int) string {
	if r != nil && r.Root != nil {
		r.Mutex.RLock()
		defer r.Mutex.RUnlock()
		x := r.Root.ceil(k)
		if x != nil {
			return x.Value
		}
		return r.min()
	}
	return ""
}

// If there is a key k, returns its value. Otherwise, if there is a
// smaller value, returns it. Otherwise, returns the largest value.
func (r *RBTree) predecessor(k int) string {
	if r != nil && r.Root != nil {
		r.Mutex.RLock()
		defer r.Mutex.RUnlock()
		x := r.Root.floor(k)
		if x != nil {
			return x.Value
		}
		return r.max()
	}
	return ""
}

// RBNode is a RBNode in a RBTree
type RBNode struct {
	Key    int     // hashed position on the ring
	Value  string  // shard ID
	Left   *RBNode // Left child
	Right  *RBNode // Right child
	Weight int     // number of RBNodes on this subtree
	Color  bool    // true = red
}

func (n *RBNode) max() *RBNode {
	if n == nil {
		return nil
	}
	x := n.Right.max()
	if x != nil {
		return x
	}
	return n
}

func (n *RBNode) delete(k int) *RBNode {
	if n == nil {
		return nil
	}

	if k < n.Key {
		if !n.Left.isRed() && !n.Left.Left.isRed() {
			n = n.moveRedLeft()
		}
		n.Left = n.Left.delete(k)
	} else {
		if n.Left.isRed() {
			n = n.rotateRight()
		}
		if k == n.Key && n.Right == nil {
			return nil
		}
		if n.Right != nil && n.Right.Left != nil && !n.Right.isRed() && !n.Right.Left.isRed() {
			n = n.moveRedRight()
		}
		if k == n.Key {
			n.Value = n.Right.min().Value
			n.Key = n.Right.min().Key
			n.Right = n.Right.deleteMin()
		} else {
			n.Right = n.Right.delete(k)
		}
	}
	return n.balance()
}

func (n *RBNode) moveRedRight() *RBNode {

	if n == nil {
		return nil
	}
	n.flipColors()
	if n.Left != nil && !n.Left.Left.isRed() {
		n = n.rotateRight()
	}
	return n
}

func (n *RBNode) deleteMin() *RBNode {
	if n == nil {
		return nil
	}
	if n.Left == nil {
		return n.Right
	}

	if n.Left != nil && !n.Left.isRed() && !n.Left.Left.isRed() {
		n = n.moveRedLeft()
	}

	n.Left = n.Left.deleteMin()

	return n.balance()
}

func (n *RBNode) moveRedLeft() *RBNode {
	if n == nil {
		return nil
	}
	n.flipColors()
	if n.Right != nil && n.Right.Left.isRed() {
		n.Right = n.Right.rotateRight()
		n = n.rotateLeft()
	}
	return n
}

func (n *RBNode) rank(k int) int {
	if n == nil {
		return 0
	}
	if k < n.Key {
		return n.Left.rank(k)
	} else if k > n.Key {
		return 1 + n.Left.size() + n.Right.rank(k)
	} else {
		return n.Left.size()
	}
}

func (n *RBNode) selection(k int) *RBNode {
	if n == nil {
		return nil
	}
	t := n.Left.size()
	if t > k {
		return n.Left.selection(k)
	} else if t < k {
		return n.Right.selection(k - t - 1)
	} else {
		return n
	}
}

func (n *RBNode) floor(key int) *RBNode {
	if n == nil {
		return nil
	}
	if key == n.Key {
		return n
	}
	if key < n.Key {
		return n.Left.floor(key)
	}

	x := n.Right.floor(key)
	if x != nil {
		return x
	}
	return n
}

func (n *RBNode) ceil(key int) *RBNode {
	if n == nil {
		return nil
	}
	if key == n.Key {
		return n
	}
	if key > n.Key {
		return n.Right.ceil(key)
	}
	x := n.Left.ceil(key)
	if x != nil {
		return x
	}
	return n
}

func (n *RBNode) min() *RBNode {
	if n != nil {
		if n.Left == nil {
			return n
		}
		return n.Left.min()
	}
	return nil
}

func (n *RBNode) put(key int, value string) *RBNode {
	if n != nil {
		if key < n.Key {
			n.Left = n.Left.put(key, value)
		} else if key > n.Key {
			n.Right = n.Right.put(key, value)
		} else {
			n.Value = value
		}
		if n.Right.isRed() && !n.Left.isRed() {
			n = n.rotateLeft()
		}
		if n.Left.isRed() && n.Left.Left.isRed() {
			n = n.rotateRight()
		}
		if n.Left.isRed() && n.Right.isRed() {
			n.flipColors()
		}
	} else {
		n = newNode(key, value, red)
	}
	n.Weight = 1 + n.Left.size() + n.Right.size()
	return n
}

func newNode(key int, value string, color bool) *RBNode {
	n := RBNode{
		Key:    key,
		Value:  value,
		Left:   nil,
		Right:  nil,
		Weight: 1,
		Color:  color,
	}
	return &n
}

func (n *RBNode) get(key int) string {
	if n != nil {
		if key < n.Key {
			return n.Left.get(key)
		}
		if key > n.Key {
			return n.Right.get(key)
		}
		return n.Value
	}
	return ""
}

func (n *RBNode) isRed() bool {
	if n != nil {
		return n.Color
	}
	return false
}

func (n *RBNode) rotateLeft() *RBNode {
	if n != nil && n.Right != nil {
		x := n.Right
		n.Right = x.Left
		x.Left = n
		x.Color = n.Color
		n.Color = red
		x.Weight = n.Weight
		n.Weight = 1 + n.Left.size() + n.Right.size()
		return x
	}
	return nil
}

func (n *RBNode) rotateRight() *RBNode {
	if n != nil && n.Left != nil {
		x := n.Left
		n.Left = x.Right
		x.Right = n
		x.Color = n.Color
		n.Color = red
		x.Weight = n.Weight
		n.Weight = 1 + n.Left.size() + n.Right.size()
		return x
	}
	return nil
}

func (n *RBNode) flipColors() {
	if n != nil {
		n.Color = !n.Color
		if n.Left != nil {
			n.Left.Color = !n.Left.Color
		}
		if n.Right != nil {
			n.Right.Color = !n.Right.Color
		}
	}
}

func (n *RBNode) size() int {
	if n != nil {
		return n.Weight
	}
	return 0
}

func (n *RBNode) balance() *RBNode {
	if n.Right.isRed() {
		n = n.rotateLeft()
	}
	if n.Left != nil && n.Left.isRed() && n.Left.Left.isRed() {
		n = n.rotateRight()
	}
	if n.Left.isRed() && n.Right.isRed() {
		n.flipColors()
	}
	n.Weight = 1 + n.Left.size() + n.Right.size()
	return n
}
