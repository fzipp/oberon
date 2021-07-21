// Copyright 2021 Frederik Zipp and others; see NOTICE file.
// Use of this source code is governed by the ISC license that
// can be found in the LICENSE file.

package serial

import (
	"errors"
	"fmt"
	"os"
)

const (
	recName = "PCLink.REC" // e.g. echo Test.Mod > PCLink.REC
	sndName = "PCLink.SND"
)

const (
	modeACK = 0x10
	modeREC = 0x21
	modeSND = 0x22
)

type PCLink struct {
	txCount     int
	rxCount     int
	filename    string
	filenameLen int
	mode        uint32
	f           *os.File
	fileLen     int64
	buf         [257]byte
}

func (p *PCLink) ReadStatus() uint32 {
	var status uint32 = 3
	if p.mode != 0 {
		return status
	}
	status = 2

	err := p.getJob(recName)
	if err == nil {
		info, err := os.Stat(p.filename)
		if err != nil {
			return status
		}
		defer cleanUp(recName, err)
		if info.Size() < 0 || info.Size() >= 0x1000000 {
			return status
		}
		p.f, err = os.Open(p.filename)
		if err != nil {
			return status
		}
		p.fileLen = info.Size()
		p.mode = modeREC
		fmt.Printf("PCLink REC Filename: %s size %d\n", p.filename, p.fileLen)
		return status
	}

	err = p.getJob(sndName)
	if err == nil {
		p.f, err = os.Create(p.filename)
		if err != nil {
			return status
		}
		defer cleanUp(sndName, err)
		p.fileLen = -1
		p.mode = modeSND
		fmt.Printf("PCLink SND Filename: %s\n", p.filename)
		return status
	}

	return status
}

func (p *PCLink) ReadData() uint32 {
	defer func() { p.rxCount++ }()
	if p.mode == 0 {
		return 0
	}
	var ch uint32
	if p.rxCount == 0 {
		ch = p.mode
	} else if p.rxCount < p.filenameLen {
		ch = uint32(p.filename[p.rxCount-1])
	} else if p.rxCount == p.filenameLen {
		ch = 0
	} else if p.mode == modeSND {
		ch = uint32(modeACK)
		if p.fileLen == 0 {
			p.mode = 0
			cleanUp(sndName, nil)
		}
	} else {
		pos := (p.rxCount - p.filenameLen - 1) % 256
		if pos == 0 || p.fileLen == 0 {
			if p.fileLen > 255 {
				ch = 255
			} else {
				ch = uint32(byte(p.fileLen))
				if p.fileLen == 0 {
					p.mode = 0
					cleanUp(recName, nil)
				}
			}
		} else {
			var buf [1]byte
			_, err := p.f.Read(buf[:])
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "can't read from file: %s", err)
			}
			ch = uint32(buf[0])
			p.fileLen--
		}
	}
	return ch
}

func (p *PCLink) WriteData(value uint32) {
	defer func() { p.txCount++ }()
	if p.mode == 0 {
		return
	}
	if p.txCount == 0 {
		if value == uint32(modeACK) {
			return
		}
		p.f.Close()
		p.f = nil
		if p.mode == modeSND {
			cleanUp(p.filename, nil) // file not found, delete file created
			cleanUp(sndName, nil)
		} else {
			cleanUp(recName, nil)
		}
		p.mode = 0
	} else if p.mode == modeSND {
		var lim int
		pos := (p.txCount - 1) % 256
		p.buf[pos] = byte(value)
		lim = int(p.buf[0])
		if pos == lim {
			_, err := p.f.Write(p.buf[1 : lim+1])
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "can't write to file: %s", err)
			}
			if lim < 255 {
				p.fileLen = 0
				p.f.Close()
			}
		}
	}
}

func (p *PCLink) getJob(jobName string) error {
	info, err := os.Stat(jobName)
	if err != nil {
		return err
	}
	if info.Size() <= 0 || info.Size() > 33 {
		return errors.New("file size must be > 0 and <= 33 bytes")
	}
	f, err := os.Open(jobName)
	if err != nil {
		return cleanUp(jobName, err)
	}
	defer f.Close()
	_, err = fmt.Fscanf(f, "%31s", &p.filename)
	if err != nil {
		return err
	}
	p.txCount = 0
	p.rxCount = 0
	p.filenameLen = len(p.filename) + 1
	return nil
}

func cleanUp(filename string, err error) error {
	err2 := os.Remove(filename)
	if err2 != nil {
		return err2
	}
	return err
}
