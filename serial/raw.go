// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package serial

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Raw struct {
	in *os.File
	r  *bufio.Reader
	w  io.WriteCloser
}

func Open(filenameIn, filenameOut string) (*Raw, error) {
	if filenameIn == "" {
		filenameIn = os.DevNull
	}
	if filenameOut == "" {
		filenameOut = os.DevNull
	}

	raw := &Raw{}

	in, err := os.Open(filenameIn)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial input file: %w", err)
	}
	raw.in = in
	raw.r = bufio.NewReader(in)

	out, err := os.OpenFile(filenameOut, os.O_RDWR, 0o666)
	if err != nil {
		_ = in.Close()
		return nil, fmt.Errorf("failed to open serial output file: %w", err)
	}
	raw.w = out

	return raw, nil
}

func (r *Raw) ReadStatus() uint32 {
	_, err := r.r.ReadByte()
	if err != nil {
		return 2
	}
	err = r.r.UnreadByte()
	if err != nil {
		return 2
	}
	return 3
}

func (r *Raw) ReadData() uint32 {
	var buf [1]byte
	_, err := r.r.Read(buf[:])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "can't read serial data: %s\n", err)
	}
	return uint32(buf[0])
}

func (r *Raw) WriteData(value uint32) {
	buf := [1]byte{byte(value)}
	_, err := r.w.Write(buf[:])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "can't write serial data: %s\n", err)
	}
}

func (r *Raw) Close() error {
	var err error
	err = r.in.Close()
	if err != nil {
		return err
	}
	err = r.w.Close()
	return err
}
