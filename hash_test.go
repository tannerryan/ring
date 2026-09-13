// Copyright (c) 2019 Tanner Ryan. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

package ring

import (
	"strconv"
	"testing"
)

// TestMurmur128 checks values from the MurmurHash3 reference implementation.
func TestMurmur128(t *testing.T) {
	for _, test := range []struct {
		input string
		h1    uint64
		h2    uint64
	}{
		{"", 0x0000000000000000, 0x0000000000000000},
		{"hello", 0xcbd8a7b341bd9b02, 0x5b1e906a48ae1d19},
		{"hello, world", 0x342fac623a5ebc8e, 0x4cdcbc079642414d},
		{"The quick brown fox jumps over the lazy dog.", 0xcd99481f9ee902c9, 0x695da1a38987b6e7},
	} {
		t.Run(test.input, func(t *testing.T) {
			h1, h2 := murmur128([]byte(test.input), false)
			if h1 != test.h1 || h2 != test.h2 {
				t.Fatalf("murmur128() = %016x %016x, want %016x %016x", h1, h2, test.h1, test.h2)
			}
		})
	}
}

// TestMurmur128Legacy records the hash used by version 1 persisted filters.
func TestMurmur128Legacy(t *testing.T) {
	h1, h2 := murmur128([]byte("The quick brown fox jumps over the lazy dog."), true)
	if h1 != 0x2a0d842046b3dbd4 || h2 != 0x0a91a598b05d8046 {
		t.Fatalf("legacy hash changed: %016x %016x", h1, h2)
	}
}

// TestMurmur128WithSuffix checks the allocation-free path against an explicit
// appended byte at every block boundary.
func TestMurmur128WithSuffix(t *testing.T) {
	for length := 0; length <= 64; length++ {
		data := make([]byte, length)
		for i := range data {
			data[i] = byte(i*31 + length)
		}
		appended := append(append([]byte(nil), data...), multiHashSuffix)
		for _, legacy := range []bool{false, true} {
			want1, want2 := murmur128(appended, legacy)
			got1, got2 := murmur128WithSuffix(data, legacy)
			if got1 != want1 || got2 != want2 {
				t.Fatalf("length %d, legacy %t: got %016x %016x, want %016x %016x", length, legacy, got1, got2, want1, want2)
			}
		}
	}
}

// TestGenerateMultiHash checks that hashing does not change caller data.
func TestGenerateMultiHash(t *testing.T) {
	data := []byte{0x00, 0x12, 0x34, 0x56, 0x78, 0x00}
	want := append([]byte(nil), data...)
	generateMultiHash(data[1:5], binaryVersion)
	for i := range data {
		if data[i] != want[i] {
			t.Fatalf("input changed at byte %d", i)
		}
	}
}

// BenchmarkGenerateMultiHash measures the four-hash generator.
func BenchmarkGenerateMultiHash(b *testing.B) {
	for _, size := range []int{8, 64, 1024, 8192} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			data := make([]byte, size)
			b.SetBytes(int64(size))
			b.ReportAllocs()
			for b.Loop() {
				generateMultiHash(data, binaryVersion)
			}
		})
	}
}
