package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type node struct {
	key   int
	value int
	left  *node
	right *node
}

type OrderedMap struct {
	root  *node
	count int
}

func NewOrderedMap() OrderedMap {
	return OrderedMap{}
}

func (m *OrderedMap) Insert(key, value int) {
	m.root = insert(m.root, key, value, &m.count)
}

func insert(n *node, key, value int, count *int) *node {
	if n == nil {
		*count++

		return &node{key: key, value: value}
	}

	switch {
	case key < n.key:
		n.left = insert(n.left, key, value, count)
	case key > n.key:
		n.right = insert(n.right, key, value, count)
	default:
		n.value = value
	}

	return n
}

func (m *OrderedMap) Erase(key int) {
	var deleted bool

	m.root, deleted = erase(m.root, key)

	if deleted {
		m.count--
	}
}

func erase(n *node, key int) (*node, bool) {
	if n == nil {
		return nil, false
	}

	var deleted bool

	switch {
	case key < n.key:
		n.left, deleted = erase(n.left, key)
	case key > n.key:
		n.right, deleted = erase(n.right, key)
	default:
		deleted = true

		if n.left == nil {
			return n.right, true
		}

		if n.right == nil {
			return n.left, true
		}

		succ := n.right

		for succ.left != nil {
			succ = succ.left
		}

		n.key, n.value = succ.key, succ.value

		n.right, _ = erase(n.right, succ.key)
	}

	return n, deleted
}

func (m *OrderedMap) Contains(key int) bool {
	cur := m.root

	for cur != nil {
		switch {
		case key < cur.key:
			cur = cur.left
		case key > cur.key:
			cur = cur.right
		default:
			return true
		}
	}

	return false
}

func (m *OrderedMap) Size() int {
	return m.count
}

func (m *OrderedMap) ForEach(action func(key int, value int)) {
	inorder(m.root, action)
}

func inorder(n *node, action func(key int, value int)) {
	if n == nil {
		return
	}

	inorder(n.left, action)
	action(n.key, n.value)
	inorder(n.right, action)
}

func TestCircularQueue(t *testing.T) {
	data := NewOrderedMap()
	assert.Zero(t, data.Size())

	data.Insert(10, 10)
	data.Insert(5, 5)
	data.Insert(15, 15)
	data.Insert(2, 2)
	data.Insert(4, 4)
	data.Insert(12, 12)
	data.Insert(14, 14)

	assert.Equal(t, 7, data.Size())
	assert.True(t, data.Contains(4))
	assert.True(t, data.Contains(12))
	assert.False(t, data.Contains(3))
	assert.False(t, data.Contains(13))

	var keys []int
	expectedKeys := []int{2, 4, 5, 10, 12, 14, 15}
	data.ForEach(func(key, _ int) {
		keys = append(keys, key)
	})

	assert.True(t, reflect.DeepEqual(expectedKeys, keys))

	data.Erase(15)
	data.Erase(14)
	data.Erase(2)

	assert.Equal(t, 4, data.Size())
	assert.True(t, data.Contains(4))
	assert.True(t, data.Contains(12))
	assert.False(t, data.Contains(2))
	assert.False(t, data.Contains(14))

	keys = nil
	expectedKeys = []int{4, 5, 10, 12}
	data.ForEach(func(key, _ int) {
		keys = append(keys, key)
	})

	assert.True(t, reflect.DeepEqual(expectedKeys, keys))
}
