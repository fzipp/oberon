// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

// Command asciidecoder decodes Oberon 'AsciiCoder.DecodeFiles' archives.
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func usage() {
	fail("Usage: asciidecoder [-v] [-C dir] [file]")
}

func main() {
	var err error

	verbose := flag.Bool("v", false, "verbose output: prints the name of each extracted file")
	directory := flag.String("C", "", "output `directory`, created if it does not exist yet")
	flag.Usage = usage
	flag.Parse()

	in := os.Stdin
	if flag.NArg() > 0 {
		in, err = os.Open(flag.Arg(0))
		check(err)
		defer in.Close()
	}

	if *directory != "" {
		err = os.MkdirAll(*directory, os.ModePerm)
		check(err)
	}

	r := bufio.NewReader(in)
	filenames, compressed, err := readHeader(r)
	check(err)

	for _, filename := range filenames {
		if *verbose {
			fmt.Println(filename)
		}
		path := filename
		if *directory != "" {
			path = filepath.Join(*directory, filename)
		}
		err = extractFile(r, path, compressed)
		check(err)
	}
}

var errInvalidArchive = errors.New("not an AsciiCoder.DecodeFiles archive")

func readHeader(r *bufio.Reader) (filenames []string, compressed bool, err error) {
	command, err := r.ReadString('~')
	if err != nil {
		return nil, false, errInvalidArchive
	}
	args := strings.Fields(command)
	if len(args) < 1 || args[0] != "AsciiCoder.DecodeFiles" {
		return nil, false, errInvalidArchive
	}
	for _, arg := range args[1:] {
		if arg == "%" {
			compressed = true
			continue
		}
		if arg == "~" {
			break
		}
		filenames = append(filenames, arg)
	}
	return filenames, compressed, nil
}

func extractFile(r *bufio.Reader, path string, compressed bool) error {
	data, err := decode(r)
	if err != nil {
		return fmt.Errorf("could not decode file: %w", err)
	}
	if compressed {
		data, err = decompress(bufio.NewReader(bytes.NewBuffer(data)))
		if err != nil {
			return fmt.Errorf("could not decompress file: %w", err)
		}
	}
	err = os.WriteFile(path, data, 0o666)
	if err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}
	return nil
}

func decode(r *bufio.Reader) (data []byte, err error) {
	const base = 48
	bits := 0
	var buf uint32
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if b <= 32 {
			continue
		}
		if base <= b && b < (base+64) {
			buf |= uint32(b-base) << bits
			bits += 6
			if bits >= 8 {
				data = append(data, byte(buf&0xFF))
				buf >>= 8
				bits -= 8
			}
			continue
		}
		if (b == '#' && bits == 0) || (b == '%' && bits == 2) || (b == '$' && bits == 4) {
			return data, nil
		}
		return nil, nil
	}
}

func decompress(r *bufio.Reader) ([]byte, error) {
	size, err := readNumber(r)
	if err != nil {
		return nil, fmt.Errorf("could not read size: %w", err)
	}
	if size < 0 {
		return nil, fmt.Errorf("negative size")
	}

	const N = 16384
	var table [N]byte
	var vec []byte
	hash := 0
	var buf uint32
	bits := 0

	for i := 0; i < size; i++ {
		if bits == 0 {
			b, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			buf = uint32(b)
			bits = 8
		}

		misprediction := (buf & 1) != 0
		buf >>= 1
		bits--

		var data byte
		if !misprediction {
			data = table[hash]
		} else {
			b, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			buf |= uint32(b) << bits
			data = byte(buf & 0xFF)
			buf >>= 8
			table[hash] = data
		}
		vec = append(vec, data)
		hash = (16*hash + int(data)) % N
	}
	return vec, nil
}

func readNumber(r *bufio.Reader) (int, error) {
	var n int
	bits := 0
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		if b >= 0x80 {
			n |= int(b-0x80) << bits
			bits += 7
			if bits >= 32 {
				return 0, errors.New("invalid bits")
			}
		} else {
			n |= ((int(b) ^ 0x40) - 0x40) << bits
			return n, nil
		}
	}
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
