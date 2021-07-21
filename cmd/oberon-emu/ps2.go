// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package main

import "github.com/fzipp/oberon/cmd/oberon-emu/internal/canvas"

type kInfo struct {
	code byte
	typ  kType
}

type kType int

const (
	kUnknown kType = iota
	kNormal
	kExtended
	kShift
)

// ps2Encode translates a canvas keyboard event into a PS/2 keyboard command
// sequence. See https://wiki.osdev.org/PS/2_Keyboard for a list of commands.
// The 'make' parameter indicates if the key is pressed (true) or released
// (false).
func ps2Encode(e canvas.KeyboardEvent, make bool) []byte {
	var out []byte
	info := keymap[e.Key]
	switch info.typ {
	case kUnknown:
		break
	case kNormal:
		if !make {
			out = append(out, 0xF0)
		}
		out = append(out, info.code)
	case kExtended:
		out = append(out, 0xE0)
		if !make {
			out = append(out, 0xF0)
		}
		out = append(out, info.code)
	case kShift:
		// This assumes Num Lock is always active
		if make {
			// fake shift press
			out = append(out, 0xE0)
			out = append(out, 0x12)
			out = append(out, 0xE0)
			out = append(out, info.code)
			// fake shift release
			out = append(out, 0xE0)
			out = append(out, 0xF0)
			out = append(out, 0x12)
		} else {
			out = append(out, 0xE0)
			out = append(out, 0xF0)
			out = append(out, info.code)
			// fake shift release
			out = append(out, 0xE0)
			out = append(out, 0xF0)
			out = append(out, 0x12)
		}
	}
	return out
}

var keymap = map[string]kInfo{
	"a": {0x1C, kNormal},
	"b": {0x32, kNormal},
	"c": {0x21, kNormal},
	"d": {0x23, kNormal},
	"e": {0x24, kNormal},
	"f": {0x2B, kNormal},
	"g": {0x34, kNormal},
	"h": {0x33, kNormal},
	"i": {0x43, kNormal},
	"j": {0x3B, kNormal},
	"k": {0x42, kNormal},
	"l": {0x4B, kNormal},
	"m": {0x3A, kNormal},
	"n": {0x31, kNormal},
	"o": {0x44, kNormal},
	"p": {0x4D, kNormal},
	"q": {0x15, kNormal},
	"r": {0x2D, kNormal},
	"s": {0x1B, kNormal},
	"t": {0x2C, kNormal},
	"u": {0x3C, kNormal},
	"v": {0x2A, kNormal},
	"w": {0x1D, kNormal},
	"x": {0x22, kNormal},
	"y": {0x35, kNormal},
	"z": {0x1A, kNormal},

	"A": {0x1C, kShift},
	"B": {0x32, kShift},
	"C": {0x21, kShift},
	"D": {0x23, kShift},
	"E": {0x24, kShift},
	"F": {0x2B, kShift},
	"G": {0x34, kShift},
	"H": {0x33, kShift},
	"I": {0x43, kShift},
	"J": {0x3B, kShift},
	"K": {0x42, kShift},
	"L": {0x4B, kShift},
	"M": {0x3A, kShift},
	"N": {0x31, kShift},
	"O": {0x44, kShift},
	"P": {0x4D, kShift},
	"Q": {0x15, kShift},
	"R": {0x2D, kShift},
	"S": {0x1B, kShift},
	"T": {0x2C, kShift},
	"U": {0x3C, kShift},
	"V": {0x2A, kShift},
	"W": {0x1D, kShift},
	"X": {0x22, kShift},
	"Y": {0x35, kShift},
	"Z": {0x1A, kShift},

	"1": {0x16, kNormal},
	"2": {0x1E, kNormal},
	"3": {0x26, kNormal},
	"4": {0x25, kNormal},
	"5": {0x2E, kNormal},
	"6": {0x36, kNormal},
	"7": {0x3D, kNormal},
	"8": {0x3E, kNormal},
	"9": {0x46, kNormal},
	"0": {0x45, kNormal},

	"Enter":     {0x5A, kNormal},
	"Escape":    {0x76, kNormal},
	"Backspace": {0x66, kNormal},
	"Tab":       {0x0D, kNormal},
	" ":         {0x29, kNormal},

	"-": {0x4E, kNormal},
	"=": {0x55, kNormal},
	"[": {0x54, kNormal},
	"]": {0x5B, kNormal},
	`\`: {0x5D, kNormal},
	";": {0x4C, kNormal},
	"'": {0x52, kNormal},
	"`": {0x0E, kNormal},
	",": {0x41, kNormal},
	".": {0x49, kNormal},
	"/": {0x4A, kNormal},

	"<": {0x41, kShift},
	">": {0x49, kShift},
	":": {0x4C, kShift},
	"_": {0x4E, kShift},
	"#": {0x26, kShift},
	"~": {0x0E, kShift},
	"@": {0x1E, kShift},
	"|": {0x5D, kShift},
	"+": {0x55, kShift},
	"?": {0x4A, kShift},
	"(": {0x46, kShift},
	")": {0x45, kShift},
	"&": {0x3D, kShift},
	"%": {0x2E, kShift},
	"$": {0x25, kShift},
	`"`: {0x52, kShift},
	"!": {0x16, kShift},
	"^": {0x36, kShift},
	"*": {0x3E, kShift},
	"{": {0x54, kShift},
	"}": {0x5B, kShift},

	"F1":  {0x05, kNormal},
	"F2":  {0x06, kNormal},
	"F3":  {0x04, kNormal},
	"F4":  {0x0C, kNormal},
	"F5":  {0x03, kNormal},
	"F6":  {0x0B, kNormal},
	"F7":  {0x83, kNormal},
	"F8":  {0x0A, kNormal},
	"F9":  {0x01, kNormal},
	"F10": {0x09, kNormal},
	"F11": {0x78, kNormal},
	"F12": {0x07, kNormal},

	/*
		// Most of the keys below are not used by Oberon

		sdl.SCANCODE_INSERT:   {0x70, kNumLockHack},
		sdl.SCANCODE_HOME:     {0x6C, kNumLockHack},
		sdl.SCANCODE_PAGEUP:   {0x7D, kNumLockHack},
		sdl.SCANCODE_DELETE:   {0x71, kNumLockHack},
		sdl.SCANCODE_END:      {0x69, kNumLockHack},
		sdl.SCANCODE_PAGEDOWN: {0x7A, kNumLockHack},
		sdl.SCANCODE_RIGHT:    {0x74, kNumLockHack},
		sdl.SCANCODE_LEFT:     {0x68, kNumLockHack},
		sdl.SCANCODE_DOWN:     {0x72, kNumLockHack},
		sdl.SCANCODE_UP:       {0x75, kNumLockHack},

		sdl.SCANCODE_NONUSBACKSLASH: {0x61, kNormal},
		sdl.SCANCODE_APPLICATION:    {0x2F, kExtended},

		sdl.SCANCODE_LCTRL:  {0x14, kNormal},
		sdl.SCANCODE_LSHIFT: {0x12, kNormal},
		sdl.SCANCODE_LALT:   {0x11, kNormal},
		sdl.SCANCODE_LGUI:   {0x1F, kExtended},
		sdl.SCANCODE_RCTRL:  {0x14, kExtended},
		sdl.SCANCODE_RSHIFT: {0x59, kNormal},
		sdl.SCANCODE_RALT:   {0x11, kExtended},
		sdl.SCANCODE_RGUI:   {0x27, kExtended},
	*/
}
