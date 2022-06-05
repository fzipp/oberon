// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package test

// The Verilog (.v) files can be downloaded at
// https://people.inf.ethz.ch/wirth/ProjectOberon/index.html
//go:generate go run v2go/v2go.go -o add/add_gen.go FPAdder.v
//go:generate go run v2go/v2go.go -o flr/add_gen.go FPAdder.v
//go:generate go run v2go/v2go.go -o flt/add_gen.go FPAdder.v
//go:generate go run v2go/v2go.go -o mul/mul_gen.go FPMultiplier.v
//go:generate go run v2go/v2go.go -o div/div_gen.go FPDivider.v
//go:generate go run v2go/v2go.go -o idiv/idiv_gen.go Divider.v

//go:generate go run numbers/numbers.go -o numbers_gen.go
