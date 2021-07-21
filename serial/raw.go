// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package serial

import (
	"fmt"
	"io"
	"os"
)

type Raw struct {
	r io.ReadCloser
	w io.WriteCloser
}

func Open(filenameIn, filenameOut string) (*Raw, error) {
	raw := &Raw{}

	r, err := os.Open(filenameIn)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial input file: %w", err)
	}
	raw.r = r

	w, err := os.OpenFile(filenameOut, os.O_RDWR, 0o666)
	if err != nil {
		_ = r.Close()
		return nil, fmt.Errorf("failed to open serial output file: %w", err)
	}
	raw.w = w

	return raw, nil
}

func (r *Raw) ReadStatus() uint32 {
	// TODO
	return 0
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
	err = r.r.Close()
	if err != nil {
		return err
	}
	err = r.w.Close()
	return err
}
