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
	count := 0
	errors := 0
	for i, a := range test.Numbers {
		for _, b := range test.Numbers {
			want := vAdd(a, b)
			got := fp.Add(a, b, false, false)
			if got != want {
				if errors < 10 {
					fmt.Printf("add: fp.Add(%08x, %08x, false, false) = %08x, want (Verilog): %08x\n", a, b, got, want)
				}
				errors++
			}
			count++
		}
		if (i % 500) == 0 {
			p := count * 100 / len(test.Numbers) / len(test.Numbers)
			fmt.Printf("add: %d%% (%d errors)\n", p, errors)
		}
	}
	fmt.Printf("add: errors: %d tests: %d\n", errors, count)
	if errors > 0 {
		os.Exit(1)
	}
}
