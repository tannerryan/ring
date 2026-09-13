// Copyright (c) 2019 Tanner Ryan. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

package ring_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/tannerryan/ring"
)

// TestInit checks constructor validation and membership behavior.
func TestInit(t *testing.T) {
	for _, test := range []struct {
		name          string
		elements      int
		falsePositive float64
	}{
		{"zero elements", 0, 0.01},
		{"negative elements", -1, 0.01},
		{"zero rate", 100, 0},
		{"negative rate", 100, -0.01},
		{"rate of one", 100, 1},
		{"rate above one", 100, 1.01},
		{"not-a-number rate", 100, math.NaN()},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ring.Init(test.elements, test.falsePositive); err == nil {
				t.Fatal("Init returned no error")
			}
		})
	}

	filter := newFilter(t, 100, 0.01)
	if filter.Test([]byte("hello")) {
		t.Fatal("new filter contains data")
	}
	filter.Add([]byte("hello"))
	if !filter.Test([]byte("hello")) {
		t.Fatal("added data is missing")
	}
}

// TestInitByParameters checks explicit sizing and validation.
func TestInitByParameters(t *testing.T) {
	for _, parameters := range []struct {
		size   uint64
		hashes uint64
	}{{0, 1}, {8, 0}, {8, 9}} {
		if _, err := ring.InitByParameters(parameters.size, parameters.hashes); err == nil {
			t.Fatalf("InitByParameters(%d, %d) returned no error", parameters.size, parameters.hashes)
		}
	}

	filter, err := ring.InitByParameters(128, 3)
	if err != nil {
		t.Fatal(err)
	}
	filter.Add([]byte("hello"))
	if !filter.Test([]byte("hello")) {
		t.Fatal("added data is missing")
	}
	encoded, err := filter.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if encoded[0] != 2 {
		t.Fatalf("binary version = %d, want 2", encoded[0])
	}
	if size := binary.BigEndian.Uint64(encoded[1:9]); size != 128 {
		t.Fatalf("binary size = %d, want 128", size)
	}
	if hashes := binary.BigEndian.Uint64(encoded[9:17]); hashes != 3 {
		t.Fatalf("binary hash rounds = %d, want 3", hashes)
	}
}

// TestFalsePositiveRate checks the configured rate on a filled filter.
func TestFalsePositiveRate(t *testing.T) {
	const (
		elements = 10_000
		rate     = 0.01
	)
	filter := newFilter(t, elements, rate)
	for i := range elements {
		filter.Add(key(0, i))
	}
	for i := range elements {
		if !filter.Test(key(0, i)) {
			t.Fatalf("added key %d is missing", i)
		}
	}

	falsePositives := 0
	for i := range elements {
		if filter.Test(key(1, i)) {
			falsePositives++
		}
	}
	if actual := float64(falsePositives) / elements; actual > rate*1.5 {
		t.Fatalf("false-positive rate = %f, want at most %f", actual, rate*1.5)
	}
}

// TestReset checks that Reset removes all membership information.
func TestReset(t *testing.T) {
	filter := newFilter(t, 100, 0.01)
	filter.Add([]byte("hello"))
	filter.Reset()
	if filter.Test([]byte("hello")) {
		t.Fatal("reset filter contains data")
	}
}

// TestMerge checks compatible, incompatible, and nil filters.
func TestMerge(t *testing.T) {
	left := newFilter(t, 100, 0.01)
	right := newFilter(t, 100, 0.01)
	left.Add([]byte("left"))
	right.Add([]byte("right"))
	if err := left.Merge(right); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"left", "right"} {
		if !left.Test([]byte(value)) {
			t.Fatalf("merged filter is missing %q", value)
		}
	}

	incompatible := newFilter(t, 10, 0.1)
	if err := left.Merge(incompatible); err == nil {
		t.Fatal("Merge accepted incompatible filters")
	}
	if err := left.Merge(nil); err == nil {
		t.Fatal("Merge accepted a nil filter")
	}
	if err := new(ring.Ring).Merge(right); err == nil {
		t.Fatal("Merge accepted an uninitialized filter")
	}
	encoded, err := right.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	encoded[0] = 1
	var legacy ring.Ring
	if err := legacy.UnmarshalBinary(encoded); err != nil {
		t.Fatal(err)
	}
	if err := left.Merge(&legacy); err == nil {
		t.Fatal("Merge accepted different binary versions")
	}
}

