// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

// Package fp implements floating-point arithmetic operations for the
// Oberon RISC emulator.
//
// See https://people.inf.ethz.ch/wirth/ProjectOberon/PO.Computer.pdf
// section 16.3. "Floating-point arithmetic".
package fp

func Add(x, y uint32, u, v bool) uint32 {
	xs := (x & 0x80000000) != 0
	var xe uint32
	var x0 int32
	if !u {
		xe = (x >> 23) & 0xFF
		xm := ((x & 0x7FFFFF) << 1) | 0x1000000
		if xs {
			x0 = int32(-xm)
		} else {
			x0 = int32(xm)
		}
	} else {
		xe = 150
		x0 = int32(x&0x00FFFFFF) << 8 >> 7
	}

	ys := (y & 0x80000000) != 0
	ye := (y >> 23) & 0xFF
	ym := (y & 0x7FFFFF) << 1
	if !u && !v {
		ym |= 0x1000000
	}
	var y0 int32
	if ys {
		y0 = int32(-ym)
	} else {
		y0 = int32(ym)
	}

	var e0 uint32
	var x3, y3 int32
	if ye > xe {
		shift := ye - xe
		e0 = ye
		if shift > 31 {
			x3 = x0 >> 31
		} else {
			x3 = x0 >> shift
		}
		y3 = y0
	} else {
		shift := xe - ye
		e0 = xe
		x3 = x0
		if shift > 31 {
			y3 = y0 >> 31
		} else {
			y3 = y0 >> shift
		}
	}

	xs_ := b2i(xs)
	ys_ := b2i(ys)
	sum := ((xs_ << 26) | (xs_ << 25) | uint32(x3&0x01FFFFFF)) +
		((ys_ << 26) | (ys_ << 25) | uint32(y3&0x01FFFFFF))

	var s uint32
	if (sum & (1 << 26)) > 0 {
		s = -sum
	} else {
		s = sum
	}
	s = (s + 1) & 0x07FFFFFF

	e1 := e0 + 1
	t3 := s >> 1
	if (s & 0x3FFFFFC) != 0 {
		for (t3 & (1 << 24)) == 0 {
			t3 <<= 1
			e1--
		}
	} else {
		t3 <<= 24
		e1 -= 24
	}

	xn := (x & 0x7FFFFFFF) == 0
	yn := (y & 0x7FFFFFFF) == 0

	if v {
		return uint32(int32(sum<<5) >> 6)
	} else if xn {
		if u || yn {
			return 0
		}
		return y
	} else if yn {
		return x
	} else if (t3&0x01FFFFFF) == 0 || (e1&0x100) != 0 {
		return 0
	} else {
		return ((sum & 0x04000000) << 5) | (e1 << 23) | ((t3 >> 1) & 0x7FFFFF)
	}
}

func Mul(x, y uint32) uint32 {
	sign := (x ^ y) & 0x80000000
	xe := (x >> 23) & 0xFF
	ye := (y >> 23) & 0xFF

	xm := (x & 0x7FFFFF) | 0x800000
	ym := (y & 0x7FFFFF) | 0x800000
	m := uint64(xm) * uint64(ym)

	e1 := (xe + ye) - 127
	var z0 uint32
	if (m & (1 << 47)) != 0 {
		e1++
		z0 = uint32(((m >> 23) + 1) & 0xFFFFFF)
	} else {
		z0 = uint32(((m >> 22) + 1) & 0xFFFFFF)
	}

	if xe == 0 || ye == 0 {
		return 0
	} else if (e1 & 0x100) == 0 {
		return sign | ((e1 & 0xFF) << 23) | (z0 >> 1)
	} else if (e1 & 0x80) == 0 {
		return sign | (0xFF << 23) | (z0 >> 1)
	} else {
		return 0
	}
}

func Div(x, y uint32) uint32 {
	sign := (x ^ y) & 0x80000000
	xe := (x >> 23) & 0xFF
	ye := (y >> 23) & 0xFF

	xm := (x & 0x7FFFFF) | 0x800000
	ym := (y & 0x7FFFFF) | 0x800000
	q1 := uint32(uint64(xm) * (1 << 25) / uint64(ym))

	e1 := (xe - ye) + 126
	var q2 uint32
	if (q1 & (1 << 25)) != 0 {
		e1++
		q2 = (q1 >> 1) & 0xFFFFFF
	} else {
		q2 = q1 & 0xFFFFFF
	}
	q3 := q2 + 1

	if xe == 0 {
		return 0
	} else if ye == 0 {
		return sign | (0xFF << 23)
	} else if (e1 & 0x100) == 0 {
		return sign | ((e1 & 0xFF) << 23) | (q3 >> 1)
	} else if (e1 & 0x80) == 0 {
		return sign | (0xFF << 23) | (q2 >> 1)
	} else {
		return 0
	}
}

type IdivResult struct {
	Quot, Rem uint32
}

func Idiv(x, y uint32, signedDiv bool) IdivResult {
	sign := (int32(x) < 0) && signedDiv
	var x0 uint32
	if sign {
		x0 = -x
	} else {
		x0 = x
	}

	RQ := uint64(x0)
	for S := 0; S < 32; S++ {
		w0 := uint32(RQ >> 31)
		w1 := w0 - y
		if int32(w1) < 0 {
			RQ = (uint64(w0) << 32) | ((RQ & 0x7FFFFFFF) << 1)
		} else {
			RQ = (uint64(w1) << 32) | ((RQ & 0x7FFFFFFF) << 1) | 1
		}
	}

	d := IdivResult{Quot: uint32(RQ), Rem: uint32(RQ >> 32)}
	if sign {
		d.Quot = -d.Quot
		if d.Rem > 0 {
			d.Quot--
			d.Rem = y - d.Rem
		}
	}
	return d
}

func b2i(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
