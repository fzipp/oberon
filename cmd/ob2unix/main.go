// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

// Converts the content of Oberon texts (.Text, .Mod, .Tool) to plain text
// with Unix-style line endings.
//
//  - Drops the file header
//  - Converts line endings from CR to LF
//  - Replaces each tab character by two spaces
//  - Removes all other control characters
//  - Converts extended Latin characters with diacritics to UTF-8
//
// Usage:
//    ob2unix [oberon_text_file]
//
// If no file is specified the command reads its input from the standard input.
// The converted plain text is written to the standard output.
// If the input is not an Oberon text it is written to the output unchanged.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func usage() {
	fail(`Converts the content of Oberon texts (.Text, .Mod, .Tool) to plain text
with Unix-style line endings.

 - Drops the file header
 - Converts line endings from CR to LF
 - Replaces each tab character by two spaces
 - Removes all other control characters
 - Converts extended Latin characters with diacritics to UTF-8

Usage:
   ob2unix [oberon_text_file]

If no file is specified the command reads its input from the standard input.
The converted plain text is written to the standard output.
If the input is not an Oberon text it is written to the output unchanged.`)
}

func main() {
	var err error

	flag.Usage = usage
	flag.Parse()

	in := os.Stdin
	if flag.NArg() > 0 {
		in, err = os.Open(flag.Arg(0))
		check(err)
		defer in.Close()
	}
	out := bufio.NewWriter(os.Stdout)

	p, err := io.ReadAll(in)
	check(err)

	if !isOberon(p) {
		// Not an Oberon text file: copy input to output
		_, err = out.Write(p)
		check(err)
		err = out.Flush()
		check(err)
		return
	}

	p = skipHeader(p)
	for _, b := range p {
		switch b {
		case '\r', '\n':
			err = out.WriteByte('\n')
		case '\t':
			_, err = out.Write([]byte("  "))
		case 0x80:
			_, err = out.WriteRune('Ä')
		case 0x81:
			_, err = out.WriteRune('Ö')
		case 0x82:
			_, err = out.WriteRune('Ü')
		case 0x83:
			_, err = out.WriteRune('ä')
		case 0x84:
			_, err = out.WriteRune('ö')
		case 0x85:
			_, err = out.WriteRune('ü')
		case 0x86:
			_, err = out.WriteRune('â')
		case 0x87:
			_, err = out.WriteRune('ê')
		case 0x88:
			_, err = out.WriteRune('î')
		case 0x89:
			_, err = out.WriteRune('ô')
		case 0x8A:
			_, err = out.WriteRune('û')
		case 0x8B:
			_, err = out.WriteRune('à')
		case 0x8C:
			_, err = out.WriteRune('è')
		case 0x8D:
			_, err = out.WriteRune('ì')
		case 0x8E:
			_, err = out.WriteRune('ò')
		case 0x8F:
			_, err = out.WriteRune('ù')
		case 0x90:
			_, err = out.WriteRune('é')
		case 0x91:
			_, err = out.WriteRune('ë')
		case 0x92:
			_, err = out.WriteRune('ï')
		case 0x93:
			_, err = out.WriteRune('ç')
		case 0x94:
			_, err = out.WriteRune('á')
		case 0x95:
			_, err = out.WriteRune('ñ')
		case 0xAB:
			_, err = out.WriteRune('ß')
		default:
			if b < 32 {
				continue
			}
			err = out.WriteByte(b)
		}
		check(err)
	}
	err = out.Flush()
	check(err)
}

func isOberon(p []byte) bool {
	return len(p) >= 6 && (p[0] == 240 && p[1] == 1) || (p[0] == 1 && p[1] == 240)
}

func skipHeader(p []byte) []byte {
	headerSize := uint64(p[2]) | uint64(p[3])<<8 | uint64(p[4])<<16 | uint64(p[5])<<24
	return p[headerSize:]
}

func check(err error) {
	if err != nil {
		fail(err)
	}
}

func fail(message interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
