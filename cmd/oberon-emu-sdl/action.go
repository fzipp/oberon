// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package main

import "github.com/veandco/go-sdl2/sdl"

type action int

const (
	actionOberonInput action = iota
	actionQuit
	actionReset
	actionToggleFullscreen
	actionFakeMouse1
	actionFakeMouse2
	actionFakeMouse3
)

type keyMapping struct {
	state      uint8
	sym        sdl.Keycode
	mod1, mod2 sdl.Keymod
	action     action
}

var keyMap = []keyMapping{
	{sdl.PRESSED, sdl.K_F4, sdl.KMOD_ALT, 0, actionQuit},
	{sdl.PRESSED, sdl.K_F12, 0, 0, actionReset},
	{sdl.PRESSED, sdl.K_DELETE, sdl.KMOD_CTRL, sdl.KMOD_SHIFT, actionReset},
	{sdl.PRESSED, sdl.K_F11, 0, 0, actionToggleFullscreen},
	{sdl.PRESSED, sdl.K_RETURN, sdl.KMOD_ALT, 0, actionToggleFullscreen},
	{sdl.PRESSED, sdl.K_f, sdl.KMOD_GUI, sdl.KMOD_CTRL, actionToggleFullscreen}, // Mac fullscreen shortcut
	{sdl.PRESSED, sdl.K_LALT, 0, 0, actionFakeMouse2},
	{sdl.RELEASED, sdl.K_LALT, 0, 0, actionFakeMouse2},

	{sdl.PRESSED, sdl.K_LCTRL, 0, 0, actionFakeMouse1},
	{sdl.RELEASED, sdl.K_LCTRL, 0, 0, actionFakeMouse1},
	{sdl.PRESSED, sdl.K_LGUI, 0, 0, actionFakeMouse3},
	{sdl.RELEASED, sdl.K_LGUI, 0, 0, actionFakeMouse3},
}

func mapKeyboardEvent(event *sdl.KeyboardEvent) action {
	for i := range keyMap {
		if (event.State == keyMap[i].state) &&
			(event.Keysym.Sym == keyMap[i].sym) &&
			((keyMap[i].mod1 == 0) || (sdl.Keymod(event.Keysym.Mod)&keyMap[i].mod1) > 0) &&
			((keyMap[i].mod2 == 0) || (sdl.Keymod(event.Keysym.Mod)&keyMap[i].mod2) > 0) {

			return keyMap[i].action
		}
	}
	return actionOberonInput
}
