// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

// Command oberon-emu is an emulator for the Project Oberon RISC machine.
// It starts a WebSocket server to render the screen in a web browser on an
// HTML canvas.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fzipp/oberon/risc"
	"github.com/fzipp/oberon/serial"
	"github.com/fzipp/oberon/spi"

	"github.com/fzipp/oberon/cmd/oberon-emu/internal/canvas"
)

const (
	cpuHz = 25000000
	fps   = 60
)

func main() {
	opt, err := optionsFromFlags()
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Visit " + httpLink(opt.http) + " in a web browser")
	err = canvas.ListenAndServe(opt.http, func(ctx *canvas.Context) {
		run(ctx, opt)
	}, opt.sizeRect)
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx *canvas.Context, opt *options) {
	r := risc.New()
	r.SetSerial(&serial.PCLink{})
	clipboard := &Clipboard{ctx: ctx}
	r.SetClipboard(clipboard)

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
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "can't use disk image: %s", err)
		return
	}
	r.SetSPI(1, disk)

	if opt.serialIn != "" || opt.serialOut != "" {
		raw, err := serial.Open(opt.serialIn, opt.serialOut)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "can't open serial I/O: %s", err)
			return
		}
		r.SetSerial(raw)
	}

	fb := r.Framebuffer()

	riscStart := getTicks()
	for {
		frameStart := getTicks()
		select {
		case event := <-ctx.Events():
			if _, ok := event.(canvas.CloseEvent); ok {
				return
			}
			handleEvent(event, r, ctx, clipboard)
		default:
			r.SetTime(uint32(frameStart - riscStart))
			err := r.Run(cpuHz / fps)
			if err != nil {
				if riscErr, ok := err.(*risc.Error); ok {
					_, _ = fmt.Fprintf(os.Stderr, "%s (PC=0x%08X)\n", riscErr, riscErr.PC)
				} else {
					_, _ = fmt.Fprintln(os.Stderr, err)
				}
			}

			ctx.UpdateDisplay(fb, r.GetFramebufferDamageAndReset())

			frameEnd := getTicks()
			delay := frameStart + 1000/fps - frameEnd
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

func handleEvent(e canvas.Event, r *risc.RISC, ctx *canvas.Context, clipboard *Clipboard) {
	switch ev := e.(type) {
	case canvas.MouseMoveEvent:
		r.MouseMoved(ev.X, ctx.CanvasHeight()-ev.Y)
	case canvas.MouseDownEvent:
		if ev.AltKey() {
			r.MouseButton(2, true)
			break
		}
		if ev.MetaKey() {
			r.MouseButton(3, true)
			break
		}
		if ev.Buttons&canvas.ButtonPrimary > 0 {
			r.MouseButton(1, true)
		}
		if ev.Buttons&canvas.ButtonAuxiliary > 0 {
			r.MouseButton(2, true)
		}
		if ev.Buttons&canvas.ButtonSecondary > 0 {
			r.MouseButton(3, true)
		}
	case canvas.MouseUpEvent:
		r.MouseButton(1, false)
		r.MouseButton(2, false)
		r.MouseButton(3, false)
	case canvas.KeyDownEvent:
		if ev.Key == "Control" {
			r.MouseButton(1, true)
			return
		}
		if ev.Key == "Alt" {
			r.MouseButton(2, true)
			return
		}
		if ev.Key == "Meta" {
			r.MouseButton(3, true)
			return
		}
		r.KeyboardInput(ps2Encode(ev.KeyboardEvent, true))
	case canvas.KeyUpEvent:
		if ev.Key == "Control" {
			r.MouseButton(1, false)
			return
		}
		if ev.Key == "Alt" {
			r.MouseButton(2, false)
			return
		}
		if ev.Key == "Meta" {
			r.MouseButton(3, false)
			return
		}
		r.KeyboardInput(ps2Encode(ev.KeyboardEvent, false))
	case canvas.ClipboardChangeEvent:
		clipboard.setText(ev.Data)
	}
}

func getTicks() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func httpLink(addr string) string {
	if addr[0] == ':' {
		addr = "localhost" + addr
	}
	return "http://" + addr
}
