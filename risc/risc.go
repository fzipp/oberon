// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package risc

import "image"

const (
	FramebufferWidth  = 1024
	FramebufferHeight = 768
)

// Our memory layout is slightly different from the FPGA implementation:
// The FPGA uses a 20-bit address bus and thus ignores the top 12 bits,
// while we use all 32 bits. This allows us to have more than 1 megabyte
// of RAM.
//
// In the default configuration, the emulator is compatible with the
// FPGA system. But If the user requests more memory, we move the
// framebuffer to make room for a larger Oberon heap. This requires a
// custom Display.Mod.

const (
	defaultMemSize      = 0x00100000 // 1 MiB
	defaultDisplayStart = 0x000E7F00
)

const (
	romStart = 0xFFFFF800
	romWords = 512
	ioStart  = 0xFFFFFFC0
)

type RISC struct {
	PC uint32     // Program counter
	R  [16]uint32 // General registers R0..R15
	H  uint32     // Auxiliary register for high 32 bits of multiplication or remainder of division
	Z  bool       // Zero flag
	N  bool       // Negative flag
	C  bool       // Carry flag
	V  bool       // Overflow flag

	displayStart uint32

	progress           uint32
	millisecondCounter uint32
	mouse              uint32
	keyBuf             []byte
	switches           uint32

	leds        LED
	serial      Serial
	spiSelected uint32
	spi         [4]SPI
	clipboard   Clipboard

	framebuffer Framebuffer
	damage      image.Rectangle

	Mem []uint32 // Memory
	rom [romWords]uint32
}

// RISC instructions set
const (
	// Moving and shifting
	opMOV = iota
	opLSL
	opASR
	opROR
	// Logical operations
	opAND
	opANN
	opIOR
	opXOR
	// Integer arithmetic
	opADD
	opSUB
	opMUL
	opDIV
	// Floating-point arithmetic
	opFAD
	opFSB
	opFML
	opFDV
)

func New() *RISC {
	screenWidth := FramebufferWidth
	screenHeight := FramebufferHeight
	r := &RISC{
		displayStart: defaultDisplayStart,
	}
	columns := screenWidth / 32
	r.damage = image.Rect(0, 0, columns, screenHeight)
	r.Mem = make([]uint32, defaultMemSize/4)
	r.framebuffer = Framebuffer{
		Rect: image.Rect(0, 0, screenWidth, screenHeight),
		Pix:  r.Mem[r.displayStart/4:],
	}
	r.rom = bootloader
	r.Reset()
	return r
}

func (r *RISC) ConfigureMemory(megabytesRAM, screenWidth, screenHeight int) {
	megabytesRAM = clamp(megabytesRAM, 1, 32)

	r.displayStart = uint32(megabytesRAM << 20)
	columns := screenWidth / 32
	r.damage = image.Rect(0, 0, columns, screenHeight)

	memSize := r.displayStart + uint32((screenWidth*screenHeight)/8)
	r.Mem = make([]uint32, memSize/4)
	r.framebuffer = Framebuffer{
		Rect: image.Rect(0, 0, screenWidth, screenHeight),
		Pix:  r.Mem[r.displayStart/4:],
	}

	// Patch the new constants in the bootloader.
	memLim := r.displayStart - 16
	r.rom[372] = 0x61000000 + (memLim >> 16)
	r.rom[373] = 0x41160000 + (memLim & 0x0000FFFF)
	stackOrg := r.displayStart / 2
	r.rom[376] = 0x61000000 + (stackOrg >> 16)

	// Inform the display driver of the framebuffer layout.
	// This isn't a very pretty mechanism, but this way our disk images
	// should still boot on the standard FPGA system.
	r.Mem[defaultDisplayStart/4] = 0x53697A67
	r.Mem[defaultDisplayStart/4+1] = uint32(screenWidth)
	r.Mem[defaultDisplayStart/4+2] = uint32(screenHeight)
	r.Mem[defaultDisplayStart/4+3] = r.displayStart

	r.Reset()
}

func (r *RISC) SetLEDs(l LED) {
	r.leds = l
}

func (r *RISC) SetSerial(s Serial) {
	r.serial = s
}

func (r *RISC) SetSPI(index int, spi SPI) {
	if index == 1 || index == 2 {
		r.spi[index] = spi
	}
}

func (r *RISC) SetClipboard(c Clipboard) {
	r.clipboard = c
}

