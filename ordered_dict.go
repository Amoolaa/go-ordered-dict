package ordered_dict

import (
	"iter"
	"sync"
)

type OrderedDict[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]*node[K, V]
	head *node[K, V]
	tail *node[K, V]
	len  int
}

type node[K comparable, V any] struct {
	prev *node[K, V]
	next *node[K, V]
	key  K
	val  V
}

// New creates a new OrderedDict.
func New[K comparable, V any]() *OrderedDict[K, V] {
	head := &node[K, V]{}
	tail := &node[K, V]{}
	head.next = tail
	tail.prev = head
	return &OrderedDict[K, V]{
		data: make(map[K]*node[K, V]),
		head: head,
		tail: tail,
		len:  0,
	}
}

// NewWithCapacity creates a new OrderedDict with pre-allocated capacity.
func NewWithCapacity[K comparable, V any](capacity int) *OrderedDict[K, V] {
	head := &node[K, V]{}
	tail := &node[K, V]{}
	head.next = tail
	tail.prev = head
	return &OrderedDict[K, V]{
		data: make(map[K]*node[K, V], capacity),
		head: head,
		tail: tail,
		len:  0,
	}
}

// Set adds or updates a key-value pair.
func (o *OrderedDict[K, V]) Set(key K, val V) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if existing, ok := o.data[key]; ok {
		existing.val = val
		return
	}

	n := &node[K, V]{key: key, val: val}
	o.linkToEnd(n)

	o.data[key] = n
	o.len++
}

// Get retrieves a value by key, returns false if key doesn't exist.
func (o *OrderedDict[K, V]) Get(key K) (V, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	node, ok := o.data[key]
	if !ok {
		var zero V
		return zero, false
	}
	return node.val, true
}

func (o *OrderedDict[K, V]) unlinkNode(n *node[K, V]) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

// Delete removes a key and returns its value, returns false if key doesn't exist.
func (o *OrderedDict[K, V]) Delete(key K) (V, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	node, ok := o.data[key]
	if !ok {
		var zero V
		return zero, false // key doesn't exist
	}
	o.unlinkNode(node)
	delete(o.data, key)
	o.len--
	return node.val, true
}

// Remove deletes a key, returns true if key existed.
func (o *OrderedDict[K, V]) Remove(key K) bool {
	_, ok := o.Delete(key)
	return ok
}

// Len returns the number of items in the dictionary.
func (o *OrderedDict[K, V]) Len() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.len
}

// Has checks if a key exists in the dictionary.
func (o *OrderedDict[K, V]) Has(key K) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	_, ok := o.data[key]
	return ok
}

// Keys returns all keys in insertion order.
func (o *OrderedDict[K, V]) Keys() []K {
	o.mu.RLock()
	defer o.mu.RUnlock()
	var k []K
	for curr := o.head.next; curr != o.tail; curr = curr.next {
		k = append(k, curr.key)
	}
	return k
}

// Values returns all values in insertion order.
func (o *OrderedDict[K, V]) Values() []V {
	o.mu.RLock()
	defer o.mu.RUnlock()
	var v []V
	for curr := o.head.next; curr != o.tail; curr = curr.next {
		v = append(v, curr.val)
	}
	return v
}

// All returns an iterator over key-value pairs in insertion order.
func (o *OrderedDict[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		o.mu.RLock()
		defer o.mu.RUnlock()
		for curr := o.head.next; curr != o.tail; curr = curr.next {
			if !yield(curr.key, curr.val) {
				return
			}
		}
	}
}

// Clear removes all items from the dictionary.
func (o *OrderedDict[K, V]) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.head = &node[K, V]{}
	o.tail = &node[K, V]{}
	o.head.next = o.tail
	o.tail.prev = o.head
	o.len = 0
	clear(o.data)
}

func (o *OrderedDict[K, V]) linkToEnd(n *node[K, V]) {
	prevTail := o.tail.prev
	n.prev = prevTail
	n.next = o.tail
	prevTail.next = n
	o.tail.prev = n
}

func (o *OrderedDict[K, V]) linkToStart(n *node[K, V]) {
	prevHead := o.head.next
	n.next = prevHead
	n.prev = o.head
	prevHead.prev = n
	o.head.next = n
}

func (o *OrderedDict[K, V]) linkAfter(n *node[K, V], after *node[K, V]) {
	afterNext := after.next
	n.next = afterNext
	n.prev = after
	after.next = n
	afterNext.prev = n
}

// MoveToEnd moves a key to the end of the order, returns false if key doesn't exist.
func (o *OrderedDict[K, V]) MoveToEnd(key K) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	node, ok := o.data[key]
	if !ok {
		return false
	}
	o.unlinkNode(node)
	o.linkToEnd(node)
	return true
}

// MoveToStart moves a key to the start of the order, returns false if key doesn't exist.
func (o *OrderedDict[K, V]) MoveToStart(key K) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	node, ok := o.data[key]
	if !ok {
		return false
	}
	o.unlinkNode(node)
	o.linkToStart(node)
	return true
}

// MoveAfter moves a key after another key, returns false if either key doesn't exist.
func (o *OrderedDict[K, V]) MoveAfter(key K, after K) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	afterNode, ok := o.data[after]
	if !ok {
		return false
	}
	node, ok := o.data[key]
	if !ok {
		return false
	}
	o.unlinkNode(node)
	o.linkAfter(node, afterNode)
	return true
}
