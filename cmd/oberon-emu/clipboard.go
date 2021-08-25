// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package main

import (
	"math"
	"strings"

	"github.com/fzipp/oberon/cmd/oberon-emu/internal/canvas"
)

type Clipboard struct {
	ctx     *canvas.Context
	state   clipboardState
	text    string
	data    string
	dataLen uint32
}

type clipboardState int

const (
	clipIdle clipboardState = iota
	clipGet
	clipPut
)

func (c *Clipboard) reset() {
	c.state = clipIdle
	c.data = ""
	c.dataLen = 0
}

func (c *Clipboard) setText(data string) {
	c.text = data
}

func (c *Clipboard) ReadControl() uint32 {
	c.reset()
	c.data = c.text
	c.data = strings.ReplaceAll(c.data, "\r\n", "\r")
	c.data = strings.ReplaceAll(c.data, "\n", "\r")
	if len(c.data) > math.MaxUint32 {
		c.reset()
		return 0
	}
	c.dataLen = uint32(len(c.data))
	c.state = clipGet
	return c.dataLen
}

func (c *Clipboard) WriteControl(length uint32) {
	c.reset()
	c.state = clipPut
	c.dataLen = length
}

func (c *Clipboard) ReadData() uint32 {
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

func (c *Clipboard) WriteData(value uint32) {
	if c.state != clipPut {
		return
	}
	ch := byte(value)
	if ch == '\r' {
		ch = '\n'
	}
	c.data += string(rune(ch))
	if len(c.data) == int(c.dataLen) {
		c.ctx.ClipboardWriteText(c.data)
		c.reset()
	}
}