func (r *RISC) SetSwitches(s int) {
	r.switches = uint32(s)
}

func (r *RISC) Reset() {
	r.PC = romStart / 4
}

func (r *RISC) Run(cycles int) error {
	r.progress = 20
	// The progress value is used to detect that the RISC cpu is busy
	// waiting on the millisecond counter or on the keyboard ready
	// bit. In that case it's better to just pause emulation until the
	// next frame.
	for i := 0; i < cycles && r.progress > 0; i++ {
		err := r.singleStep()
		if err != nil {
			r.Reset()
			return err
		}
	}
	return nil
}

func (r *RISC) singleStep() error {
	var IR uint32 // Instruction register
	if r.PC < uint32(len(r.Mem)) {
		IR = r.Mem[r.PC]
	} else if r.PC >= romStart/4 && r.PC < romStart/4+romWords {
		IR = r.rom[r.PC-romStart/4]
	} else {
		return &Error{PC: r.PC, message: "branched into the void"}
	}
	r.PC++

	const (
		pBit = 0x80000000
		qBit = 0x40000000
		uBit = 0x20000000
		vBit = 0x10000000
	)

	if (IR & pBit) == 0 {
		// Register instructions (formats F0 and F1)

		a := (IR & 0x0F000000) >> 24
		b := (IR & 0x00F00000) >> 20
		op := (IR & 0x000F0000) >> 16

		var n uint32
		if (IR & qBit) == 0 {
			// F0
			c := IR & 0x0000000F
			Rc := r.R[c]
			n = Rc
		} else {
			// F1
			im := IR & 0x0000FFFF
			if (IR & vBit) == 0 {
				// 0-extend n
				n = im
			} else {
				// 1-extend n
				n = im | 0xFFFF0000
			}
		}

		var Ra uint32
		Rb := r.R[b]

		switch op {
		case opMOV:
			if (IR & uBit) == 0 {
				Ra = n
			} else {
				// Special features
				if (IR & qBit) == 0 {
					// F0
					if (IR & vBit) == 0 {
						Ra = r.H
					} else {
						// From RISC5.v:
						// {N, Z, C, OV, 20'b0, 8'h53}
						NZCV := (b2i(r.N) * 0b1000) |
							(b2i(r.Z) * 0b0100) |
							(b2i(r.C) * 0b0010) |
							(b2i(r.V) * 0b0001)
						Ra = (NZCV << 28) | 0x53
					}
				} else {
					// F1
					Ra = n << 16
				}
			}
		case opLSL:
			// shift left by n bits
			Ra = Rb << (n & 31)
		case opASR:
			// shift right by n bits with sign extension
			Ra = uint32(int32(Rb) >> (n & 31))
		case opROR:
			// rotate right by n bits
			Ra = (Rb >> (n & 31)) | (Rb << (-n & 31))
		case opAND:
			Ra = Rb & n
		case opANN:
			Ra = Rb &^ n
		case opIOR:
			Ra = Rb | n
		case opXOR:
			Ra = Rb ^ n
		case opADD:
			Ra = Rb + n
			if (IR&uBit) != 0 && r.C {
				// ADD' (add also carry C)
				Ra++
			}
			r.C = Ra < Rb
			r.V = (((Ra ^ n) & (Ra ^ Rb)) >> 31) > 0
		case opSUB:
			Ra = Rb - n
			if (IR&uBit) != 0 && r.C {
				// SUB' (subtract also carry C)
				Ra--
			}
			r.C = Ra > Rb
			r.V = (((Rb ^ n) & (Ra ^ Rb)) >> 31) > 0
		case opMUL:
			var tmp uint64
			if (IR & uBit) == 0 {
				tmp = uint64(int64(int32(Rb)) * int64(int32(n)))
			} else {
				// MUL' (unsigned multiplication)
				tmp = uint64(Rb) * uint64(n)
			}
			Ra = uint32(tmp)
			r.H = uint32(tmp >> 32)
		case opDIV:
			if int32(n) > 0 {
				if (IR & uBit) == 0 {
					Ra = uint32(int32(Rb) / int32(n))
					r.H = uint32(int32(Rb) % int32(n))
					if int32(r.H) < 0 {
						Ra--
						r.H += n
					}
				} else {
					Ra = Rb / n
					r.H = Rb % n
				}
			} else {
				q := makeIdiv(Rb, n, IR&uBit > 0)
				Ra = q.quot
				r.H = q.rem
			}
		case opFAD:
			Ra = fpAdd(Rb, n, IR&uBit > 0, IR&vBit > 0)
		case opFSB:
			Ra = fpAdd(Rb, n^0x80000000, IR&uBit > 0, IR&vBit > 0)
		case opFML:
			Ra = fpMul(Rb, n)
		case opFDV:
			Ra = fpDiv(Rb, n)
		default:
			panic("unreachable")
		}
		r.setRegister(a, Ra)
	} else if (IR & qBit) == 0 {
		// Memory instructions (format F2)

		a := (IR & 0x0F000000) >> 24
		b := (IR & 0x00F00000) >> 20
		im := int32(IR & 0x000FFFFF)
		off := uint32((im ^ 0x00080000) - 0x00080000) // sign-extend

		address := r.R[b] + off
		if (IR & uBit) == 0 {
			// LD (load)
			var Ra uint32
			if (IR & vBit) == 0 {
				// word access
				Ra = r.loadWord(address)
			} else {
				// single byte access
				Ra = uint32(r.loadByte(address))
			}
			r.setRegister(a, Ra)
		} else {
			// ST (store)
			Ra := r.R[a]
			if (IR & vBit) == 0 {
				// word access
				r.storeWord(address, Ra)
			} else {
				// single byte access
				r.storeByte(address, byte(Ra))
			}
		}
	} else {
		// Branch instructions (format F3)
		// TODO: interrupts?

		t := ((IR >> 27) & 1) > 0
		switch (IR >> 24) & 0b0111 {
		case 0b0000: // MI: negative (minus)
			t = t != r.N
		case 0b0001: // EQ: equal (zero)
			t = t != r.Z
		case 0b0010: // CS: carry set (lower)
			t = t != r.C
		case 0b0011: // VS: overflow set
			t = t != r.V
		case 0b0100: // LS: less or same
			// Note: RISC-Arch.pdf says ~C|Z (as of 2021-03-28),
			// but RISC5.v says (C|Z). The latter is the correct one.
			t = t != (r.C || r.Z)
		case 0b0101: // LT: less than
			t = t != (r.N != r.V)
		case 0b0110: // LE: less or equal
			t = t != ((r.N != r.V) || r.Z)
		case 0b0111: // always
			t = t != true
		default:
			panic("unreachable")
		}
		if t {
			if (IR & vBit) != 0 {
				const LNK = 15 // R15 is the link register
				r.setRegister(LNK, r.PC*4)
			}
			if (IR & uBit) == 0 {
				c := IR & 0x0000000F
				r.PC = r.R[c] / 4
			} else {
				off := int32(IR & 0x00FFFFFF)
				off = (off ^ 0x00800000) - 0x00800000 // sign-extend
				r.PC = r.PC + uint32(off)
			}
		}
	}
	return nil
}

