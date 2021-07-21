// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package risc

type Serial interface {
	ReadStatus() uint32
	ReadData() uint32
	WriteData(value uint32)
}

type SPI interface {
	ReadData() uint32
	WriteData(value uint32)
}

type Clipboard interface {
	WriteControl(len uint32)
	ReadControl() uint32
	WriteData(value uint32)
	ReadData() uint32
}

type LED interface {
	Write(value uint32)
}
