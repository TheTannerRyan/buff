// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buff_test

import (
	"bytes"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/thetannerryan/buff"
)

const (
	size = 1000 // size of buffers
)

var (
	bufferRecent, _ = buff.Init(size, buff.Recent)
	bufferOldest, _ = buff.Init(size, buff.Oldest)
)

// TestMain performs unit tests and benchmarks.
func TestMain(m *testing.M) {
	// run tests
	rand.Seed(time.Now().UTC().UnixNano())
	ret := m.Run()
	os.Exit(ret)
}

// TestBadParameters ensures that errornous parameters return an error.
func TestBadParameters(t *testing.T) {
	_, err := buff.Init(0, buff.Recent)
	if err == nil {
		t.Fatal("size 0 not captured")
	}
	_, err = buff.Init(-1, buff.Recent)
	if err == nil {
		t.Fatal("size -1 not captured")
	}
	_, err = buff.Init(1, 2)
	if err == nil {
		t.Fatal("invalid mode not captured")
	}
}

// TestReset ensures the buffer is cleared on Reset().
func TestReset(t *testing.T) {
	data := []byte("testing")

	bufferRecent.Add(data)
	bufferOldest.Add(data)

	bufferRecent.Reset()
	bufferOldest.Reset()

	// added data should not be in the buffer anymore
	if bufferRecent.Test(data) || bufferOldest.Test(data) {
		t.Fatalf("data not cleared on Reset()")
	}
}

// TestOverwrite ensures that the oldest data is overwritten (proper wrap
// around).
func TestOverwrite(t *testing.T) {
	data := []byte("testing")
	buff := make([]byte, 4)

	bufferRecent.Add(data)
	bufferOldest.Add(data)

	// loading elements the size of the buffer should bump out the original
	// element
	for i := 0; i < size; i++ {
		intToByte(buff, i)
		bufferRecent.Add(buff)
		bufferOldest.Add(buff)
	}

	// original element should be bumped out after size elements have been added
	if bufferRecent.Test(data) || bufferOldest.Test(data) {
		t.Fatalf("data not properly overwritten when buffer is full")
	}

	// ensure all new elements are present
	for i := 0; i < size; i++ {
		intToByte(buff, i)
		if !bufferRecent.Test(buff) || !bufferOldest.Test(buff) {
			t.Fatalf("elements are missing on wrap around")
		}
	}
}

// TestData ensures that data is properly labeled before and after adding the
// data.
func TestData(t *testing.T) {
	// clear before starting
	bufferRecent.Reset()
	bufferOldest.Reset()

	var buff []byte

	for i := 0; i < size; i++ {
		buff = make([]byte, 4)
		intToByte(buff, i)

		// test that data is not added in the buffer before
		if bufferRecent.Test(buff) || bufferOldest.Test(buff) {
			t.Fatal("data falsely flagged as being in buffer")
		}

		bufferRecent.Add(buff)
		bufferOldest.Add(buff)

		// test that data is in the buffer after
		if !bufferRecent.Test(buff) || !bufferOldest.Test(buff) {
			t.Fatal("data not being flagged as being in buffer")
		}
	}
}

// TestGetRecent checks to ensure that GetRecent() returns the correct data.
func TestGetRecent(t *testing.T) {
	buff := make([]byte, 4)

	// clear before testing
	bufferRecent.Reset()
	bufferOldest.Reset()

	// when empty, the most recent element should be null
	if bufferRecent.GetRecent() != nil || bufferOldest.GetRecent() != nil {
		t.Fatal("most recent element in empty buffer is not nil")
	}

	// after adding the data, it should be returned by GetRecent()
	for i := 0; i < size; i++ {
		intToByte(buff, i)
		bufferRecent.Add(buff)
		bufferOldest.Add(buff)

		if !bytes.Equal(bufferRecent.GetRecent(), buff) || !bytes.Equal(bufferOldest.GetRecent(), buff) {
			t.Fatal("most recent element not returned")
		}
	}
}

// TestGetOldest checks to ensure that GetOldest() returns the correct data.
func TestGetOldest(t *testing.T) {
	data := []byte("testing")
	buff := make([]byte, 4)

	// clear before testing
	bufferRecent.Reset()
	bufferOldest.Reset()

	// when empty, the oldest element should be null
	if bufferRecent.GetOldest() != nil || bufferOldest.GetOldest() != nil {
		t.Fatalf("oldest element in in empty buffer is not null")
	}

	bufferRecent.Add(data)
	bufferOldest.Add(data)

	// after adding the data, the oldest element should be the first element
	// added (before wrapping around)
	for i := 0; i < size-1; i++ {
		intToByte(buff, i)
		bufferRecent.Add(buff)
		bufferOldest.Add(buff)

		if !bytes.Equal(bufferRecent.GetOldest(), data) || !bytes.Equal(bufferOldest.GetOldest(), data) {
			t.Fatalf("oldest element is not returned")
		}
	}

	// adding one more element should bump out the original oldest element (wrap
	// around)
	bufferRecent.Add(data)
	bufferOldest.Add(data)

	if bytes.Equal(bufferRecent.GetOldest(), data) || bytes.Equal(bufferOldest.GetOldest(), data) {
		t.Fatalf("oldest element is not returned")
	}
}

// TestRace is used for testing race condition (requires "-race" flag). By
// spawning multiple goroutines performing "critical actions", this allows Go's
// race dector to detect the presence of a race condition. Don't believe me? Try
// commenting out one of the Lock/Unlock pairs and retesting.
func TestRace(t *testing.T) {
	for i := 0; i < size; i++ {
		// add some elements
		go func(i int) {
			data := make([]byte, 4)
			intToByte(data, i)
			bufferRecent.Add(data)
			bufferOldest.Add(data)
		}(i)

		// test some elements
		go func(i int) {
			data := make([]byte, 4)
			intToByte(data, i)
			bufferRecent.Test(data)
			bufferOldest.Test(data)
		}(i)

		// get the oldest elements
		go func() {
			bufferRecent.GetOldest()
			bufferRecent.GetOldest()
		}()

		// get the most recent elements
		go func() {
			bufferRecent.GetRecent()
			bufferOldest.GetRecent()
		}()

		// reset the buffers
		go func() {
			bufferRecent.Reset()
			bufferOldest.Reset()
		}()
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
