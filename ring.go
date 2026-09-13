// Copyright (c) 2019 Tanner Ryan. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

package ring

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
)

const (
	// legacyBinaryVersion identifies filters made with the original hash.
	legacyBinaryVersion byte = 1
	// binaryVersion identifies filters made with the corrected hash.
	binaryVersion byte = 2
	// binaryHeaderSize is the version, size, and hash-round header length.
	binaryHeaderSize = 17
)

var (
	// errElements reports an invalid expected element count.
	errElements = errors.New("error: elements must be greater than 0")
	// errFalsePositive reports an invalid false-positive rate.
	errFalsePositive = errors.New("error: falsePositive must be greater than 0 and less than 1")
	// errInvalidRing reports an uninitialized or corrupt ring.
	errInvalidRing = errors.New("ring is not initialized or is invalid")
	// errIncompatibleRings reports filters with different parameters.
	errIncompatibleRings = errors.New("rings must have the same m/k parameters")
	// errIncompatibleVersions reports filters with different hash behavior.
	errIncompatibleVersions = errors.New("rings must have the same binary version")
)

// Ring is a thread-safe Bloom filter. A Ring must be created with Init or
// decoded with UnmarshalBinary before use. Ring values should not be copied.
type Ring struct {
	// version selects the binary format and hash behavior.
	version byte
	// size is the number of usable bits.
	size uint64
	// bits stores the filter bits.
	bits []byte
	// hash is the number of hash rounds.
	hash uint64
	// mutex protects the filter fields.
	mutex *sync.RWMutex
}

// Init returns a filter sized for a positive element count and a false-positive
// rate between 0 and 1.
func Init(elements int, falsePositive float64) (*Ring, error) {
	if elements <= 0 {
		return nil, errElements
	}
	if math.IsNaN(falsePositive) || falsePositive <= 0 || falsePositive >= 1 {
		return nil, errFalsePositive
	}

	// Calculate the number of bits and hash rounds using the standard Bloom
	// filter estimates.
	m := (-1 * float64(elements) * math.Log(falsePositive)) / (math.Ln2 * math.Ln2)
	if math.IsInf(m, 0) || m >= float64(math.MaxUint64) {
		return nil, errors.New("ring size is too large")
	}
	size := uint64(math.Ceil(m))
	hash := uint64(math.Ceil((float64(size) / float64(elements)) * math.Ln2))
	return newRing(size, hash)
}

// InitByParameters returns a filter with the given bit count and hash rounds.
// Both values must be positive, and hashes cannot exceed size.
func InitByParameters(size, hashes uint64) (*Ring, error) {
	return newRing(size, hashes)
}

// Add records data in the filter.
func (r *Ring) Add(data []byte) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	hash := generateMultiHash(data, r.version)
	for i := uint64(0); i < r.hash; i++ {
		index := getRound(hash, i) % r.size
		r.bits[index/8] |= (1 << (index % 8))
	}
}

// Reset clears the ring.
func (r *Ring) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	clear(r.bits)
}

// Test reports whether data may be in the filter. A false result rules out
// membership.
func (r *Ring) Test(data []byte) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	hash := generateMultiHash(data, r.version)
	for i := uint64(0); i < r.hash; i++ {
		index := getRound(hash, i) % r.size
		if (r.bits[index/8] & (1 << (index % 8))) == 0 {
			return false
		}
	}
	return true
}

// Merge adds another filter with the same parameters and binary version to r.
func (r *Ring) Merge(m *Ring) error {
	if r == nil || m == nil || r.mutex == nil || m.mutex == nil {
		return errInvalidRing
	}
	if r == m {
		return nil
	}

	m.mutex.RLock()
	version, size, hash := m.version, m.size, m.hash
	bits := append([]byte(nil), m.bits...)
	m.mutex.RUnlock()

	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.version != version {
		return errIncompatibleVersions
	}
	if r.size != size || r.hash != hash || len(r.bits) != len(bits) {
		return errIncompatibleRings
	}
	for i := range bits {
		r.bits[i] |= bits[i]
	}
	return nil
}

// MarshalBinary encodes the filter. New filters use format version 2.
func (r *Ring) MarshalBinary() ([]byte, error) {
	if r == nil || r.mutex == nil {
		return nil, errInvalidRing
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	length, err := bitBytes(r.size)
	if err != nil || !validBinaryVersion(r.version) || r.hash == 0 || r.hash > r.size || len(r.bits) != length {
		return nil, errInvalidRing
	}
	out := make([]byte, len(r.bits)+binaryHeaderSize)
	out[0] = r.version
	binary.BigEndian.PutUint64(out[1:9], r.size)
	binary.BigEndian.PutUint64(out[9:binaryHeaderSize], r.hash)
	copy(out[binaryHeaderSize:], r.bits)
	return out, nil
}

// UnmarshalBinary replaces the filter with a valid version 1 or 2 encoding.
func (r *Ring) UnmarshalBinary(data []byte) error {
	if r == nil {
		return errInvalidRing
	}
	if len(data) < binaryHeaderSize+1 {
		return fmt.Errorf("incorrect length: %d", len(data))
	}
	if !validBinaryVersion(data[0]) {
		return fmt.Errorf("unexpected version: %d", data[0])
	}
	size := binary.BigEndian.Uint64(data[1:9])
	hash := binary.BigEndian.Uint64(data[9:binaryHeaderSize])
	length, err := bitBytes(size)
	if err != nil || hash == 0 || hash > size {
		return errors.New("invalid ring parameters")
	}
	if len(data) != binaryHeaderSize+length {
		return fmt.Errorf("incorrect length: %d", len(data))
	}
	bits := append([]byte(nil), data[binaryHeaderSize:]...)

	if r.mutex == nil {
		r.mutex = new(sync.RWMutex)
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.version = data[0]
	r.size = size
	r.hash = hash
	r.bits = bits
	return nil
}

// newRing creates a filter from validated parameters.
func newRing(size, hash uint64) (*Ring, error) {
	length, err := bitBytes(size)
	if err != nil || hash == 0 || hash > size {
		return nil, errors.New("invalid ring parameters")
	}
	return &Ring{
		version: binaryVersion,
		size:    size,
		bits:    make([]byte, length),
		hash:    hash,
		mutex:   new(sync.RWMutex),
	}, nil
}

// validBinaryVersion reports whether a format version is supported.
func validBinaryVersion(version byte) bool {
	return version == legacyBinaryVersion || version == binaryVersion
}

// bitBytes returns the byte length used by the binary format.
func bitBytes(size uint64) (int, error) {
	if size == 0 {
		return 0, errors.New("ring size must be greater than 0")
	}
	length := size/8 + 1
	maxInt := uint64(math.MaxInt)
	if length > maxInt-binaryHeaderSize {
		return 0, errors.New("ring size is too large")
	}
	return int(length), nil
}
