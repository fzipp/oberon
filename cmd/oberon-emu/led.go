// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package main

import "fmt"

type ConsoleLEDs struct{}

func (led *ConsoleLEDs) Write(value uint32) {
	fmt.Print("LEDs: ")
	for i := 7; i >= 0; i-- {
		if value&(1<<i) > 0 {
			fmt.Printf("%d", i)
		} else {
			fmt.Print("-")
		}
	}
	fmt.Println()
}
