// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/thetannerryan/ring"
)

func main() {
	// Support up to 100 elements with less than 1% false positives, no circular
	// buffer
	r, err := ring.Init(100, 0.01, 0)
	if err != nil {
		// error will only occur if parameters are set incorrectly
		panic(err)
	}

	data := []byte("hello")

	// check if data is in ring
	fmt.Printf("%s in ring :: %t\n", data, r.Test(data))

	// add data to ring
	r.Add(data)
	fmt.Printf("%s in ring :: %t\n", data, r.Test(data))

	// reset ring
	r.Reset()
	fmt.Printf("%s in ring :: %t\n", data, r.Test(data))
}
