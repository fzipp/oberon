// Dumps the ASCII content of Oberon texts.
// Doesn't properly parse the files so doesn't work very well.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	var err error

	in := os.Stdin
	if len(os.Args) > 1 {
		in, err = os.Open(os.Args[1])
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
		case '\r':
			// translate '\r' to '\n'
			err = out.WriteByte('\n')
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

func fail(message interface{}) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
