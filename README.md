# go-ordered-dict

[![Go Reference](https://pkg.go.dev/badge/github.com/amoolaa/go-ordered-dict.svg)](https://pkg.go.dev/github.com/amoolaa/go-ordered-dict)

A thread-safe ordered dictionary implementation in Go that maintains insertion order while providing O(1) lookups, inserts, and deletes.

## Installation

```bash
go get github.com/amoolaa/go-ordered-dict
```

## Usage

### Basic Operations

```go
package main

import (
    "fmt"
    "github.com/amoolaa/go-ordered-dict"
)

func main() {
    // Create a new ordered dictionary
    dict := ordereddict.New[string, int]()

    // Add items
    dict.Set("first", 1)
    dict.Set("second", 2)
    dict.Set("third", 3)

    // Get items
    if val, ok := dict.Get("second"); ok {
        fmt.Println(val) // Output: 2
    }

    // Check if key exists
    if dict.Has("first") {
        fmt.Println("Key exists")
    }

    // Get length
    fmt.Println(dict.Len()) // Output: 3

    // Delete items
    if val, ok := dict.Delete("second"); ok {
        fmt.Println("Deleted:", val)
    }

    // Pretty print the dictionary
    fmt.Println(dict) // Output: OrderedDict[first:1 third:3]
}
```

### Pretty Printing

The `String()` method implements the `Stringer` interface, allowing you to easily print the dictionary in a format similar to Go's built-in maps:

```go
dict := ordereddict.New[string, int]()
dict.Set("apple", 1)
dict.Set("banana", 2)
dict.Set("cherry", 3)

fmt.Println(dict)  // Output: OrderedDict[apple:1 banana:2 cherry:3]
fmt.Printf("Dict: %v\n", dict)  // Also works with Printf
```

### Iterating Over Items

```go
// Get all keys in insertion order
keys := dict.Keys()
fmt.Println(keys) // ["first", "third"]

// Get all values in insertion order
values := dict.Values()
fmt.Println(values) // [1, 3]

// Using iterator (Go 1.23+)
for key, val := range dict.All() {
    fmt.Printf("%s: %d\n", key, val)
}
```

### Reordering Items

```go
dict := ordereddict.New[string, string]()
dict.Set("a", "first")
dict.Set("b", "second")
dict.Set("c", "third")

// Move to end
dict.MoveToEnd("a")
fmt.Println(dict.Keys()) // ["b", "c", "a"]

// Move to start
dict.MoveToStart("c")
fmt.Println(dict.Keys()) // ["c", "b", "a"]

// Move after another key
dict.MoveAfter("c", "a")
fmt.Println(dict.Keys()) // ["b", "a", "c"]
```

### Pre-allocating Capacity

```go
// Create dictionary with pre-allocated capacity
dict := ordereddict.NewWithCapacity[string, int](100)
```

## Features

- Thread-safe with read-write mutex
- Generic types support (Go 1.18+)
- O(1) insert, lookup, and delete operations
- Maintains insertion order
- Ability to reorder items
- Iterator support (Go 1.23+)
- Pretty printing via `String()` method (implements `fmt.Stringer`)
