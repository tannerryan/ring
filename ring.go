// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ring

import (
	"errors"
	"math"
	"sync"
)

// Ring contains the information for a ring data store.
type Ring struct {
	size       int           // size of bit array
	bits       []byte        // main bit array
	hash       int           // number of hash rounds
	bufferSize int           // size of circular data array
	bufferPtr  int           // pointer to last added data point
	buffer     []uint64      // circular data array
	mutex      *sync.RWMutex // mutex for locking Add, Test, and Reset operations
}

// Init initializes and returns a new ring, or an error. Given a number of
// elements, it accurately states if data is not added. Within a falsePositive
// rate, it will indicate if the data has been added. When bufferSize is greater
// than zero, ring will test against the last bufferSize elements.
func Init(elements int, falsePositive float64, bufferSize int) (*Ring, error) {
	r := Ring{}
	// length of filter
	m := (-1 * float64(elements) * math.Log(falsePositive)) / math.Pow(math.Log(2), 2)
	// number of hash rounds
	k := (m / float64(elements)) * math.Log(2)

	// check parameters
	if m <= 0 || k <= 0 || bufferSize < 0 {
		return nil, errors.New("invalid parameters")
	}

	// if bufferSize is greater than 0, generate a circular buffer
	r.bufferSize = bufferSize
	if r.bufferSize > 0 {
		r.buffer = make([]uint64, r.bufferSize)
		r.bufferPtr = 0
	}

	// ring parameters
	r.mutex = &sync.RWMutex{}
	r.size = int(math.Ceil(m))
	r.hash = int(math.Ceil(k))
	r.bits = make([]byte, r.size)
	return &r, nil
}

// Add adds the data to the ring.
func (r *Ring) Add(data []byte) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// generate hashes
	hash := generateMultiHash(data)
	for i := 0; i < r.hash; i++ {
		index := getRound(hash, uint64(i)) % uint64(r.size)
		r.bits[index] = 1
	}
	// add to circular buffer if initialized
	if r.bufferSize > 0 {
		r.buffAdd(hash[0])
	}
}

// Reset clears the ring.
func (r *Ring) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// reset bits and circular buffer
	r.bits = make([]byte, r.size)
	if r.bufferSize > 0 {
		r.buffer = make([]uint64, r.bufferSize)
		r.bufferPtr = 0
	}
}

// Test returns a bool if the data is in the ring. If the buffer is not enabled,
// true indicates that the data may be in the ring, while false indicates that
// the data is not in the ring. If the buffer is enabled, true indicates that
// the data is in the buffer, while false indicates that the data is not in the
// buffer.
func (r *Ring) Test(data []byte) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// generate hashes
	hash := generateMultiHash(data)
	for i := 0; i < r.hash; i++ {
		index := getRound(hash, uint64(i)) % uint64(r.size)
		if r.bits[index] != 1 {
			return false
		}
	}
	// check circular buffer if initialized
	if r.bufferSize > 0 && !r.buffContains(hash[0]) {
		return false
	}
	return true
}

// buffAdd adds the key to the circular buffer. It also advances the buffer
// pointer.
func (r *Ring) buffAdd(key uint64) {
	r.buffer[r.bufferPtr] = key
	r.bufferPtr++
	// wrap pointer back to start (moving right)
	if r.bufferPtr == r.bufferSize {
		r.bufferPtr = 0
	}
}

// buffContains searches the circular buffer for a key. If the key is found,
// true is returned, else false will be returned.
func (r *Ring) buffContains(key uint64) bool {
	// pointer to start (scanning left)
	for i := r.bufferPtr - 1; i >= 0; i-- {
		if key == r.buffer[i] {
			return true
		}
	}
	// end to pointer (scanning left)
	for i := r.bufferSize - 1; i >= r.bufferPtr; i-- {
		if key == r.buffer[i] {
			return true
		}
	}
	return false
}
