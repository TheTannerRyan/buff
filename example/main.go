// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/thetannerryan/buff"
)

func main() {
	// create a circular buffer of size 100 in recent mode (searches the buffer
	// from the most recent to oldest element)
	buffer, err := buff.Init(100, buff.Recent)
	if err != nil {
		// error will only occur if size is less than 1, or if incorrect mode is
		// provided
		panic(err)
	}

	data := []byte("hello")

	// check if data is in buffer
	fmt.Printf("%s in buffer :: %t\n", data, buffer.Test(data))

	// add data to buffer
	buffer.Add(data)
	fmt.Printf("%s in buffer :: %t\n", data, buffer.Test(data))

	buffer.Add([]byte("hello2"))
	buffer.Add([]byte("hello3"))

	// get the most recent and oldest elements
	fmt.Printf("most recent :: %s\n", buffer.GetRecent())
	fmt.Printf("oldest :: %s\n", buffer.GetOldest())

	// reset buffer
	buffer.Reset()
	fmt.Printf("%s in buffer :: %t\n", data, buffer.Test(data))
}
