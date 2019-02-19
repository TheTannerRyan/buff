// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buff

import (
	"bytes"
	"errors"
	"sync"
)

var (
	errSize = errors.New("error: size of buff must be greater than 0")
	errMode = errors.New("error: invalid search mode specified")
)

// Mode represents the search mode.
type Mode uint8

const (
	// Recent searches the buffer from the most recent to oldest element.
	Recent Mode = 0
	// Oldest searches the buffer from the oldest to most recent element.
	Oldest Mode = 1
)

// Buff contains the information for the circular buffer.
type Buff struct {
	size  int           // size of buffer
	mode  Mode          // search mode of buffer
	ptr   int           // pointer to last added data point
	data  [][]byte      // byte store of buffer
	mutex *sync.RWMutex // mutex for locking Add, Test, and Reset operations
}

// Init initializes and returns a new circular buffer. If the size is less than
// one, or if an incorrect mode is provided, an error will be returned.
func Init(size int, mode Mode) (*Buff, error) {
	if size < 1 {
		return nil, errSize
	}
	if mode != 0 && mode != 1 {
		return nil, errMode
	}

	b := Buff{}
	b.size = size
	b.mode = mode
	b.ptr = 0
	b.data = make([][]byte, size)
	b.mutex = &sync.RWMutex{}

	return &b, nil
}

// Add adds data to the buffer.
func (b *Buff) Add(data []byte) {
	b.mutex.Lock()

	// add data and increment pointer
	b.data[b.ptr] = data
	b.ptr++

	// wrap pointer back if at end
	if b.ptr == b.size {
		b.ptr = 0
	}

	b.mutex.Unlock()
}

// Test returns a bool if the data is in the buffer.
func (b *Buff) Test(key []byte) bool {
	if b.mode == Oldest {
		return b.testOldest(key)
	}
	return b.testRecent(key)
}

// Reset clears the buffer.
func (b *Buff) Reset() {
	b.mutex.Lock()
	b.data = make([][]byte, b.size)
	b.ptr = 0
	b.mutex.Unlock()
}

// GetRecent returns the most recent element.
func (b *Buff) GetRecent() []byte {
	b.mutex.RLock()
	var buff []byte

	// if pointer at 0, get last element in data
	if b.ptr == 0 {
		data := b.data[b.size-1]
		// copy does not return nil (if data is nil, must return nil)
		if data == nil {
			b.mutex.RUnlock()
			return nil
		}
		// data is not nil, copy and return
		buff = make([]byte, len(data))
		copy(buff, data)
		b.mutex.RUnlock()
		return buff
	}

	// pointer not at 0, get element before pointer
	data := b.data[b.ptr-1]
	buff = make([]byte, len(data))
	copy(buff, data)
	b.mutex.RUnlock()
	return buff
}

// GetOldest returns the oldest element. Nil is returned if all of the data is
// nil.
func (b *Buff) GetOldest() []byte {
	b.mutex.RLock()
	var buff []byte

	// pointer to end (scanning right)
	for i := b.ptr; i < b.size; i++ {
		if b.data[i] != nil {
			data := b.data[i]
			buff = make([]byte, len(data))
			copy(buff, data)
			b.mutex.RUnlock()
			return buff
		}
	}

	// start to pointer (scanning right)
	for i := 0; i < b.ptr; i++ {
		if b.data[i] != nil {
			data := b.data[i]
			buff = make([]byte, len(data))
			copy(buff, data)
			b.mutex.RUnlock()
			return buff
		}
	}

	b.mutex.RUnlock()
	return nil
}

// testRecent tests for the key in the buffer, starting at the most recent
// element.
func (b *Buff) testRecent(key []byte) bool {
	b.mutex.RLock()

	// pointer to start (scanning left)
	for i := b.ptr - 1; i >= 0; i-- {
		if bytes.Equal(key, b.data[i]) {
			b.mutex.RUnlock()
			return true
		}
	}

	// end to pointer (scanning left)
	for i := b.size - 1; i >= b.ptr; i-- {
		if bytes.Equal(key, b.data[i]) {
			b.mutex.RUnlock()
			return true
		}
	}

	b.mutex.RUnlock()
	return false
}

// testOldest tests for the key in the buffer, starting at the oldest element.
func (b *Buff) testOldest(key []byte) bool {
	b.mutex.RLock()

	// pointer to end (scanning right)
	for i := b.ptr; i < b.size; i++ {
		if bytes.Equal(key, b.data[i]) {
			b.mutex.RUnlock()
			return true
		}
	}

	// start to pointer (scanning right)
	for i := 0; i < b.ptr; i++ {
		if bytes.Equal(key, b.data[i]) {
			b.mutex.RUnlock()
			return true
		}
	}

	b.mutex.RUnlock()
	return false
}
