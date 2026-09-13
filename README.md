# ring

[![Go
Reference](https://pkg.go.dev/badge/github.com/tannerryan/ring.svg)](https://pkg.go.dev/github.com/tannerryan/ring)
[![License](https://img.shields.io/github/license/tannerryan/ring.svg)](LICENSE)

A small, thread-safe Bloom filter for Go. A Bloom filter can rule out membership
with certainty. Positive results are probabilistic.

Go 1.27 or newer is required.

## Installation

```sh
go get github.com/tannerryan/ring@latest
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/tannerryan/ring"
)

func main() {
	filter, err := ring.Init(100_000, 0.001)
	if err != nil {
		panic(err)
	}

	filter.Add([]byte("hello"))
	fmt.Println(filter.Test([]byte("hello"))) // true
	fmt.Println(filter.Test([]byte("world"))) // probably false
}
```

`Init` takes the expected number of elements and the desired false-positive
rate. Adding more than the expected number increases the rate. After
initialization, all methods are safe for concurrent use. Use `InitByParameters`
when the exact bit count and number of hash rounds are already known. Both
values must be positive, and the hash rounds cannot exceed the bit count.
Entries cannot be removed individually. `Reset` clears the entire filter.

## Persistence

```go
encoded, err := filter.MarshalBinary()
if err != nil {
	panic(err)
}

var restored ring.Ring
err = restored.UnmarshalBinary(encoded)
if err != nil {
	panic(err)
}

fmt.Println(restored.Test([]byte("hello"))) // true
```

`MarshalBinary` and `UnmarshalBinary` save and restore a filter. `Merge`
combines filters with the same sizing and binary version. Version 1 filters
remain readable and retain their original hash behavior. New encodings use
version 2 and cannot be read by older releases.

The implementation uses MurmurHash3, a fast non-cryptographic hash. Treat the
filter as a probabilistic data structure, not as a security boundary.

## Development

Install the development tools with `make deps`, then run `make check`.

## License

This project is available under the [BSD 2-Clause License](LICENSE).