func (r *RISC) setRegister(reg uint32, value uint32) {
	r.R[reg] = value
	r.Z = value == 0
	r.N = int32(value) < 0
}

func (r *RISC) loadWord(address uint32) uint32 {
	if address < uint32(r.memSize()) {
		return r.Mem[address/4]
	}
	return r.loadIO(address)
}

func (r *RISC) loadByte(address uint32) byte {
	w := r.loadWord(address)
	return byte(w >> (address % 4 * 8))
}

func (r *RISC) updateDamage(word int) {
	columns := r.framebuffer.Rect.Max.X / 32
	row := word / columns
	if row >= r.framebuffer.Rect.Max.Y {
		return
	}
	col := word % columns
	if col < r.damage.Min.X {
		r.damage.Min.X = col
	}
	if col > r.damage.Max.X {
		r.damage.Max.X = col
	}
	if row < r.damage.Min.Y {
		r.damage.Min.Y = row
	}
	if row > r.damage.Max.Y {
		r.damage.Max.Y = row
	}
}

func (r *RISC) storeWord(address, value uint32) {
	if address < r.displayStart {
		r.Mem[address/4] = value
	} else if address < uint32(r.memSize()) {
		r.Mem[address/4] = value
		r.updateDamage(int(address/4 - r.displayStart/4))
	} else {
		r.storeIO(address, value)
	}
}

