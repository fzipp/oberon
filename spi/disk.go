// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package spi

import (
	"fmt"
	"io"
	"os"
)

type Disk struct {
	state  diskState
	file   *os.File
	offset uint32

	rxBuf [128]uint32
	rxIdx int

	txBuf [128 + 2]uint32
	txCnt int
	txIdx int
}

type diskState int

const (
	diskCommand diskState = iota
	diskRead
	diskWrite
	diskWriting
)

func NewDisk(filename string) (*Disk, error) {
	disk := &Disk{
		state: diskCommand,
	}

	if filename == "" {
		return disk, nil
	}

	var err error
	disk.file, err = os.OpenFile(filename, os.O_RDWR, 0o666)
	if err != nil {
		return nil, fmt.Errorf("can't open file \"%s\": %w", filename, err)
	}

	// Check for filesystem-only image, starting directly at sector 1 (DiskAdr 29)
	err = readSector(disk.file, disk.txBuf[:])
	if err != nil {
		return nil, fmt.Errorf("can't read sector: %w", err)
	}
	if disk.txBuf[0] == 0x9B1EA38D {
		disk.offset = 0x80002
	}

	return disk, nil
}

func (d *Disk) WriteData(value uint32) {
	d.txIdx++
	switch d.state {
	case diskCommand:
		if uint8(value) != 0xFF || d.rxIdx != 0 {
			d.rxBuf[d.rxIdx] = value
			d.rxIdx++
			if d.rxIdx == 6 {
				err := d.runCommand()
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "can't run disk command: %s", err)
				}
				d.rxIdx = 0
			}
		}
	case diskRead:
		if d.txIdx == d.txCnt {
			d.state = diskCommand
			d.txCnt = 0
			d.txIdx = 0
		}
	case diskWrite:
		if value == 254 {
			d.state = diskWriting
		}
	case diskWriting:
		if d.rxIdx < 128 {
			d.rxBuf[d.rxIdx] = value
		}
		d.rxIdx++
		if d.rxIdx == 128 {
			err := writeSector(d.file, d.rxBuf[:])
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "can't write to disk: %s", err)
			}
		}
		if d.rxIdx == 130 {
			d.txBuf[0] = 5
			d.txCnt = 1
			d.txIdx = -1
			d.rxIdx = 0
			d.state = diskCommand
		}
	}
}

func (d *Disk) ReadData() uint32 {
	if d.txIdx >= 0 && d.txIdx < d.txCnt {
		return d.txBuf[d.txIdx]
	}
	return 255
}

func (d *Disk) runCommand() error {
	cmd := d.rxBuf[0]
	arg := d.rxBuf[1]<<24 |
		(d.rxBuf[2] << 16) |
		(d.rxBuf[3] << 8) |
		d.rxBuf[4]

	switch cmd {
	case 81:
		d.state = diskRead
		d.txBuf[0] = 0
		d.txBuf[1] = 254
		err := seekSector(d.file, arg-d.offset)
		if err != nil {
			return fmt.Errorf("can't seek sector for read: %w", err)
		}
		err = readSector(d.file, d.txBuf[2:])
		if err != nil {
			return fmt.Errorf("can't read sector: %w", err)
		}
		d.txCnt = 2 + 128
	case 88:
		d.state = diskWrite
		err := seekSector(d.file, arg-d.offset)
		if err != nil {
			return fmt.Errorf("can't seek sector for write: %w", err)
		}
		d.txBuf[0] = 0
		d.txCnt = 1
	default:
		d.txBuf[0] = 0
		d.txCnt = 1
	}
	d.txIdx = -1
	return nil
}

func seekSector(s io.Seeker, secnum uint32) error {
	_, err := s.Seek(int64(secnum)*512, io.SeekStart)
	if err != nil {
		return fmt.Errorf("can't seek sector %d: %w", secnum, err)
	}
	return nil
}

func readSector(r io.Reader, buf []uint32) error {
	var bytes [512]byte
	_, err := r.Read(bytes[:])
	if err != nil && err != io.EOF {
		return fmt.Errorf("can't read bytes: %w", err)
	}
	for i := 0; i < 128; i++ {
		buf[i] = uint32(bytes[i*4+0]) |
			(uint32(bytes[i*4+1]) << 8) |
			(uint32(bytes[i*4+2]) << 16) |
			(uint32(bytes[i*4+3]) << 24)
	}
	return nil
}

func writeSector(w io.Writer, buf []uint32) error {
	var bytes [512]byte
	for i := 0; i < 128; i++ {
		bytes[i*4+0] = uint8(buf[i])
		bytes[i*4+1] = uint8(buf[i] >> 8)
		bytes[i*4+2] = uint8(buf[i] >> 16)
		bytes[i*4+3] = uint8(buf[i] >> 24)
	}
	_, err := w.Write(bytes[:])
	if err != nil {
		return fmt.Errorf("can write bytes: %w", err)
	}
	return nil
}
