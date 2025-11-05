package ordereddict

import (
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	od := New[string, int]()

	if od == nil {
		t.Fatal("New() returned nil")
	}
	if od.data == nil {
		t.Fatal("data map not initialized")
	}
	if od.head == nil {
		t.Fatal("head sentinel not initialized")
	}
	if od.tail == nil {
		t.Fatal("tail sentinel not initialized")
	}
	if od.len != 0 {
		t.Errorf("expected len=0, got %d", od.len)
	}
	if od.head.next != od.tail {
		t.Error("head.next should point to tail")
	}
	if od.tail.prev != od.head {
		t.Error("tail.prev should point to head")
	}
}

func TestNewWithCapacity(t *testing.T) {
	od := NewWithCapacity[string, int](100)

	if od == nil {
		t.Fatal("NewWithCapacity() returned nil")
	}
	if od.data == nil {
		t.Fatal("data map not initialized")
	}
	if od.len != 0 {
		t.Errorf("expected len=0, got %d", od.len)
	}
}

func TestSetSingleKey(t *testing.T) {
	od := New[string, int]()
	od.Set("key1", 100)

	if od.len != 1 {
		t.Errorf("expected len=1, got %d", od.len)
	}
	if node, ok := od.data["key1"]; !ok {
		t.Error("key1 not found in map")
	} else if node.val != 100 {
		t.Errorf("expected value=100, got %d", node.val)
	}
}

func TestSetUpdateExistingKey(t *testing.T) {
	od := New[string, int]()
	od.Set("key1", 100)
	od.Set("key1", 200)

	if od.len != 1 {
		t.Errorf("expected len=1 after update, got %d", od.len)
	}
	if node, ok := od.data["key1"]; !ok {
		t.Error("key1 not found in map")
	} else if node.val != 200 {
		t.Errorf("expected updated value=200, got %d", node.val)
	}
}

func TestSetMultipleKeys(t *testing.T) {
	od := New[string, int]()
	od.Set("first", 1)
	od.Set("second", 2)
	od.Set("third", 3)

	if od.len != 3 {
		t.Errorf("expected len=3, got %d", od.len)
	}

	expected := map[string]int{
		"first":  1,
		"second": 2,
		"third":  3,
	}

	for key, expectedVal := range expected {
		if node, ok := od.data[key]; !ok {
			t.Errorf("key %s not found in map", key)
		} else if node.val != expectedVal {
			t.Errorf("key %s: expected value=%d, got %d", key, expectedVal, node.val)
		}
	}
}

func TestSetInsertionOrder(t *testing.T) {
	od := New[string, int]()
	od.Set("first", 1)
	od.Set("second", 2)
	od.Set("third", 3)

	keys := []string{}
	for node := od.head.next; node != od.tail; node = node.next {
		keys = append(keys, node.key)
	}

	expected := []string{"first", "second", "third"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}

	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}

func TestSetUpdatePreservesOrder(t *testing.T) {
	od := New[string, int]()
	od.Set("first", 1)
	od.Set("second", 2)
	od.Set("third", 3)
	od.Set("second", 200)

	if od.len != 3 {
		t.Errorf("expected len=3, got %d", od.len)
	}

	keys := []string{}
	for node := od.head.next; node != od.tail; node = node.next {
		keys = append(keys, node.key)
	}

	expected := []string{"first", "second", "third"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}

	if node := od.data["second"]; node.val != 200 {
		t.Errorf("expected updated value=200, got %d", node.val)
	}
}

func TestSetConcurrent(t *testing.T) {
	od := New[int, int]()
	var wg sync.WaitGroup
	numGoroutines := 50
	numOperations := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := id*numOperations + j
				od.Set(key, key*2)
			}
		}(i)
	}

	wg.Wait()

	expectedLen := numGoroutines * numOperations
	if od.Len() != expectedLen {
		t.Errorf("expected len=%d, got %d", expectedLen, od.Len())
	}
}

func TestGet(t *testing.T) {
	od := New[string, int]()
	od.Set("key1", 100)

	val, ok := od.Get("key1")
	if !ok {
		t.Error("expected key1 to exist")
	}
	if val != 100 {
		t.Errorf("expected value=100, got %d", val)
	}

	val, ok = od.Get("nonexistent")
	if ok {
		t.Error("expected nonexistent key to not exist")
	}
	if val != 0 {
		t.Errorf("expected zero value, got %d", val)
	}
}