// TestMergeConcurrency checks self-merge and reciprocal merge for deadlocks.
func TestMergeConcurrency(t *testing.T) {
	left := newFilter(t, 100, 0.01)
	right := newFilter(t, 100, 0.01)
	if err := mergeWithTimeout(left, left); err != nil {
		t.Fatalf("self merge: %v", err)
	}

	done := make(chan error, 2)
	go func() { done <- left.Merge(right) }()
	go func() { done <- right.Merge(left) }()
	for range 2 {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("reciprocal merge deadlocked")
		}
	}
}

// TestConcurrentUse exercises all stateful methods together.
func TestConcurrentUse(t *testing.T) {
	filter := newFilter(t, 1_000, 0.01)
	peer := newFilter(t, 1_000, 0.01)
	peer.Add([]byte("peer"))
	encoded, err := filter.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{}, 5)
	for worker := range 5 {
		go func() {
			for i := range 100 {
				data := key(byte(worker), i)
				filter.Add(data)
				filter.Test(data)
				if _, err := filter.MarshalBinary(); err != nil {
					t.Error(err)
				}
				if i%10 == 0 {
					if err := filter.Merge(peer); err != nil {
						t.Error(err)
					}
				}
				if i%25 == 0 {
					filter.Reset()
					if err := filter.UnmarshalBinary(encoded); err != nil {
						t.Error(err)
					}
				}
			}
			done <- struct{}{}
		}()
	}
	for range 5 {
		<-done
	}
}

// TestBinaryRoundTrip checks encoding, decoding, and input ownership.
func TestBinaryRoundTrip(t *testing.T) {
	filter := newFilter(t, 100, 0.01)
	data := []byte("a serialized value long enough to exercise the corrected hash tail")
	filter.Add(data)
	encoded, err := filter.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if encoded[0] != 2 {
		t.Fatalf("binary version = %d, want 2", encoded[0])
	}

	var decoded ring.Ring
	if err := decoded.UnmarshalBinary(encoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Test(data) {
		t.Fatal("decoded filter is missing data")
	}
	clear(encoded[17:])
	if !decoded.Test(data) {
		t.Fatal("decoded filter retained the input buffer")
	}
	reencoded, err := decoded.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(encoded, reencoded) {
		t.Fatal("mutating encoded input changed the decoded filter")
	}
}

// TestLegacyBinary checks version 1 hash and persistence compatibility.
func TestLegacyBinary(t *testing.T) {
	encoded := make([]byte, 19)
	encoded[0] = 1
	binary.BigEndian.PutUint64(encoded[1:9], 8)
	binary.BigEndian.PutUint64(encoded[9:17], 1)
	var filter ring.Ring
	if err := filter.UnmarshalBinary(encoded); err != nil {
		t.Fatal(err)
	}
	data := []byte("The quick brown fox jumps over the lazy dog.")
	filter.Add(data)
	if !filter.Test(data) {
		t.Fatal("version 1 filter is missing added data")
	}
	encoded, err := filter.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if encoded[0] != 1 || encoded[17] != 0x10 || encoded[18] != 0 {
		t.Fatalf("version 1 encoding changed: %x", encoded)
	}
}

// TestBinaryErrors checks malformed encodings and atomic decode failures.
func TestBinaryErrors(t *testing.T) {
	filter := newFilter(t, 100, 0.01)
	filter.Add([]byte("keep"))
	valid, err := filter.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	wrongVersion := append([]byte(nil), valid...)
	wrongVersion[0]++
	zeroSize := append([]byte(nil), valid...)
	binary.BigEndian.PutUint64(zeroSize[1:9], 0)
	zeroHash := append([]byte(nil), valid...)
	binary.BigEndian.PutUint64(zeroHash[9:17], 0)
	tooManyHashes := append([]byte(nil), valid...)
	binary.BigEndian.PutUint64(tooManyHashes[9:17], math.MaxUint64)
	hugeSize := append([]byte(nil), valid...)
	binary.BigEndian.PutUint64(hugeSize[1:9], math.MaxUint64)
	binary.BigEndian.PutUint64(hugeSize[9:17], 1)

	for _, test := range []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"wrong version", wrongVersion},
		{"zero size", zeroSize},
		{"zero hash", zeroHash},
		{"too many hashes", tooManyHashes},
		{"huge size", hugeSize},
		{"truncated", valid[:len(valid)-1]},
		{"trailing data", append(append([]byte(nil), valid...), 0)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := filter.UnmarshalBinary(test.data); err == nil {
				t.Fatal("UnmarshalBinary returned no error")
			}
			if !filter.Test([]byte("keep")) {
				t.Fatal("failed decode changed the filter")
			}
		})
	}

	var nilFilter *ring.Ring
	if _, err := nilFilter.MarshalBinary(); err == nil {
		t.Fatal("MarshalBinary accepted a nil filter")
	}
	if err := nilFilter.UnmarshalBinary(valid); err == nil {
		t.Fatal("UnmarshalBinary accepted a nil filter")
	}
	if _, err := new(ring.Ring).MarshalBinary(); err == nil {
		t.Fatal("MarshalBinary accepted an uninitialized filter")
	}
}

