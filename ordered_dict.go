package main

import (
	"iter"
	"sync"
)

type node[K comparable, V any] struct {
	prev *node[K, V]
	next *node[K, V]
	key  K
	val  V
}

type OrderedDict[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]*node[K, V]
	head *node[K, V]
	tail *node[K, V]
	len  int
}

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

func (o *OrderedDict[K, V]) Set(key K, val V) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if existing, ok := o.data[key]; ok {
		existing.val = val
		return
	}

	n := &node[K, V]{key: key, val: val}

	lastTail := o.tail.prev
	o.tail.prev = n
	lastTail.next = n
	n.next = o.tail
	n.prev = lastTail

	o.data[key] = n
	o.len++
}

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

func (o *OrderedDict[K, V]) Delete(key K) (V, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	node, ok := o.data[key]
	if !ok {
		var zero V
		return zero, false // key doesn't exist
	}
	node.prev.next = node.next
	node.next.prev = node.prev
	delete(o.data, key)
	o.len--
	return node.val, true
}

func (o *OrderedDict[K, V]) Remove(key K) bool {
	_, ok := o.Delete(key)
	return ok
}

func (o *OrderedDict[K, V]) Len() int {
	return o.len
}

func (o *OrderedDict[K, V]) Has(key K) bool {
	_, ok := o.Get(key)
	return ok
}

func (o *OrderedDict[K, V]) Keys() []K {
	o.mu.RLock()
	defer o.mu.RUnlock()
	var k []K
	for curr := o.head.next; curr != o.tail; curr = curr.next {
		k = append(k, curr.key)
	}
	return k
}

func (o *OrderedDict[K, V]) Values() []V {
	o.mu.RLock()
	defer o.mu.RUnlock()
	var v []V
	for curr := o.head.next; curr != o.tail; curr = curr.next {
		v = append(v, curr.val)
	}
	return v
}

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
