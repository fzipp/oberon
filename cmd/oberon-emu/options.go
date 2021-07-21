// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"image"

	"github.com/fzipp/oberon/risc"
)

const (
	maxHeight = 2048
	maxWidth  = 2048
)

type options struct {
	http           string
	fullscreen     bool
	zoom           float64
	leds           bool
	mem            int
	size           string
	sizeRect       image.Rectangle
	bootFromSerial bool
	serialIn       string
	serialOut      string
	diskImageFile  string
}

func optionsFromFlags() (*options, error) {
	http := flag.String("http", ":8080", "HTTP service address (e.g., '127.0.0.1:8080' or just ':8080')")
	fullscreen := flag.Bool("fullscreen", false, "Start the emulator in full screen mode")
	zoom := flag.Float64("zoom", 0, "Scale the display in windowed mode by the given factor")
	leds := flag.Bool("leds", false, "Log LED state on stdout")
	mem := flag.Int("mem", 0, "Set memory size in `MEGS`")
	size := flag.String("size", "", "Set framebuffer size to `WIDTHxHEIGHT`")
	bootFromSerial := flag.Bool("boot-from-serial", false, "Boot from serial line (disk image not required)")
	serialIn := flag.String("serial-in", "", "Read serial input from `FILE`")
	serialOut := flag.String("serial-out", "", "Read serial input from `FILE`")

	flag.Parse()

	if flag.NArg() < 1 {
		return nil, errors.New("missing argument")
	}

	sizeRect := image.Rect(0, 0, risc.FramebufferWidth, risc.FramebufferHeight)

	if *size != "" {
		var w, h int
		_, err := fmt.Sscanf(*size, "%dx%d", &w, &h)
		if err != nil {
			return nil, errors.New("invalid size")
		}
		sizeRect = image.Rect(0, 0,
			clamp(w, 32, maxWidth)&^31,
			clamp(h, 32, maxHeight),
		)
	}

	return &options{
		http:           *http,
		fullscreen:     *fullscreen,
		zoom:           *zoom,
		leds:           *leds,
		mem:            *mem,
		size:           *size,
		sizeRect:       sizeRect,
		bootFromSerial: *bootFromSerial,
		serialIn:       *serialIn,
		serialOut:      *serialOut,
		diskImageFile:  flag.Arg(0),
	}, nil
}

func clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