// FuzzUnmarshalBinary checks that binary decoding is safe and stable.
func FuzzUnmarshalBinary(f *testing.F) {
	filter, err := ring.Init(100, 0.01)
	if err != nil {
		f.Fatal(err)
	}
	filter.Add([]byte("seed"))
	seed, err := filter.MarshalBinary()
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte("not a filter"))

	f.Fuzz(func(t *testing.T, data []byte) {
		var filter ring.Ring
		if err := filter.UnmarshalBinary(data); err != nil {
			return
		}
		encoded, err := filter.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(encoded, data) {
			t.Fatal("binary round trip changed the encoding")
		}
	})
}

// BenchmarkAdd measures insertion into a pre-sized filter.
func BenchmarkAdd(b *testing.B) {
	filter, err := ring.Init(1_000_000, 0.001)
	if err != nil {
		b.Fatal(err)
	}
	data := make([]byte, 8)
	var value uint64
	b.ReportAllocs()
	for b.Loop() {
		binary.LittleEndian.PutUint64(data, value)
		filter.Add(data)
		value++
	}
}

// BenchmarkTest measures membership checks on a populated filter.
func BenchmarkTest(b *testing.B) {
	filter, err := ring.Init(1_000_000, 0.001)
	if err != nil {
		b.Fatal(err)
	}
	data := []byte("hello")
	filter.Add(data)
	b.ReportAllocs()
	for b.Loop() {
		filter.Test(data)
	}
}

// newFilter constructs a filter for a test.
func newFilter(t testing.TB, elements int, falsePositive float64) *ring.Ring {
	t.Helper()
	filter, err := ring.Init(elements, falsePositive)
	if err != nil {
		t.Fatal(err)
	}
	return filter
}

// key returns a stable key from a namespace and integer.
func key(namespace byte, value int) []byte {
	data := make([]byte, 9)
	data[0] = namespace
	binary.LittleEndian.PutUint64(data[1:], uint64(value))
	return data
}

// mergeWithTimeout runs one merge with a deadlock guard.
func mergeWithTimeout(target, source *ring.Ring) error {
	done := make(chan error, 1)
	go func() { done <- target.Merge(source) }()
	select {
	case err := <-done:
		return err
	case <-time.After(time.Second):
		return errors.New("merge deadlocked")
	}
}
