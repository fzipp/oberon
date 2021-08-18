// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/fzipp/oberon/risc/internal/fp"
	"github.com/fzipp/oberon/risc/internal/fp/test"
)

func vAdd(inX, inY uint32) uint32 {
	x = uint64(inX)
	y = uint64(inY)
	run = 0
	cycle()
	run = 1
	for {
		cycle()
		if stall() == 0 {
			break
		}
	}
	return uint32(z())
}

func main() {
	u = 1
	count := 0
	errors := 0
	for _, a := range test.Numbers {
		b := uint32(0x4B00 << 16)
		want := vAdd(a, b)
		got := fp.Add(a, b, true, false)
		if got != want {
			if errors < 10 {
				fmt.Printf("flt: fp.Add(%08x, %08x, true, false) = %08x, want (Verilog): %08x\n", a, b, got, want)
			}
			errors++
		}
		count++
	}
	fmt.Printf("flt: errors: %d tests: %d\n", errors, count)
	if errors > 0 {
		os.Exit(1)
	}
}
