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

func vIdiv(inX, inY uint32, signedDiv bool) fp.IdivResult {
	x = uint64(inX)
	y = uint64(inY)
	u = 0
	if signedDiv {
		u = 1
	}
	run = 0
	cycle()
	run = 1
	for {
		cycle()
		if stall() == 0 {
			break
		}
	}
	return fp.IdivResult{Quot: uint32(quot()), Rem: uint32(rem())}
}

func main() {
	count := 0
	errors := 0
	for _, s := range []bool{false, true} {
		for i, a := range test.Numbers {
			for _, b := range test.Numbers {
				want := vIdiv(a, b, s)
				got := fp.Idiv(a, b, s)
				if got != want {
					if errors < 20 {
						fmt.Printf("idiv: fp.Idiv(%08x, %d, %v) = %v, want (Verilog): %v\n", a, b, s, got, want)
					}
					errors++
				}
				count++
			}
			if (i % 500) == 0 {
				p := count * 100 / len(test.Numbers) / len(test.Numbers) / 2
				fmt.Printf("idiv: %d%% (%d errors)\n", p, errors)
			}
		}
	}
	fmt.Printf("idiv: errors: %d tests: %d\n", errors, count)
	if errors > 0 {
		os.Exit(1)
	}
}
