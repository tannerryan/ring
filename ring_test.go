// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ring_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

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
	// false positive count
	positiveCount = 0
	// false negative count
	negativeCount = 0
)

// TestMain performs unit tests and benchmarks.
func TestMain(m *testing.M) {
	// run tests
	rand.Seed(time.Now().UTC().UnixNano())
	ret := m.Run()

	// print stats
	fmt.Printf(">> Number of elements:  %d\n", tests)
	fmt.Printf(">> Target false positive rate:  %f\n", fpRate)
	fmt.Printf(">> Number of false positives:  %d\n", positiveCount)
	fmt.Printf(">> Actual false positive rate:  %f\n", float64(positiveCount)/tests)
	fmt.Printf(">> Number of false negatives:  %d\n", negativeCount)
	fmt.Printf(">> Actual false negative rate:  %f\n", float64(negativeCount)/tests)

	// benchmarks
	fmt.Printf(">> Benchmark Add():  %s\n", testing.Benchmark(BenchmarkAdd))
	fmt.Printf(">> Benchmark Test():  %s\n", testing.Benchmark(BenchmarkTest))

	// actual failure if actual exceeds desired false positive rate
	if ret != 0 {
		os.Exit(ret)
	} else if float64(positiveCount)/tests > fpRate {
		fmt.Printf("False positive threshold exceeded !!\n")
		os.Exit(1)
	} else if negativeCount > 0 {
		fmt.Printf("False negative threshold exceeded !!\n")
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

	// InitByParameters tests

	_, err = ring.InitByParameters(1, 0)
	if err != nil {
		t.Fatal("Hash function cannot be <=0")
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
	min, max := 8, 8192
	for i := 0; i < tests; i++ {
		// generate random data
		size := rand.Intn(max-min) + min
		token = make([]byte, size)
		rand.Read(token)

		// test before adding
		if r.Test(token) {
			positiveCount++
		}
		r.Add(token)
		// test after adding
		if !r.Test(token) {
			negativeCount++
		}
	}
}

// TestMerge ensures that a Merge produces the right Ring.
func TestMerge(t *testing.T) {
	var token []byte
	// byte range of random data
	min, max := 8, 8192
	// test a range of sizes
	for i := uint(0); i < 20; i++ {
		innerCount := 1 << i
		elems := make([][]byte, innerCount)
		r, _ := ring.Init(tests, fpRate)
		r2, _ := ring.Init(tests, fpRate)
		for j := 0; j < innerCount; j++ {
			// generate random data
			size := rand.Intn(max-min) + min
			token = make([]byte, size)
			rand.Read(token)
			elems[j] = token
			if size&2 == 0 {
				r.Add(token)
			} else {
				r2.Add(token)
			}
		}
		if err := r.Merge(r2); err != nil {
			t.Errorf("Error calling Merge: %v", err)
			break
		}
		notFound := 0
		for j := 0; j < innerCount; j++ {
			if !r.Test(elems[j]) {
				notFound++
			}
		}
		if notFound > 0 {
			t.Errorf("Unexpected number of tokens not found: %v", notFound)
			break
		}
	}

	r, _ := ring.Init(tests, fpRate)
	// different params should fail to merge
	r2, _ := ring.Init(tests, 0.1)
	if r.Merge(r2) == nil {
		t.Errorf("Expected error calling Merge with different size")
	}
	r2, _ = ring.Init(100, fpRate)
	if r.Merge(r2) == nil {
		t.Errorf("Expected error calling Merge with different fp")
	}
}

// TestMarshal ensures that the Marshal and Unmarshal methods produce
// duplicate Ring's.
func TestMarshal(t *testing.T) {
	// Travis CI has strict memory limits that we hit if too high
	size := tests / 100
	r, _ := ring.Init(size, fpRate)
	elems := make([][]byte, size)
	var token []byte
	// byte range of random data
	min, max := 8, 8192
	// test a range of sizes
	for i := uint(0); i < uint(size); i++ {
		// generate random data
		size := rand.Intn(max-min) + min
		token = make([]byte, size)
		rand.Read(token)
		elems[i] = token
		r.Add(token)
	}

	out, err := r.MarshalBinary()
	if err != nil {
		t.Errorf("Unexpected error from MarshalBinary: %v", err)
		return
	}

	r2 := new(ring.Ring)
	r2.UnmarshalBinary(out)

	notFound := 0
	for _, el := range elems {
		if !r.Test(el) {
			notFound++
		}
	}
	if notFound > 0 {
		t.Errorf("Unexpected number of tokens not found: %v", notFound)
	}

	// unexpected length should error
	if r2.UnmarshalBinary(nil) == nil {
		t.Errorf("Expected error calling UnmarshalBinary with nil")
	}
	// unexpected version should error
	out[0] = 0
	if r2.UnmarshalBinary(out) == nil {
		t.Errorf("Expected error calling UnmarshalBinary with wrong version")
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
