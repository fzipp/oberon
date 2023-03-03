# Project Oberon RISC Emulator

This is an emulator for the [Project Oberon (Revised Edition 2013)](https://people.inf.ethz.ch/wirth/ProjectOberon/index.html)
RISC processor, written in the [Go programming language](https://golang.org).
It is a port of [Peter De Wachter's emulator](https://github.com/pdewacht/oberon-risc-emu) from C to Go.

Project Oberon is a design for a complete desktop computer system from scratch by
[Niklaus Wirth](https://people.inf.ethz.ch/wirth/) and
[Jürg Gutknecht](https://en.wikipedia.org/wiki/J%C3%BCrg_Gutknecht).
Its simplicity and clarity enables a single person
to know and implement the whole system.
This makes Project Oberon a great educational tool. It consists of:
- a RISC CPU design
- a programming language with a compiler written in itself
- an operating system with a text-oriented, mouse-controlled graphical user interface,
  written in the Oberon programming language.

If you like this project then you should also check out
[this Oberon compiler in Go](https://github.com/fzipp/oberon-compiler)
which is a direct port of Wirth's compiler for the RISC architecture
from Oberon to the Go programming language.
Of course, you can also use the original compiler in the emulator itself.

## Install

```
$ go install github.com/fzipp/oberon/cmd/oberon-emu@latest
```

## Run

First download an Oberon disk image (.dsk file), e.g. from
[this repository](https://github.com/pdewacht/oberon-risc-emu/tree/master/DiskImage).

Then start the emulator with the disk image file as command line argument:

```
$ oberon-emu Oberon-2020-08-18.dsk
Visit http://localhost:8080 in a web browser
```
Open the link http://localhost:8080 in a web browser.

This is the Project Oberon user interface directly after start:

![Project Oberon](doc/screenshot1.png?raw=true "Project Oberon directly after start")

## Using Oberon

[How to use the Oberon System](https://people.inf.ethz.ch/wirth/ProjectOberon/UsingOberon.pdf) (PDF)

Oberon's user interface is designed for use with a three-button mouse,
but the emulator additionally allows to simulate all three mouse buttons via
the keyboard.

| Mouse button | Function           | Mac keyboard | PC keyboard |
|--------------|--------------------|--------------|-------------|
| Left         | Set caret (cursor) | ⌃ control    | Ctrl        |
| Middle       | Execute command    | ⌥ option     | Alt         |
| Right        | Select text        | ⌘ command    | Meta (Win)  |


| Key   | Function            |
|-------|---------------------|
| Esc   | Undo all selections |
| F1    | Set global marker   |

## About the Oberon language

Oberon is the latest programming language
in Wirth's succession of language designs,
with predecessors like Pascal and Modula.
There are also various versions and dialects of these three languages.
Wirth strives to make the language simpler whenever possible.

Oberon's grammar is designed in such a way that it is possible to
implement it with a single-pass compiler.
Oberon code is structured as modules that can be compiled separately.
The language is statically typed and it uses a garbage collector for memory management.
Access to low-level and unsafe facilities is possible through a
designated SYSTEM module.

Oberon also influenced some aspects of Go, as Robert Griesemer,
one of the original creators of Go,
[explains in his GopherCon 2015 talk](https://www.youtube.com/watch?v=0ReKdcpNyQg&t=1070s)
"The Evolution of Go".

## License

This project is free and open source software licensed under the
[ISC License](LICENSE).