func (r *RISC) storeByte(address uint32, value byte) {
	if address < uint32(r.memSize()) {
		w := r.loadWord(address)
		shift := (address & 3) * 8
		w &= ^(0xFF << shift)
		w |= uint32(value) << shift
		r.storeWord(address, w)
	} else {
		r.storeIO(address, uint32(value))
	}
}

func (r *RISC) loadIO(address uint32) uint32 {
	switch address - ioStart {
	case 0:
		// Millisecond counter
		r.progress--
		return r.millisecondCounter
	case 4:
		// Switches
		return r.switches
	case 8:
		// RS-232 data
		if r.serial == nil {
			return 0
		}
		return r.serial.ReadData()
	case 12:
		// RS-232 status
		if r.serial == nil {
			return 0
		}
		return r.serial.ReadStatus()
	case 16:
		// disk, net SPI data
		spi := r.spi[r.spiSelected]
		if spi == nil {
			return 255
		}
		return spi.ReadData()
	case 20:
		// disk, net SPI status
		// Bit 0: rx ready
		// Other bits unused
		return 1
	case 24:
		// Mouse input / keyboard status
		mouse := r.mouse
		if len(r.keyBuf) > 0 {
			mouse |= 0x10000000
		} else {
			r.progress--
		}
		return mouse
	case 28:
		// Keyboard data (PS2)
		if len(r.keyBuf) == 0 {
			return 0
		}
		scancode := r.keyBuf[0]
		r.keyBuf = r.keyBuf[1:]
		return uint32(scancode)
	case 40:
		// Clipboard control
		if r.clipboard == nil {
			return 0
		}
		return r.clipboard.ReadControl()
	case 44:
		// Clipboard data
		if r.clipboard == nil {
			return 0
		}
		return r.clipboard.ReadData()
	default:
		return 0
	}
}

func (r *RISC) storeIO(address, value uint32) {
	switch address - ioStart {
	case 4:
		// LEDs
		if r.leds != nil {
			r.leds.Write(value)
		}
	case 8:
		// RS-232 data
		if r.serial != nil {
			r.serial.WriteData(value)
		}
	case 16:
		// SPI data
		spi := r.spi[r.spiSelected]
		if spi != nil {
			spi.WriteData(value)
		}
	case 20:
		// SPI control
		// Bit 0-1: chip select
		// Bit 2:   fast mode
		// Bit 3:   network enable
		// Other bits unused
		r.spiSelected = value & 0b0011
	case 40:
		// Clipboard control
		if r.clipboard != nil {
			r.clipboard.WriteControl(value)
		}
	case 44:
		// Clipboard data
		if r.clipboard != nil {
			r.clipboard.WriteData(value)
		}
	}
}

func (r *RISC) SetTime(millis uint32) {
	r.millisecondCounter = millis
}

func (r *RISC) MouseMoved(x, y int) {
	if x >= 0 && x <= 0xFFF {
		r.mouse = (r.mouse &^ 0x00000FFF) | uint32(x)
	}
	if y >= 0 && y <= 0xFFF {
		r.mouse = (r.mouse &^ 0x00FFF000) | (uint32(y) << 12)
	}
}

func (r *RISC) MouseButton(button int, down bool) {
	if button < 1 || button > 3 {
		return
	}
	bit := uint32(1 << (27 - button))
	if down {
		r.mouse = r.mouse | bit
	} else {
		r.mouse = r.mouse &^ bit
	}
}

func (r *RISC) KeyboardInput(ps2commands []byte) {
	r.keyBuf = append(r.keyBuf, ps2commands...)
}

func (r *RISC) Framebuffer() *Framebuffer {
	return &r.framebuffer
}

func (r *RISC) GetFramebufferDamageAndReset() image.Rectangle {
	d := r.damage
	r.resetFramebufferDamage()
	return d
}

func (r *RISC) resetFramebufferDamage() {
	r.damage = image.Rectangle{
		Min: image.Point{X: r.framebuffer.Rect.Max.X / 32, Y: r.framebuffer.Rect.Max.Y},
		Max: image.Point{},
	}
}

// memSize returns the memory size in bytes
func (r *RISC) memSize() int {
	return len(r.Mem) * 4
}

func b2i(b bool) uint32 {
	if b {
		return 1
	}
	return 0
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

type Error struct {
	PC      uint32
	message string
}

func (e *Error) Error() string {
	return e.message
}
