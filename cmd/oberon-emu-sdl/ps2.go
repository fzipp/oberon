// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package main

import "github.com/veandco/go-sdl2/sdl"

type kInfo struct {
	code byte
	typ  kType
}

type kType int

const (
	kUnknown kType = iota
	kNormal
	kExtended
	kNumLockHack
	kShiftHack
)

// ps2Encode translates an SDL keyboard scancode into a PS/2 keyboard command
// sequence. See https://wiki.osdev.org/PS/2_Keyboard for a list of commands.
// The 'make' parameter indicates if the key is pressed (true) or released
// (false).
func ps2Encode(sdlScancode sdl.Scancode, make bool) []byte {
	var out []byte
	info := keymap[sdlScancode]
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
	case kNumLockHack:
		// This assumes Num Lock is always active
		if make {
			// fake shift press
			out = append(out, 0xE0)
			out = append(out, 0x12)
			out = append(out, 0xE0)
			out = append(out, info.code)
		} else {
			out = append(out, 0xE0)
			out = append(out, 0xF0)
			out = append(out, info.code)
			// fake shift release
			out = append(out, 0xE0)
			out = append(out, 0xF0)
			out = append(out, 0x12)
		}
	case kShiftHack:
		mod := sdl.GetModState()
		if make {
			// fake shift release
			if mod&sdl.KMOD_LSHIFT > 0 {
				out = append(out, 0xE0)
				out = append(out, 0xF0)
				out = append(out, 0x12)
			}
			if mod&sdl.KMOD_RSHIFT > 0 {
				out = append(out, 0xE0)
				out = append(out, 0xF0)
				out = append(out, 0x59)
			}
			out = append(out, 0xE0)
			out = append(out, info.code)
		} else {
			out = append(out, 0xE0)
			out = append(out, 0xF0)
			out = append(out, info.code)
			// fake shift press
			if mod&sdl.KMOD_RSHIFT > 0 {
				out = append(out, 0xE0)
				out = append(out, 0x59)
			}
			if mod&sdl.KMOD_LSHIFT > 0 {
				out = append(out, 0xE0)
				out = append(out, 0x12)
			}
		}
	}
	return out
}

