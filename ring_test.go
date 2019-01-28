// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ring_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/thetannerryan/ring"
)

const (
	tests  = 1000000 // number of elements to test with (default: 1 million)
	fpRate = 0.001   // acceptable false positive rate (default: 0.1%)
)

var (
	// main testing
	r, _ = ring.Init(tests, fpRate)
	// benchmark
	rBench, _ = ring.Init(tests, fpRate)
	// error count
	errorCount = 0
)

// TestMain performs unit tests and benchmarks.
func TestMain(m *testing.M) {
	// run tests
	ret := m.Run()

	// print stats
	fmt.Printf(">> Number of elements:  %d\n", tests)
	fmt.Printf(">> Target false positive rate:  %f\n", fpRate)
	fmt.Printf(">> Number of false positives:  %d\n", errorCount)
	fmt.Printf(">> Actual false positive rate:  %f\n", float64(errorCount)/tests)

	// benchmarks
	fmt.Printf(">> Benchmark Add():  %s\n", testing.Benchmark(BenchmarkAdd))
	fmt.Printf(">> Benchmark Test():  %s\n", testing.Benchmark(BenchmarkTest))

	// actual failure if actual exceeds desired false positive rate
	if ret != 0 {
		os.Exit(ret)
	} else if float64(errorCount)/tests > fpRate {
		fmt.Printf("False positive threshold exceeded !!\n")
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

// BenchmarkAdd tests adding elements to a Ring.
func BenchmarkAdd(b *testing.B) {
	buff := make([]byte, 4)
	for i := 0; i < b.N; i++ {
		intToByte(buff, i)
		rBench.Add(buff)
	}
}

// BenchmarkTest tests elements in a Ring.
func BenchmarkTest(b *testing.B) {
	buff := make([]byte, 4)
	for i := 0; i < b.N; i++ {
		intToByte(buff, i)
		rBench.Test(buff)
	}
}

// TestBadParameters ensures that errornous parameters return an error.
func TestBadParameters(t *testing.T) {
	_, err := ring.Init(100, 1)
	if err == nil {
		t.Fatal("falsePositive >= 1 not captured")
	}
	_, err = ring.Init(100, 1.1)
	if err == nil {
		t.Fatal("falsePositive >= 1 not captured")
	}
	_, err = ring.Init(100, 0)
	if err == nil {
		t.Fatal("falsePositive <= 0 not captured")
	}
	_, err = ring.Init(100, -0.1)
	if err == nil {
		t.Fatal("falsePositive <= 0 not captured")
	}
	_, err = ring.Init(0, 0.1)
	if err == nil {
		t.Fatal("element <= 0 not captured")
	}
	_, err = ring.Init(-1, 0.1)
	if err == nil {
		t.Fatal("element <= 0 not captured")
	}
}

// TestReset ensures the Ring is cleared on Reset().
func TestReset(t *testing.T) {
	buff := make([]byte, 4)

	for i := 0; i < tests; i++ {
		intToByte(buff, i)
		r.Add(buff)
	}

	// ensure all data was removed
	r.Reset()
	for i := 0; i < tests; i++ {
		intToByte(buff, i)
		if r.Test(buff) {
			fmt.Printf("Data not removed !!\n")
			os.Exit(1)
		}
	}
}

// TestData performs unit tests on the Ring.
func TestData(t *testing.T) {
	var token []byte
	// byte range of random data
	min, max := 8, 4096
	for i := 0; i < tests; i++ {
		// generate random data
		size := rand.Intn(max-min) + min
		token = make([]byte, size)
		rand.Read(token)

		// test before adding
		if r.Test(token) {
			errorCount++
		}
		r.Add(token)
		// test after adding
		if !r.Test(token) {
			errorCount++
		}
	}
}

// intToByte converts an int (32-bit max) to byte array.
func intToByte(b []byte, v int) {
	_ = b[3] // memory safety
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}
