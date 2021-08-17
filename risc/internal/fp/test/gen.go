// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package test

// The Verilog (.v) files can be downloaded at
// http://people.inf.ethz.ch/wirth/ProjectOberon/index.html
//go:generate go run v2go/v2go.go -o add/add_gen.go /Users/frederik/Downloads/verilog/FPAdder.v
//go:generate go run v2go/v2go.go -o mul/mul_gen.go /Users/frederik/Downloads/verilog/FPMultiplier.v
//go:generate go run v2go/v2go.go -o div/div_gen.go /Users/frederik/Downloads/verilog/FPDivider.v
//go:generate go run v2go/v2go.go -o idiv/idiv_gen.go /Users/frederik/Downloads/verilog/Divider.v

//go:generate go run numbers/numbers.go -o numbers_gen.go
