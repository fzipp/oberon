# Tests for floating-point arithmetic

This directory contains test infrastructure for the floating-point arithmetic
operations of the RISC emulator.

The tests are run by transpiling the original Verilog sources to Go code and
then comparing the emulator implementation against the Verilog implementation.

Set the `VERILOG` environment variable to the directory with the Verilog
sources. You can download the Verilog sources from:
https://people.inf.ethz.ch/wirth/ProjectOberon/index.html

The following files are required:

    Divider.v  FPAdder.v  FPDivider.v  FPMultiplier.v

Use `go generate` to transpile the Verilog sources to Go:

    $ go generate

The tests are implemented as main programs in the following sub-directories:

    add  div  flr  flt  idiv  mul

Run the tests:

    $ go run add/*.go
    $ go run div/*.go
    $ go run flr/*.go
    $ go run flt/*.go
    $ go run idiv/*.go
    $ go run mul/*.go
