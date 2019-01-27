// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ring_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/thetannerryan/ring"
)

const (
	tests  = 1000000 // number of elements to test with (default: 1 million)
	fpRate = 0.001   // acceptable false positive rate (default: 0.1%)
)

var (
	// main testing (no circular buffer)
	r1, _ = ring.Init(tests, fpRate, 0)
	// main testing (circular buffer)
	r2, _ = ring.Init(tests, fpRate, tests)
	// benchmark (no circular buffer)
	rBench1, _ = ring.Init(tests, fpRate, 0)
	// benchmark (circular buffer)
	rBench2, _ = ring.Init(tests, fpRate, tests)
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
	fmt.Printf(">> Actual false positive rate:  %f\n", float64(errorCount)/float64(tests))

	// benchmarks
	fmt.Printf(">> Benchmark Add() (no buffer):  %s\n", testing.Benchmark(BenchmarkNoBufferAdd))
	fmt.Printf(">> Benchmark Test() (no buffer):  %s\n", testing.Benchmark(BenchmarkNoBufferTest))
	fmt.Printf(">> Benchmark Add() (buffered):  %s\n", testing.Benchmark(BenchmarkBufferAdd))
	fmt.Printf(">> Benchmark Test() (buffered):  %s\n", testing.Benchmark(BenchmarkBufferTest))

	// actual failure if actual exceeds desired false positive rate
	if ret != 0 {
		os.Exit(ret)
	} else if float64(errorCount)/float64(tests) > fpRate {
		fmt.Printf("False positive threshold exceeded !!\n")
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

// BenchmarkNoBufferAdd tests adding elements to a Ring with no buffer.
func BenchmarkNoBufferAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rBench1.Add([]byte(strconv.Itoa(i)))
	}
}

// BenchmarkNoBufferTest tests elements in a Ring with no buffer.
func BenchmarkNoBufferTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rBench1.Test([]byte(strconv.Itoa(i)))
	}
}

// BenchmarkBufferAdd tests adding elements to a Ring with a buffer.
func BenchmarkBufferAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rBench2.Add([]byte(strconv.Itoa(i)))
	}
}

// BenchmarkBufferTest tests elements in a Ring with a buffer.
func BenchmarkBufferTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rBench2.Test([]byte(strconv.Itoa(i)))
	}
}

// TestBadParameters ensures that errornous parameters return an error.
func TestBadParameters(t *testing.T) {
	_, err := ring.Init(100, 1, 0)
	if err == nil {
		t.Error("invalid parameters not captured")
	}
	_, err = ring.Init(0, 0.05, 0)
	if err == nil {
		t.Error("invalid parameters not captured")
	}
	_, err = ring.Init(0, 0.1, 0)
	if err == nil {
		t.Error("invalid parameters not captured")
	}
}

// TestReset ensures the filter and buffer are properly cleared on Reset().
func TestReset(t *testing.T) {
	r1.Reset()
	r2.Reset()

	for i := 0; i < tests; i++ {
		r1.Add([]byte(strconv.Itoa(i)))
		r2.Add([]byte(strconv.Itoa(i)))
	}

	r1.Reset()
	r2.Reset()

	// ensure all data was removed
	for i := 0; i < tests; i++ {
		if r1.Test([]byte(strconv.Itoa(i))) {
			fmt.Printf("Data not removed !!\n")
			os.Exit(1)
		}
		if r2.Test([]byte(strconv.Itoa(i))) {
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
		if r1.Test(token) {
			errorCount++
		}
		r1.Add(token)
		// test after adding
		if !r1.Test(token) {
			errorCount++
		}
	}
}

// TestBuffer performs unit tests on the Ring with a buffer.
func TestBuffer(t *testing.T) {
	var token []byte
	bufferErrorCount := 0
	// byte range of random data
	min, max := 8, 4096
	for i := 0; i < tests; i++ {
		// generate random data
		size := rand.Intn(max-min) + min
		token = make([]byte, size)
		rand.Read(token)

		// test before adding
		if r2.Test(token) {
			bufferErrorCount++
		}
		r2.Add(token)
		// test after adding
		if !r2.Test(token) {
			bufferErrorCount++
		}
	}

	if bufferErrorCount > 0 {
		fmt.Printf("Buffer lost data !!\n")
		os.Exit(1)
	}
}
