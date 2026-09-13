// Copyright (c) 2019 Tanner Ryan. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/tannerryan/ring"
)

// main demonstrates adding, testing, and resetting a filter.
func main() {
	// Size the filter for 100 elements and a 1% false-positive rate.
	r, err := ring.Init(100, 0.01)
	if err != nil {
		panic(err)
	}

	data := []byte("hello")

	fmt.Printf("%q may be present: %t\n", data, r.Test(data))

	r.Add(data)
	fmt.Printf("%q may be present: %t\n", data, r.Test(data))

	r.Reset()
	fmt.Printf("%q may be present: %t\n", data, r.Test(data))
}