func TestGetConcurrent(t *testing.T) {
	od := New[int, int]()
	for i := 0; i < 100; i++ {
		od.Set(i, i*2)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				val, ok := od.Get(j)
				if !ok {
					t.Errorf("key %d should exist", j)
				}
				if val != j*2 {
					t.Errorf("key %d: expected %d, got %d", j, j*2, val)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestHas(t *testing.T) {
	od := New[string, int]()
	od.Set("exists", 1)

	if !od.Has("exists") {
		t.Error("expected key to exist")
	}
	if od.Has("nonexistent") {
		t.Error("expected key to not exist")
	}
}

func TestLen(t *testing.T) {
	od := New[string, int]()

	if od.Len() != 0 {
		t.Errorf("expected len=0, got %d", od.Len())
	}

	od.Set("a", 1)
	if od.Len() != 1 {
		t.Errorf("expected len=1, got %d", od.Len())
	}

	od.Set("b", 2)
	od.Set("c", 3)
	if od.Len() != 3 {
		t.Errorf("expected len=3, got %d", od.Len())
	}

	od.Set("a", 10)
	if od.Len() != 3 {
		t.Errorf("expected len=3 after update, got %d", od.Len())
	}
}

func TestDelete(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	val, ok := od.Delete("b")
	if !ok {
		t.Error("expected delete to succeed")
	}
	if val != 2 {
		t.Errorf("expected deleted value=2, got %d", val)
	}
	if od.Len() != 2 {
		t.Errorf("expected len=2, got %d", od.Len())
	}

	if od.Has("b") {
		t.Error("key b should not exist after delete")
	}

	keys := []string{}
	for node := od.head.next; node != od.tail; node = node.next {
		keys = append(keys, node.key)
	}
	expected := []string{"a", "c"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}

func TestDeleteNonexistent(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)

	val, ok := od.Delete("nonexistent")
	if ok {
		t.Error("expected delete to return false for nonexistent key")
	}
	if val != 0 {
		t.Errorf("expected zero value, got %d", val)
	}
	if od.Len() != 1 {
		t.Errorf("expected len=1, got %d", od.Len())
	}
}

func TestDeleteAll(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)

	od.Delete("a")
	od.Delete("b")

	if od.Len() != 0 {
		t.Errorf("expected len=0, got %d", od.Len())
	}
	if od.head.next != od.tail {
		t.Error("head should point to tail in empty dict")
	}
	if od.tail.prev != od.head {
		t.Error("tail should point to head in empty dict")
	}
}

func TestKeys(t *testing.T) {
	od := New[string, int]()
	od.Set("first", 1)
	od.Set("second", 2)
	od.Set("third", 3)

	keys := od.Keys()
	expected := []string{"first", "second", "third"}

	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}

	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}

func TestKeysEmpty(t *testing.T) {
	od := New[string, int]()
	keys := od.Keys()

	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestValues(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 10)
	od.Set("b", 20)
	od.Set("c", 30)

	values := od.Values()
	expected := []int{10, 20, 30}

	if len(values) != len(expected) {
		t.Fatalf("expected %d values, got %d", len(expected), len(values))
	}

	for i, val := range values {
		if val != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], val)
		}
	}
}

func TestAll(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	expected := map[string]int{"a": 1, "b": 2, "c": 3}
	expectedOrder := []string{"a", "b", "c"}
	count := 0

	for key, value := range od.All() {
		if expected[key] != value {
			t.Errorf("key %s: expected value=%d, got %d", key, expected[key], value)
		}
		if key != expectedOrder[count] {
			t.Errorf("position %d: expected key=%s, got %s", count, expectedOrder[count], key)
		}
		count++
	}

	if count != 3 {
		t.Errorf("expected 3 iterations, got %d", count)
	}
}

func TestAllEmpty(t *testing.T) {
	od := New[string, int]()
	count := 0

	for range od.All() {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 iterations, got %d", count)
	}
}

func TestAllBreak(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	count := 0
	for key := range od.All() {
		count++
		if key == "b" {
			break
		}
	}

	if count != 2 {
		t.Errorf("expected 2 iterations before break, got %d", count)
	}
}

