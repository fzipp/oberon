// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type SDLClipboard struct {
	state   clipboardState
	data    string
	dataLen uint32
}

type clipboardState int

const (
	clipIdle clipboardState = iota
	clipGet
	clipPut
)

func (c *SDLClipboard) reset() {
	c.state = clipIdle
	c.data = ""
	c.dataLen = 0
}

func (c *SDLClipboard) ReadControl() uint32 {
	c.reset()
	data, err := sdl.GetClipboardText()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "can't get clipboard text: %s", err)
	}
	c.data = strings.ReplaceAll(data, "\r\n", "\r")
	c.data = strings.ReplaceAll(c.data, "\n", "\r")
	if len(c.data) > math.MaxUint32 {
		c.reset()
		return 0
	}
	c.dataLen = uint32(len(c.data))
	c.state = clipGet
	return c.dataLen
}

func (c *SDLClipboard) WriteControl(length uint32) {
	c.reset()
	c.state = clipPut
	c.dataLen = length
}

func (c *SDLClipboard) ReadData() uint32 {
	if c.state != clipGet {
		return 0
	}
	if len(c.data) == 0 {
		c.reset()
		return 0
	}
	result := uint32(c.data[0])
	c.data = c.data[1:]
	return result
}

func (c *SDLClipboard) WriteData(value uint32) {
	if c.state != clipPut {
		return
	}
	ch := byte(value)
	if ch == '\r' {
		ch = '\n'
	}
	c.data += string(rune(ch))
	if len(c.data) == int(c.dataLen) {
		err := sdl.SetClipboardText(c.data)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "can't set clipboard text: %s", err)
		}
		c.reset()
	}
}
