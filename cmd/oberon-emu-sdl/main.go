// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

// Command oberon-emu-sdl is an emulator for the Project Oberon RISC machine.
// It uses SDL to render the screen.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"log"
	"math"
	"os"

	"github.com/fzipp/oberon/risc"
	"github.com/fzipp/oberon/serial"
	"github.com/fzipp/oberon/spi"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	cpuHz = 25000000
	fps   = 60
)

const (
	colorBlack = 0x657b83
	colorWhite = 0xfdf6e3
)

func main() {
	opt, err := optionsFromFlags()
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	r := risc.New()
	r.SetSerial(&serial.PCLink{})
	r.SetClipboard(&SDLClipboard{})

	if opt.leds {
		r.SetLEDs(&ConsoleLEDs{})
	}

	if opt.bootFromSerial {
		r.SetSwitches(1)
	}

	if opt.mem > 0 || opt.size != "" {
		r.ConfigureMemory(opt.mem, opt.sizeRect.Dx(), opt.sizeRect.Dy())
	}

	disk, err := spi.NewDisk(opt.diskImageFile)
	check(err)
	r.SetSPI(1, disk)

	if opt.serialIn != "" || opt.serialOut != "" {
		raw, err := serial.Open(opt.serialIn, opt.serialOut)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "can't open serial I/O: %s", err)
			return
		}
		r.SetSerial(raw)
	}

	riscRect := sdl.Rect{
		W: int32(opt.sizeRect.Dx()),
		H: int32(opt.sizeRect.Dy()),
	}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()
	sdl.EnableScreenSaver()
	_, err = sdl.ShowCursor(0)
	check(err)
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "best")

	windowFlags := sdl.WINDOW_HIDDEN
	display := 0
	if opt.fullscreen {
		windowFlags |= sdl.WINDOW_FULLSCREEN_DESKTOP
		display, err = bestDisplay(riscRect)
		check(err)
	}
	if opt.zoom == 0 {
		bounds, err := sdl.GetDisplayBounds(display)
		check(err)
		if bounds.H >= riscRect.H*2 && bounds.W >= riscRect.W*2 {
			opt.zoom = 2
		} else {
			opt.zoom = 1
		}
	}
	window, err := sdl.CreateWindow("Project Oberon",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		int32(float64(riscRect.W)*(opt.zoom)),
		int32(float64(riscRect.H)*(opt.zoom)),
		uint32(windowFlags))
	check(err)
	renderer, err := sdl.CreateRenderer(window, -1, 0)
	check(err)
	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_ARGB8888,
		sdl.TEXTUREACCESS_STREAMING,
		riscRect.W,
		riscRect.H,
	)
	check(err)

	fb := r.Framebuffer()
	displayRect, displayScale := scaleDisplay(window, riscRect)
	err = updateTexture(fb, r.GetFramebufferDamageAndReset(), texture, riscRect)
	check(err)
	window.Show()
	err = renderer.Clear()
	check(err)
	err = renderer.Copy(texture, &riscRect, &displayRect)
	check(err)
	renderer.Present()

	done := false
	mouseWasOffscreen := false
	for !done {
		frameStart := sdl.GetTicks()

		for {
			event := sdl.PollEvent()
			if event == nil {
				break
			}
			switch event.GetType() {
			case sdl.QUIT:
				done = true

			case sdl.WINDOWEVENT:
				ev := event.(*sdl.WindowEvent)
				if ev.Event == sdl.WINDOWEVENT_RESIZED {
					displayRect, displayScale = scaleDisplay(window, riscRect)
				}

			case sdl.MOUSEMOTION:
				ev := event.(*sdl.MouseMotionEvent)
				scaledX := int(math.Round(float64(ev.X-displayRect.X) / displayScale))
				scaledY := int(math.Round(float64(ev.Y-displayRect.Y) / displayScale))
				x := clamp(scaledX, 0, int(riscRect.W)-1)
				y := clamp(scaledY, 0, int(riscRect.H)-1)
				mouseIsOffscreen := x != scaledX || y != scaledY
				if mouseIsOffscreen != mouseWasOffscreen {
					var toggle int
					if mouseIsOffscreen {
						toggle = sdl.ENABLE
					} else {
						toggle = sdl.DISABLE
					}
					_, err = sdl.ShowCursor(toggle)
					check(err)
					mouseWasOffscreen = mouseIsOffscreen
				}
				r.MouseMoved(x, int(riscRect.H)-y-1)

			case sdl.MOUSEBUTTONDOWN, sdl.MOUSEBUTTONUP:
				ev := event.(*sdl.MouseButtonEvent)
				down := ev.State == sdl.PRESSED
				r.MouseButton(int(ev.Button), down)

			case sdl.KEYDOWN, sdl.KEYUP:
				ev := event.(*sdl.KeyboardEvent)
				down := ev.State == sdl.PRESSED
				switch mapKeyboardEvent(ev) {
				case actionReset:
					r.Reset()
				case actionToggleFullscreen:
					opt.fullscreen = !opt.fullscreen
					if opt.fullscreen {
						err = window.SetFullscreen(sdl.WINDOW_FULLSCREEN_DESKTOP)
					} else {
						err = window.SetFullscreen(0)
					}
					check(err)
				case actionQuit:
					_, err = sdl.PushEvent(&sdl.QuitEvent{
						Type:      sdl.QUIT,
						Timestamp: sdl.GetTicks(),
					})
					check(err)
				case actionFakeMouse1:
					r.MouseButton(1, down)
				case actionFakeMouse2:
					r.MouseButton(2, down)
				case actionFakeMouse3:
					r.MouseButton(3, down)
				case actionOberonInput:
					r.KeyboardInput(ps2Encode(ev.Keysym.Scancode, down))
				}
			}
		}

		r.SetTime(frameStart)
		err = r.Run(cpuHz / fps)
		if err != nil {
			if riscErr, ok := err.(*risc.Error); ok {
				_, _ = fmt.Fprintf(os.Stderr, "%s (PC=0x%08X)\n", riscErr, riscErr.PC)
			} else {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}

		err = updateTexture(fb, r.GetFramebufferDamageAndReset(), texture, riscRect)
		check(err)
		err = renderer.Clear()
		check(err)
		err = renderer.Copy(texture, &riscRect, &displayRect)
		check(err)
		renderer.Present()

		frameEnd := sdl.GetTicks()
		delay := int(frameStart) + 1000/fps - int(frameEnd)
		if delay > 0 {
			sdl.Delay(uint32(delay))
		}
	}
}

