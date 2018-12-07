// rbtree.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson   lelawson
// Pete Wilcox       pcwilcox
// Annie Shen        ashen7
// Victoria Tran     vilatran
//
// unit tests for the red black tree

package main

import "testing"

func TestNodeSizeReturnsSize(t *testing.T) {
	var u *RBNode
	assert(t, u.size() == 0, "size failed on nil node")

	w := RBNode{
		Key:    1,
		Value:  "A",
		Color:  false,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	assert(t, w.size() == 1, "Size is wrong")

	x := RBNode{
		Key:    3,
		Value:  "B",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    2,
		Value:  "C",
		Color:  black,
		Weight: 3,
		Left:   &w,
		Right:  &x,
	}
	assert(t, y.size() == 3, "Size is wrong")
}

func TestNodeFlipColorsFlipsColors(t *testing.T) {
	var u *RBNode
	u.flipColors()

	w := RBNode{
		Key:    1,
		Value:  "A",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	w.flipColors()
	assert(t, w.isRed(), "Flip Colors didn't work")

	w.Color = black
	x := RBNode{
		Key:    3,
		Value:  "A",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    2,
		Value:  "A",
		Color:  red,
		Weight: 3,
		Left:   &w,
		Right:  &x,
	}
	y.flipColors()

	assert(t, x.isRed(), "flipcolors on right child failed")
	assert(t, w.isRed(), "flipcolors on left child failed")
	assert(t, !y.isRed(), "flipcolors didn't work on parent")
}

func TestNodeGet(t *testing.T) {
	var u *RBNode
	assert(t, u.get(1) == "", "Get failed on nil node")

	w := RBNode{
		Key:    1,
		Value:  "A",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	assert(t, w.get(1) == "A", "get failed to return value")

	x := RBNode{
		Key:    3,
		Value:  "AA",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    2,
		Value:  "BB",
		Color:  black,
		Weight: 3,
		Left:   &w,
		Right:  &x,
	}

	assert(t, y.get(1) == "A", "get failed to return left value")
	assert(t, y.get(2) == "BB", "get failed to return middle value")
	assert(t, y.get(3) == "AA", "get failed to return right value")
}

func TestIsRed(t *testing.T) {
	var u *RBNode
	assert(t, !u.isRed(), "isRed returned true for nil node")
	w := RBNode{
		Key:    1,
		Value:  "A",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	assert(t, !w.isRed(), "isRed returned true for black node")
	w.Color = true
	assert(t, w.isRed(), "isRed returned false for red node")
}

func TestNewNode(t *testing.T) {
	w := newNode(1, "A", black)
	equals(t, 1, w.Key)
	equals(t, "A", w.Value)
	equals(t, black, w.Color)
	x := newNode(3, "B", red)
	equals(t, 3, x.Key)
	equals(t, "B", x.Value)
	equals(t, red, x.Color)
	assert(t, x.Left == nil, "left child assigned")
	assert(t, x.Right == nil, "right child assigned")
}

func TestNodeMin(t *testing.T) {
	var u *RBNode
	assert(t, u.min() == nil, "min returned a non-nil value")
	w := RBNode{
		Key:    1,
		Value:  "A",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	equals(t, &w, w.min())

	y := RBNode{
		Key:    2,
		Value:  "B",
		Color:  black,
		Weight: 3,
		Left:   &w,
		Right:  nil,
	}

	equals(t, &w, y.min())
}

func TestNodeRotateRight(t *testing.T) {
	var u *RBNode
	assert(t, u.rotateRight() == nil, "rotateRight returned non-nil for nil node")

	w := RBNode{
		Key:    1,
		Value:  "A",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	assert(t, w.rotateRight() == nil, "rotateRight returned non-nil for a leaf node")

	x := RBNode{
		Key:    3,
		Value:  "B",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    2,
		Value:  "C",
		Color:  black,
		Weight: 3,
		Left:   &w,
		Right:  &x,
	}

	equals(t, &w, y.rotateRight())
	equals(t, true, y.Color)
	equals(t, false, w.Color)
	equals(t, &y, w.Right)
	equals(t, &x, y.Right)
	equals(t, 3, w.Weight)
	equals(t, 2, y.Weight)
}

func TestNodeRotateLeft(t *testing.T) {
	var u *RBNode
	assert(t, u.rotateLeft() == nil, "rotateLeft returned non-nil for nil node")

	w := RBNode{
		Key:    1,
		Value:  "A",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	assert(t, w.rotateLeft() == nil, "rotateLeft returned non-nil for a leaf node")

	x := RBNode{
		Key:    3,
		Value:  "B",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    2,
		Value:  "C",
		Color:  black,
		Weight: 3,
		Left:   &w,
		Right:  &x,
	}

	equals(t, &x, y.rotateLeft())
	equals(t, red, y.Color)
	equals(t, black, x.Color)
	equals(t, &y, x.Left)
	equals(t, &w, y.Left)
	equals(t, 3, x.Weight)
	equals(t, 2, y.Weight)
}

func TestNodePut(t *testing.T) {
	var u *RBNode
	u = u.put(20, "A")
	assert(t, u != nil, "put on a nil node returned nil")
	equals(t, 20, u.Key)
	equals(t, "A", u.Value)
	equals(t, red, u.Color)
	u.Color = black

	u = u.put(10, "AA")
	assert(t, u != nil, "put on a non-nil node returned nil")
	equals(t, 20, u.Key)
	equals(t, "A", u.Value)
	equals(t, 10, u.Left.Key)
	equals(t, "AA", u.Left.Value)
	equals(t, red, u.Left.Color)
	u.Color = black

	u = u.put(30, "CC")
	assert(t, u != nil, "put on a non-nil node returned nil")
	equals(t, 20, u.Key)
	equals(t, "A", u.Value)
	equals(t, 10, u.Left.Key)
	equals(t, "AA", u.Left.Value)
	equals(t, black, u.Left.Color)
	equals(t, 30, u.Right.Key)
	equals(t, "CC", u.Right.Value)
	equals(t, black, u.Right.Color)
	equals(t, red, u.Color)
	u.Color = black

	u = u.put(30, "D")
	equals(t, "D", u.Right.Value)

	u = u.put(40, "E")
	equals(t, black, u.Color)
	u.Color = black

	u = u.put(50, "F")
	equals(t, red, u.Left.Color)
	u.Color = black

	u = u.put(1, "alice")
	u.Color = black
	u = u.put(2, "bob")
	u.Color = black
	u = u.put(3, "carol")
	u.Color = black
	equals(t, red, u.Left.Right.Left.Color)
}

func TestNodeCeil(t *testing.T) {
	var u *RBNode
	assert(t, u.ceil(1) == nil, "ceil returned a non-nil value")
	w := RBNode{
		Key:    1,
		Value:  "alice",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	equals(t, &w, w.ceil(1))

	x := RBNode{
		Key:    5,
		Value:  "bob",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    3,
		Value:  "carol",
		Color:  black,
		Weight: 3,
		Left:   &w,
		Right:  &x,
	}

	equals(t, &x, y.ceil(4))
	equals(t, &w, y.ceil(0))
}

func TestNodeFloor(t *testing.T) {
	var u *RBNode
	assert(t, u.floor(1) == nil, "ceil returned a non-nil value")
	w := RBNode{
		Key:    1,
		Value:  "alice",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	equals(t, &w, w.floor(1))

	x := RBNode{
		Key:    5,
		Value:  "bob",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    3,
		Value:  "carol",
		Color:  black,
		Weight: 3,
		Left:   &w,
		Right:  &x,
	}

	equals(t, &w, y.floor(2))
	equals(t, &x, y.floor(100))
}

func TestNodeSelection(t *testing.T) {
	var u *RBNode
	assert(t, u.selection(1) == nil, "selection returned a non-nil value")

	w := RBNode{
		Key:    1,
		Value:  "alice",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	equals(t, &w, w.selection(0))

	x := RBNode{
		Key:    5,
		Value:  "bob",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    3,
		Value:  "carol",
		Color:  black,
		Weight: 3,
		Left:   &w,
		Right:  &x,
	}

	equals(t, &w, y.selection(0))
	equals(t, &y, y.selection(1))
	equals(t, &x, y.selection(2))
}

func TestNodeRank(t *testing.T) {
	var u *RBNode
	assert(t, u.rank(1) == 0, "rank returned incorrect number of keys")

	w := RBNode{
		Key:    1,
		Value:  "alice",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	equals(t, 0, w.rank(1))

	x := RBNode{
		Key:    5,
		Value:  "bob",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    3,
		Value:  "carol",
		Color:  black,
		Weight: 3,
		Left:   &w,
		Right:  &x,
	}

	equals(t, 0, y.rank(1))
	equals(t, 1, y.rank(2))
	equals(t, 2, y.rank(5))
	equals(t, 3, y.rank(6))
}

func TestNodeMoveRedLeft(t *testing.T) {
	var u *RBNode
	assert(t, u.moveRedLeft() == nil, "moveRedLeft returned non-nil")

	w := RBNode{
		Key:    1,
		Value:  "alice",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	equals(t, &w, w.moveRedLeft())

	x := RBNode{
		Key:    5,
		Value:  "bob",
		Color:  red,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    3,
		Value:  "charles",
		Color:  black,
		Weight: 3,
		Left:   nil,
		Right:  nil,
	}

	v := RBNode{
		Key:    6,
		Value:  "david",
		Color:  red,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	u = &RBNode{
		Key:    7,
		Value:  "eddie",
		Color:  black,
		Weight: 2,
		Left:   &v,
		Right:  nil,
	}

	x.Left = &y
	y.Left = &w
	x.Right = u

	z := x.moveRedLeft()
	equals(t, red, z.Left.Color)
	equals(t, black, z.Color)
	equals(t, 6, z.Key)
}

func TestNodeMoveRedRight(t *testing.T) {
	var u *RBNode
	assert(t, u.moveRedRight() == nil, "moveRedRight returned non-nil")

	w := RBNode{
		Key:    10,
		Value:  "alice",
		Color:  red,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	equals(t, &w, w.moveRedRight())
	w.Color = red

	x := RBNode{
		Key:    50,
		Value:  "bob",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    30,
		Value:  "carol",
		Color:  black,
		Weight: 3,
		Left:   nil,
		Right:  nil,
	}

	w.Right = &x
	x.Left = &y

	v := RBNode{
		Key:    1,
		Value:  "donnie",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	u = &RBNode{
		Key:    7,
		Value:  "eleanor",
		Color:  black,
		Weight: 2,
		Left:   &v,
		Right:  nil,
	}
	w.Left = u
	u.Left = &v

	z := w.moveRedRight()
	equals(t, black, z.Left.Color)
	equals(t, black, z.Color)
	equals(t, red, z.Right.Color) // this would seem to be an error except:
	// moveRedRight is called from delete() and the next line is to check if
	// its right child is red, in which case it rotates left
}

func TestNodeDeleteMin(t *testing.T) {
	var u *RBNode
	assert(t, u.deleteMin() == nil, "deleteMin returned non-nil for nil entry")

	w := RBNode{
		Key:    10,
		Value:  "adam",
		Color:  red,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	assert(t, w.deleteMin() == nil, "deleteMin returned non-nil for single node")
	w = RBNode{
		Key:    10,
		Value:  "billy",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	x := RBNode{
		Key:    50,
		Value:  "chuck",
		Color:  black,
		Weight: 1,
		Left:   &w,
		Right:  nil,
	}

	y := RBNode{
		Key:    30,
		Value:  "dean",
		Color:  black,
		Weight: 3,
		Left:   nil,
		Right:  nil,
	}

	w.Right = &y

	z := x.deleteMin()
	equals(t, z, &x)
	equals(t, 50, x.Key)
	equals(t, 30, x.Left.Key)

	w = RBNode{
		Key:    40,
		Value:  "eric",
		Color:  red,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	x.Left.Right = &w

	u = &RBNode{
		Key:   1,
		Value: "frank",
		Color: black,
		Left:  nil,
		Right: nil,
	}

	x.Left.Left = u

	z = x.deleteMin()
	equals(t, z, &x)
	equals(t, 50, x.Key)
	equals(t, 30, x.Left.Key)
	equals(t, 40, x.Left.Right.Key)

}

func TestNodeDelete(t *testing.T) {
	var u *RBNode
	assert(t, u.delete(0) == nil, "delete returned non-nil for nil entry")

	w := RBNode{
		Key:    10,
		Value:  "willy",
		Color:  red,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	assert(t, w.delete(100) == &w, "delete deleted the incorrect node")
	assert(t, w.delete(10) == nil, "delete returned non-nil for single node")

	v := RBNode{
		Key:    35,
		Value:  "xavier",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	r := RBNode{
		Key:    25,
		Value:  "tomas",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	y := RBNode{
		Key:    30,
		Value:  "ralph",
		Color:  black,
		Weight: 3,
		Left:   &r,
		Right:  &v,
	}

	w = RBNode{
		Key:    10,
		Value:  "brian",
		Color:  black,
		Weight: 4,
		Left:   nil,
		Right:  &y,
	}
	x := RBNode{
		Key:    50,
		Value:  "timmy",
		Color:  red,
		Weight: 5,
		Left:   &w,
		Right:  nil,
	}
	z := x.delete(30)
	equals(t, z, &x)
	equals(t, 50, x.Key)
	equals(t, 35, x.Left.Key)
	equals(t, 10, x.Left.Left.Key)

}

func TestNodeMax(t *testing.T) {
	var u *RBNode
	assert(t, u.max() == nil, "max returned non-nil for nil entry")

	w := RBNode{
		Key:    10,
		Value:  "yvonne",
		Color:  red,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}
	assert(t, w.max() == &w, "max returned wrong value for single node")
	w = RBNode{
		Key:    10,
		Value:  "ezekiel",
		Color:  black,
		Weight: 1,
		Left:   nil,
		Right:  nil,
	}

	x := RBNode{
		Key:    50,
		Value:  "steve",
		Color:  black,
		Weight: 1,
		Left:   &w,
		Right:  nil,
	}

	y := RBNode{
		Key:    30,
		Value:  "hector",
		Color:  black,
		Weight: 3,
		Left:   nil,
		Right:  nil,
	}

	w.Right = &y

	z := x.max()
	equals(t, z, &x)
	equals(t, 50, z.Key)

	v := RBNode{
		Key:    75,
		Value:  "leiland",
		Color:  black,
		Weight: 3,
		Left:   nil,
		Right:  nil,
	}
	x.Right = &v
	z = x.max()
	equals(t, &v, z)
	equals(t, 75, z.Key)

}

func TestTreeSize(t *testing.T) {
	var r *RBTree
	equals(t, 0, r.size())

	w := newNode(1, "scott", true)
	r = &RBTree{Root: w}
	equals(t, 1, r.size())
	x := newNode(10, "amy", false)
	x.Left = w
	x.Weight = 2
	r.Root = x
	equals(t, 2, r.size())
}

func TestTreeGet(t *testing.T) {
	var r *RBTree
	equals(t, "", r.get(10))
	r = &RBTree{}

	r.put(10, "harold")
	equals(t, "harold", r.get(10))

	r.put(20, "charles")
	equals(t, "charles", r.get(20))
}

func TestTreePut(t *testing.T) {
	var r *RBTree

	r.put(10, "mark")
	equals(t, "", r.get(10))

	r = &RBTree{}
	r.put(10, "nicholas")
	equals(t, "nicholas", r.get(10))
}

func TestTreeMin(t *testing.T) {
	var r *RBTree
	equals(t, "", r.min())

	r = &RBTree{}
	r.put(10, "thomas")
	equals(t, "thomas", r.min())
	r.put(100, "gary")
	equals(t, "thomas", r.min())

}

func TestTreeFloor(t *testing.T) {
	var r *RBTree
	equals(t, -1, r.floor(1))

	r = &RBTree{}
	r.put(10, "larry")
	equals(t, 10, r.floor(50))
	equals(t, -1, r.floor(1))
}

func TestTreeSelection(t *testing.T) {
	var r *RBTree
	equals(t, -1, r.selection(1))

	r = &RBTree{}
	r.put(10, "terry")
	equals(t, 10, r.selection(0))
	equals(t, -1, r.selection(10))
}

func TestTreeRank(t *testing.T) {
	var r *RBTree
	equals(t, -1, r.rank(1))

	r = &RBTree{}
	r.put(10, "maury")
	equals(t, 1, r.rank(11))
	equals(t, 0, r.rank(5))
}

func TestTreeDeleteMin(t *testing.T) {
	var r *RBTree
	r.deleteMin()

	r = &RBTree{}
	r.put(10, "barry")
	r.put(1, "gerry")
	r.put(25, "harry")
	r.put(100, "kerry")
	r.put(2, "perry")
	equals(t, "gerry", r.min())
	r.deleteMin()
	equals(t, "perry", r.min())

}

func TestTreeDelete(t *testing.T) {
	var r *RBTree
	r.delete(10)

	r = &RBTree{}
	r.put(10, "aiden")
	r.put(1, "tyler")
	r.put(25, "dakota")
	r.put(50, "abcde")
	r.delete(10)
	r.delete(50)
	equals(t, "", r.get(10))
	equals(t, "tyler", r.get(1))
	equals(t, "dakota", r.get(25))

}

func TestTreeMax(t *testing.T) {
	var r *RBTree
	equals(t, "", r.max())

	r = &RBTree{}

	r.put(10, "dakota")
	equals(t, "dakota", r.max())
	r.put(100, "bruce")
	equals(t, "bruce", r.max())
}

func TestTreeSuccessor(t *testing.T) {
	var r *RBTree
	equals(t, "", r.successor(10))

	r = &RBTree{}
	r.put(10, "sasha")
	r.put(20, "michelle")
	r.put(1, "barack")
	equals(t, "sasha", r.successor(10))
	equals(t, "sasha", r.successor(5))
	equals(t, "michelle", r.successor(15))
	equals(t, "barack", r.successor(25))
}

func TestTreePredecessor(t *testing.T) {
	var r *RBTree
	equals(t, "", r.predecessor(1))

	r = &RBTree{}
	r.put(10, "peter")
	r.put(20, "dimitris")
	r.put(1, "heiner")
	equals(t, "peter", r.predecessor(10))
	equals(t, "peter", r.predecessor(15))
	equals(t, "dimitris", r.predecessor(25))
	equals(t, "heiner", r.predecessor(5))
}