var keymap = [sdl.NUM_SCANCODES]kInfo{
	sdl.SCANCODE_A: {0x1C, kNormal},
	sdl.SCANCODE_B: {0x32, kNormal},
	sdl.SCANCODE_C: {0x21, kNormal},
	sdl.SCANCODE_D: {0x23, kNormal},
	sdl.SCANCODE_E: {0x24, kNormal},
	sdl.SCANCODE_F: {0x2B, kNormal},
	sdl.SCANCODE_G: {0x34, kNormal},
	sdl.SCANCODE_H: {0x33, kNormal},
	sdl.SCANCODE_I: {0x43, kNormal},
	sdl.SCANCODE_J: {0x3B, kNormal},
	sdl.SCANCODE_K: {0x42, kNormal},
	sdl.SCANCODE_L: {0x4B, kNormal},
	sdl.SCANCODE_M: {0x3A, kNormal},
	sdl.SCANCODE_N: {0x31, kNormal},
	sdl.SCANCODE_O: {0x44, kNormal},
	sdl.SCANCODE_P: {0x4D, kNormal},
	sdl.SCANCODE_Q: {0x15, kNormal},
	sdl.SCANCODE_R: {0x2D, kNormal},
	sdl.SCANCODE_S: {0x1B, kNormal},
	sdl.SCANCODE_T: {0x2C, kNormal},
	sdl.SCANCODE_U: {0x3C, kNormal},
	sdl.SCANCODE_V: {0x2A, kNormal},
	sdl.SCANCODE_W: {0x1D, kNormal},
	sdl.SCANCODE_X: {0x22, kNormal},
	sdl.SCANCODE_Y: {0x35, kNormal},
	sdl.SCANCODE_Z: {0x1A, kNormal},

	sdl.SCANCODE_1: {0x16, kNormal},
	sdl.SCANCODE_2: {0x1E, kNormal},
	sdl.SCANCODE_3: {0x26, kNormal},
	sdl.SCANCODE_4: {0x25, kNormal},
	sdl.SCANCODE_5: {0x2E, kNormal},
	sdl.SCANCODE_6: {0x36, kNormal},
	sdl.SCANCODE_7: {0x3D, kNormal},
	sdl.SCANCODE_8: {0x3E, kNormal},
	sdl.SCANCODE_9: {0x46, kNormal},
	sdl.SCANCODE_0: {0x45, kNormal},

	sdl.SCANCODE_RETURN:    {0x5A, kNormal},
	sdl.SCANCODE_ESCAPE:    {0x76, kNormal},
	sdl.SCANCODE_BACKSPACE: {0x66, kNormal},
	sdl.SCANCODE_TAB:       {0x0D, kNormal},
	sdl.SCANCODE_SPACE:     {0x29, kNormal},

	sdl.SCANCODE_MINUS:        {0x4E, kNormal},
	sdl.SCANCODE_EQUALS:       {0x55, kNormal},
	sdl.SCANCODE_LEFTBRACKET:  {0x54, kNormal},
	sdl.SCANCODE_RIGHTBRACKET: {0x5B, kNormal},
	sdl.SCANCODE_BACKSLASH:    {0x5D, kNormal},
	sdl.SCANCODE_NONUSHASH:    {0x5D, kNormal}, // same key as BACKSLASH

	sdl.SCANCODE_SEMICOLON:  {0x4C, kNormal},
	sdl.SCANCODE_APOSTROPHE: {0x52, kNormal},
	sdl.SCANCODE_GRAVE:      {0x0E, kNormal},
	sdl.SCANCODE_COMMA:      {0x41, kNormal},
	sdl.SCANCODE_PERIOD:     {0x49, kNormal},
	sdl.SCANCODE_SLASH:      {0x4A, kNormal},

	sdl.SCANCODE_F1:  {0x05, kNormal},
	sdl.SCANCODE_F2:  {0x06, kNormal},
	sdl.SCANCODE_F3:  {0x04, kNormal},
	sdl.SCANCODE_F4:  {0x0C, kNormal},
	sdl.SCANCODE_F5:  {0x03, kNormal},
	sdl.SCANCODE_F6:  {0x0B, kNormal},
	sdl.SCANCODE_F7:  {0x83, kNormal},
	sdl.SCANCODE_F8:  {0x0A, kNormal},
	sdl.SCANCODE_F9:  {0x01, kNormal},
	sdl.SCANCODE_F10: {0x09, kNormal},
	sdl.SCANCODE_F11: {0x78, kNormal},
	sdl.SCANCODE_F12: {0x07, kNormal},

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

	sdl.SCANCODE_KP_DIVIDE:   {0x4A, kShiftHack},
	sdl.SCANCODE_KP_MULTIPLY: {0x7C, kNormal},
	sdl.SCANCODE_KP_MINUS:    {0x7B, kNormal},
	sdl.SCANCODE_KP_PLUS:     {0x79, kNormal},
	sdl.SCANCODE_KP_ENTER:    {0x5A, kExtended},
	sdl.SCANCODE_KP_1:        {0x69, kNormal},
	sdl.SCANCODE_KP_2:        {0x72, kNormal},
	sdl.SCANCODE_KP_3:        {0x7A, kNormal},
	sdl.SCANCODE_KP_4:        {0x6B, kNormal},
	sdl.SCANCODE_KP_5:        {0x73, kNormal},
	sdl.SCANCODE_KP_6:        {0x74, kNormal},
	sdl.SCANCODE_KP_7:        {0x6C, kNormal},
	sdl.SCANCODE_KP_8:        {0x75, kNormal},
	sdl.SCANCODE_KP_9:        {0x7D, kNormal},
	sdl.SCANCODE_KP_0:        {0x70, kNormal},
	sdl.SCANCODE_KP_PERIOD:   {0x71, kNormal},

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
}