func scaleDisplay(window *sdl.Window, riscRect sdl.Rect) (sdl.Rect, float64) {
	winW, winH := window.GetSize()
	oberonAspect := float64(riscRect.W) / float64(riscRect.H)
	windowAspect := float64(winW) / float64(winH)

	var scale float64
	if oberonAspect > windowAspect {
		scale = float64(winW) / float64(riscRect.W)
	} else {
		scale = float64(winH) / float64(riscRect.H)
	}

	w := int32(math.Ceil(float64(riscRect.W) * scale))
	h := int32(math.Ceil(float64(riscRect.H) * scale))
	return sdl.Rect{
		W: w,
		H: h,
		X: (winW - w) / 2,
		Y: (winH - h) / 2,
	}, scale
}

// Only used in update_texture(), but some systems complain if you
// allocate three megabyte on the stack.
var pixelBuf [maxWidth * maxHeight * 4]byte

func updateTexture(fb *risc.Framebuffer, damage image.Rectangle, texture *sdl.Texture, riscRect sdl.Rect) error {
	if damage.Min.Y > damage.Max.Y {
		return nil
	}

	var outIdx uint32

	for line := damage.Max.Y; line >= damage.Min.Y; line-- {
		lineStart := line * (int(riscRect.W) / 32)
		for col := damage.Min.X; col <= damage.Max.X; col++ {
			pixels := fb.Pix[lineStart+col]
			for b := 0; b < 32; b++ {
				var color uint32
				if pixels&1 > 0 {
					color = colorWhite
				} else {
					color = colorBlack
				}
				binary.LittleEndian.PutUint32(pixelBuf[outIdx*4:], color)
				pixels >>= 1
				outIdx++
			}
		}
	}

	rect := sdl.Rect{
		X: int32(damage.Min.X) * 32,
		Y: riscRect.H - int32(damage.Max.Y) - 1,
		W: int32((damage.Max.X - damage.Min.X + 1) * 32),
		H: int32(damage.Max.Y - damage.Min.Y + 1),
	}
	return texture.Update(&rect, pixelBuf[:], int(rect.W)*4)
}

func bestDisplay(rect sdl.Rect) (int, error) {
	best := 0
	displayCnt, err := sdl.GetNumVideoDisplays()
	if err != nil {
		return best, err
	}
	for i := 0; i < displayCnt; i++ {
		bounds, err := sdl.GetDisplayBounds(i)
		if err != nil {
			return best, err
		}
		if bounds.H == rect.H && bounds.W >= rect.W {
			best = i
			if bounds.W == rect.W {
				break // exact match
			}
		}
	}
	return best, nil
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
