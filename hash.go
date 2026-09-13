// Copyright (c) 2019 Tanner Ryan. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

package ring

const (
	// murmur64c1 is the first block mixing constant.
	murmur64c1 uint64 = 0x87c37b91114253d5
	// murmur64c2 is the second block mixing constant.
	murmur64c2 uint64 = 0x4cf5ad432745937f
	// murmur64c3 is the first block addition constant.
	murmur64c3 uint64 = 0x52dce729
	// murmur64c4 is the second block addition constant.
	murmur64c4 uint64 = 0x38495ab5
	// murmur64c5 is the first finalization constant.
	murmur64c5 uint64 = 0xff51afd7ed558ccd
	// murmur64c6 is the second finalization constant.
	murmur64c6 uint64 = 0xc4ceb9fe1a85ec53

	// multiHashSuffix separates the second pair of base hashes.
	multiHashSuffix byte = 1
)

// murmur128 returns two 64-bit outputs of a 128-bit MurmurHash3 hash. Legacy
// mode preserves the version 1 partial-block behavior.
func murmur128(data []byte, legacy bool) (uint64, uint64) {
	return murmur128Internal(data, legacy, false)
}

// murmur128WithSuffix hashes data followed by multiHashSuffix without copying
// the complete input.
func murmur128WithSuffix(data []byte, legacy bool) (uint64, uint64) {
	return murmur128Internal(data, legacy, true)
}

// murmur128Internal implements MurmurHash3 with an optional virtual suffix.
func murmur128Internal(data []byte, legacy, suffix bool) (uint64, uint64) {
	var h1, h2, k1, k2 uint64
	blocks := len(data) / 16

	for i := range blocks {
		k1 = bytesToUint64(data[i*16:])
		k2 = bytesToUint64(data[(i*16)+8:])
		h1, h2, k1, k2 = murmurBlock(h1, h2, k1, k2)
	}

	var tailData [16]byte
	tailLength := copy(tailData[:], data[blocks*16:])
	totalLength := uint64(len(data))
	if suffix {
		tailData[tailLength] = multiHashSuffix
		tailLength++
		totalLength++
		if tailLength == len(tailData) {
			k1 = bytesToUint64(tailData[:])
			k2 = bytesToUint64(tailData[8:])
			h1, h2, k1, k2 = murmurBlock(h1, h2, k1, k2)
			tailLength = 0
		}
	}
	if !legacy {
		k1, k2 = 0, 0
	}

	switch tailLength {
	case 15:
		k2 ^= uint64(tailData[14]) << 48
		fallthrough
	case 14:
		k2 ^= uint64(tailData[13]) << 40
		fallthrough
	case 13:
		k2 ^= uint64(tailData[12]) << 32
		fallthrough
	case 12:
		k2 ^= uint64(tailData[11]) << 24
		fallthrough
	case 11:
		k2 ^= uint64(tailData[10]) << 16
		fallthrough
	case 10:
		k2 ^= uint64(tailData[9]) << 8
		fallthrough
	case 9:
		k2 ^= uint64(tailData[8])
		k2 *= murmur64c2
		k2 = (k2 << 33) | (k2 >> (64 - 33))
		k2 *= murmur64c1
		h2 ^= k2
		fallthrough

	case 8:
		k1 ^= uint64(tailData[7]) << 56
		fallthrough
	case 7:
		k1 ^= uint64(tailData[6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(tailData[5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(tailData[4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(tailData[3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(tailData[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(tailData[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(tailData[0])
		k1 *= murmur64c1
		k1 = (k1 << 31) | (k1 >> (64 - 31))
		k1 *= murmur64c2
		h1 ^= k1
	}

	h1 ^= totalLength
	h2 ^= totalLength

	h1 += h2
	h2 += h1

	h1 = fmix(h1)
	h2 = fmix(h2)

	h1 += h2
	h2 += h1

	return h1, h2
}

// murmurBlock mixes one 16-byte block into the running hash state.
func murmurBlock(h1, h2, k1, k2 uint64) (uint64, uint64, uint64, uint64) {
	k1 *= murmur64c1
	k1 = (k1 << 31) | (k1 >> (64 - 31))
	k1 *= murmur64c2
	h1 ^= k1

	h1 = (h1 << 27) | (h1 >> (64 - 27))
	h1 += h2
	h1 = h1*5 + murmur64c3

	k2 *= murmur64c2
	k2 = (k2 << 33) | (k2 >> (64 - 33))
	k2 *= murmur64c1
	h2 ^= k2

	h2 = (h2 << 31) | (h2 >> (64 - 31))
	h2 += h1
	h2 = h2*5 + murmur64c4
	return h1, h2, k1, k2
}

// fmix is the 64-bit MurmurHash3 finalizer to avalanche bits.
func fmix(h uint64) uint64 {
	h ^= h >> 33
	h *= murmur64c5
	h ^= h >> 33
	h *= murmur64c6
	h ^= h >> 33
	return h
}

// bytesToUint64 reads a little-endian uint64.
func bytesToUint64(b []byte) uint64 {
	_ = b[7] // memory safety
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

// generateMultiHash returns four 64-bit values from two MurmurHash3 hashes.
func generateMultiHash(data []byte, version byte) [4]uint64 {
	legacy := version == legacyBinaryVersion
	h1, h2 := murmur128(data, legacy)
	h3, h4 := murmur128WithSuffix(data, legacy)
	return [4]uint64{h1, h2, h3, h4}
}

// getRound derives hash round n from four base values.
func getRound(hash [4]uint64, n uint64) uint64 {
	index := 2 + (((n + (n % 2)) % 4) / 2)
	pre := hash[n%2]
	post := n * hash[index]
	return pre + post
}