func TestDifferentTypes(t *testing.T) {
	t.Run("int keys", func(t *testing.T) {
		od := New[int, string]()
		od.Set(1, "one")
		od.Set(2, "two")

		val, ok := od.Get(1)
		if !ok || val != "one" {
			t.Error("int key failed")
		}
	})

	t.Run("struct keys", func(t *testing.T) {
		type Key struct{ id int }
		od := New[Key, bool]()
		k1 := Key{1}
		k2 := Key{2}
		od.Set(k1, true)
		od.Set(k2, false)

		val, ok := od.Get(k1)
		if !ok || val != true {
			t.Error("struct key failed")
		}
	})

	t.Run("pointer values", func(t *testing.T) {
		od := New[string, *int]()
		val1 := 100
		val2 := 200
		od.Set("a", &val1)
		od.Set("b", &val2)

		retrieved, ok := od.Get("a")
		if !ok || *retrieved != 100 {
			t.Error("pointer value failed")
		}
	})
}

func TestConcurrentMixed(t *testing.T) {
	od := New[int, int]()
	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range 100 {
				key := id*100 + j
				od.Set(key, key)
			}
		}(i)
	}

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range 100 {
				od.Get(j)
				od.Has(j)
				od.Len()
			}
		}()
	}

	wg.Wait()
}

func TestZeroValues(t *testing.T) {
	od := New[string, int]()
	od.Set("zero", 0)
	od.Set("", 42)

	val, ok := od.Get("zero")
	if !ok || val != 0 {
		t.Error("zero value storage failed")
	}

	val, ok = od.Get("")
	if !ok || val != 42 {
		t.Error("empty key failed")
	}
}

func TestClear(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	if od.Len() != 3 {
		t.Errorf("expected len=3, got %d", od.Len())
	}

	od.Clear()

	if od.Len() != 0 {
		t.Errorf("expected len=0 after clear, got %d", od.Len())
	}

	if od.Has("a") || od.Has("b") || od.Has("c") {
		t.Error("keys should not exist after clear")
	}

	keys := od.Keys()
	if len(keys) != 0 {
		t.Errorf("expected 0 keys after clear, got %d", len(keys))
	}

	// Verify can add items after clear
	od.Set("new", 100)
	if od.Len() != 1 {
		t.Errorf("expected len=1 after adding to cleared dict, got %d", od.Len())
	}
	val, ok := od.Get("new")
	if !ok || val != 100 {
		t.Error("failed to add item after clear")
	}
}

func TestClearEmpty(t *testing.T) {
	od := New[string, int]()
	od.Clear()

	if od.Len() != 0 {
		t.Errorf("expected len=0, got %d", od.Len())
	}
}

func TestRemove(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	ok := od.Remove("b")
	if !ok {
		t.Error("expected remove to succeed")
	}
	if od.Len() != 2 {
		t.Errorf("expected len=2, got %d", od.Len())
	}

	if od.Has("b") {
		t.Error("key b should not exist after remove")
	}

	keys := od.Keys()
	expected := []string{"a", "c"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}

func TestRemoveNonexistent(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)

	ok := od.Remove("nonexistent")
	if ok {
		t.Error("expected remove to return false for nonexistent key")
	}
	if od.Len() != 1 {
		t.Errorf("expected len=1, got %d", od.Len())
	}
}

func TestMoveToEnd(t *testing.T) {
	od := New[string, int]()
	od.Set("first", 1)
	od.Set("second", 2)
	od.Set("third", 3)

	ok := od.MoveToEnd("first")
	if !ok {
		t.Error("expected MoveToEnd to succeed")
	}

	keys := od.Keys()
	expected := []string{"second", "third", "first"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}

	// Verify value is preserved
	val, ok := od.Get("first")
	if !ok || val != 1 {
		t.Error("value should be preserved after move")
	}

	// Verify length unchanged
	if od.Len() != 3 {
		t.Errorf("expected len=3, got %d", od.Len())
	}
}

func TestMoveToEndAlreadyAtEnd(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	ok := od.MoveToEnd("c")
	if !ok {
		t.Error("expected MoveToEnd to succeed even when already at end")
	}

	keys := od.Keys()
	expected := []string{"a", "b", "c"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}

func TestMoveToEndNonexistent(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)

	ok := od.MoveToEnd("nonexistent")
	if ok {
		t.Error("expected MoveToEnd to return false for nonexistent key")
	}

	if od.Len() != 1 {
		t.Errorf("expected len=1, got %d", od.Len())
	}
}

