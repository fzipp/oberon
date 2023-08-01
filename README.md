# Project Oberon RISC Emulator

This is an emulator for the
[Project Oberon (Revised Edition 2013)](https://people.inf.ethz.ch/wirth/ProjectOberon/index.html)
RISC processor,
written in the [Go programming language](https://go.dev).
It is a port of
[Peter De Wachter's C-based emulator](https://github.com/pdewacht/oberon-risc-emu)
to Go.

Project Oberon is a design for a complete desktop computer system from scratch,
created by [Niklaus Wirth](https://people.inf.ethz.ch/wirth/)
and [Jürg Gutknecht](https://en.wikipedia.org/wiki/J%C3%BCrg_Gutknecht).
Its simplicity and clarity enable a single person
to understand and implement the entire system,
making Project Oberon an excellent educational tool.
It consists of:
- A RISC CPU design.
- A programming language with a compiler written in itself.
- An operating system with a text-oriented,
  mouse-controlled graphical user interface,
  written in the Oberon programming language.

If you find this project interesting,
you should also explore
[this Oberon compiler in Go](https://github.com/fzipp/oberon-compiler),
which is a direct port of Wirth's compiler for the RISC architecture
from Oberon to the Go programming language.
Additionally, you can still use the original compiler
within the emulator itself.

## Install

```
$ go install github.com/fzipp/oberon/cmd/oberon-emu@latest
```

## Run

First, download an Oberon disk image (.dsk file), e.g. from
[this repository](https://github.com/pdewacht/oberon-risc-emu/tree/master/DiskImage).

Next, initiate the emulator by providing the disk image file
as a command-line argument:

```
$ oberon-emu Oberon-2020-08-18.dsk
Visit http://localhost:8080 in a web browser
```
Open the following link in your web browser: http://localhost:8080.

This is the Project Oberon user interface right after starting:

![Project Oberon](doc/screenshot1.png?raw=true "Project Oberon directly after start")

## Using Oberon

[How to use the Oberon System](https://people.inf.ethz.ch/wirth/ProjectOberon/UsingOberon.pdf) (PDF)

Oberon's user interface is designed for use with a three-button mouse,
but the emulator also provides the option to
simulate all three mouse buttons via the keyboard.

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
in Wirth's series of language designs,
with predecessors like Pascal and Modula.
Additionally, there are various versions and dialects
of these three languages.
Wirth's goal has always been
to simplify the language whenever possible.

Oberon's grammar is designed in a way that allows for
implementation with a single-pass compiler.
Oberon code is organized into modules
that can be compiled separately.
The language is statically typed
and incorporates a garbage collector for memory management.
Access to low-level and unsafe facilities is possible
through a designated SYSTEM module.

Furthermore, Oberon's influence extends to some aspects of Go,
as Robert Griesemer,
one of the original creators of Go,
[explains in this GopherCon 2015 talk](https://www.youtube.com/watch?v=0ReKdcpNyQg&t=1070s)
"The Evolution of Go".

## License

This project is free and open source software licensed under the
[ISC License](LICENSE).
