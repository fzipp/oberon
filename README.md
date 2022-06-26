# Project Oberon RISC Emulator

This is an emulator for the [Project Oberon (Revised Edition 2013)](https://people.inf.ethz.ch/wirth/ProjectOberon/index.html)
RISC5 processor, written in the [Go programming language](https://golang.org).
It is a port of [Peter De Wachter's emulator](https://github.com/pdewacht/oberon-risc-emu) from C to Go.

Project Oberon is a design for a complete desktop computer system from scratch by
[Niklaus Wirth](https://people.inf.ethz.ch/wirth/) and
[Jürg Gutknecht](https://en.wikipedia.org/wiki/J%C3%BCrg_Gutknecht).
It comprises a RISC CPU, a programming language and an operating system.

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

## Oberon Compiler

Project Oberon contains Wirth's Oberon compiler written in Oberon.
If you're interested in an Oberon compiler written in Go check out
[oberon-compiler](https://github.com/fzipp/oberon-compiler), which is a port
of Wirth's compiler to Go.

## License

This project is free and open source software licensed under the
[ISC License](LICENSE).