func TestMoveToEndSingleElement(t *testing.T) {
	od := New[string, int]()
	od.Set("only", 42)

	ok := od.MoveToEnd("only")
	if !ok {
		t.Error("expected MoveToEnd to succeed")
	}

	if od.Len() != 1 {
		t.Errorf("expected len=1, got %d", od.Len())
	}

	keys := od.Keys()
	if len(keys) != 1 || keys[0] != "only" {
		t.Error("single element should remain")
	}
}

func TestMoveToStart(t *testing.T) {
	od := New[string, int]()
	od.Set("first", 1)
	od.Set("second", 2)
	od.Set("third", 3)

	ok := od.MoveToStart("third")
	if !ok {
		t.Error("expected MoveToStart to succeed")
	}

	keys := od.Keys()
	expected := []string{"third", "first", "second"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}

	// Verify value is preserved
	val, ok := od.Get("third")
	if !ok || val != 3 {
		t.Error("value should be preserved after move")
	}

	// Verify length unchanged
	if od.Len() != 3 {
		t.Errorf("expected len=3, got %d", od.Len())
	}
}

func TestMoveToStartAlreadyAtStart(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	ok := od.MoveToStart("a")
	if !ok {
		t.Error("expected MoveToStart to succeed even when already at start")
	}

	keys := od.Keys()
	expected := []string{"a", "b", "c"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}

func TestMoveToStartNonexistent(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)

	ok := od.MoveToStart("nonexistent")
	if ok {
		t.Error("expected MoveToStart to return false for nonexistent key")
	}

	if od.Len() != 1 {
		t.Errorf("expected len=1, got %d", od.Len())
	}
}

func TestMoveToStartSingleElement(t *testing.T) {
	od := New[string, int]()
	od.Set("only", 42)

	ok := od.MoveToStart("only")
	if !ok {
		t.Error("expected MoveToStart to succeed")
	}

	if od.Len() != 1 {
		t.Errorf("expected len=1, got %d", od.Len())
	}

	keys := od.Keys()
	if len(keys) != 1 || keys[0] != "only" {
		t.Error("single element should remain")
	}
}

func TestMoveAfter(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)
	od.Set("d", 4)

	// Move "a" after "c": should become b, c, a, d
	ok := od.MoveAfter("a", "c")
	if !ok {
		t.Error("expected MoveAfter to succeed")
	}

	keys := od.Keys()
	expected := []string{"b", "c", "a", "d"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}

	// Verify value is preserved
	val, ok := od.Get("a")
	if !ok || val != 1 {
		t.Error("value should be preserved after move")
	}

	// Verify length unchanged
	if od.Len() != 4 {
		t.Errorf("expected len=4, got %d", od.Len())
	}
}

func TestMoveAfterToEnd(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	// Move "a" after "c" (last element): should become b, c, a
	ok := od.MoveAfter("a", "c")
	if !ok {
		t.Error("expected MoveAfter to succeed")
	}

	keys := od.Keys()
	expected := []string{"b", "c", "a"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}

func TestMoveAfterAdjacentElements(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	// Move "c" after "a": should become a, c, b
	ok := od.MoveAfter("c", "a")
	if !ok {
		t.Error("expected MoveAfter to succeed")
	}

	keys := od.Keys()
	expected := []string{"a", "c", "b"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}

func TestMoveAfterNonexistentKey(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)

	ok := od.MoveAfter("nonexistent", "a")
	if ok {
		t.Error("expected MoveAfter to return false for nonexistent key")
	}

	if od.Len() != 2 {
		t.Errorf("expected len=2, got %d", od.Len())
	}

	// Order should be unchanged
	keys := od.Keys()
	expected := []string{"a", "b"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}

func TestMoveAfterSameKey(t *testing.T) {
	od := New[string, int]()
	od.Set("a", 1)
	od.Set("b", 2)
	od.Set("c", 3)

	// Move "b" after "b" - edge case
	ok := od.MoveAfter("b", "b")
	if !ok {
		t.Error("expected MoveAfter to succeed")
	}

	if od.Len() != 3 {
		t.Errorf("expected len=3, got %d", od.Len())
	}

	// After moving "b" after itself (which was deleted first),
	// the order becomes: a, c, b (b moved to after where it used to be)
	keys := od.Keys()
	expected := []string{"a", "c", "b"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], key)
		}
	}
}
