// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

// Command ob2unix converts the content of Oberon texts (.Text, .Mod, .Tool)
// to plain text with Unix-style line endings.
//
//   - Drops the file header
//   - Converts line endings from CR to LF
//   - Replaces each tab character by two spaces
//   - Removes all other control characters
//   - Converts extended Latin characters with diacritics to UTF-8
//
// Usage:
//
//	ob2unix [oberon_text_file]
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

var charMapping = [...]rune{
	'Ä', 'Ö', 'Ü',
	'ä', 'ö', 'ü',
	'â', 'ê', 'î', 'ô', 'û',
	'à', 'è', 'ì', 'ò', 'ù',
	'é', 'ë', 'ï', 'ç', 'á', 'ñ',
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
		switch {
		case b == '\r', b == '\n':
			err = out.WriteByte('\n')
		case b == '\t':
			_, err = out.Write([]byte("  "))
		case b < 32:
			continue
		case 0x80 <= b && b <= 0x95:
			_, err = out.WriteRune(charMapping[b-0x80])
		case b == 0xAB:
			_, err = out.WriteRune('ß')
		default:
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

func fail(message any) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
