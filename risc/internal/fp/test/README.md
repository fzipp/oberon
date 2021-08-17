This directory contains test infrastructure for the floating-point arithmetic
of the RISC emulator.

The tests are run by transpiling the original Verilog sources to Go code and
then comparing the emulator implementation against the Verilog implementation.

Set the `VERILOG` environment variable to to the directory with the Verilog
sources. You can download the Verilog sources from:
http://people.inf.ethz.ch/wirth/ProjectOberon/index.html

The following files are required:

    Divider.v  FPAdder.v  FPDivider.v  FPMultiplier.v

Use `go generate` to transpile the Verilog sources to Go:

    $ go generate

The tests are implemented as main programs in the following sub-directories:

    add  div  idiv  mul  flr  flt

