package canvas

import (
	"image"

	"github.com/fzipp/oberon/risc"
)

type Context struct {
	config config
	draws  chan<- []byte
	events <-chan Event
	buf    buffer
}

func newContext(draws chan<- []byte, events <-chan Event, config config) *Context {
	return &Context{
		config: config,
		draws:  draws,
		events: events,
	}
}

func (ctx *Context) Events() <-chan Event {
	return ctx.events
}

func (ctx *Context) CanvasWidth() int {
	return ctx.config.width
}

func (ctx *Context) CanvasHeight() int {
	return ctx.config.height
}

const (
	colorBlack = 0x657b83ff
	colorWhite = 0xfdf6e3ff
)

const (
	bUpdateDisplay byte = 1 + iota
	bClipboardWriteText
)

func (ctx *Context) UpdateDisplay(fb *risc.Framebuffer, r image.Rectangle) {
	if r.Min.Y > r.Max.Y {
		return
	}

	cw := ctx.config.width
	ch := ctx.config.height

	x := uint32(r.Min.X * 32)
	y := uint32(ch - r.Max.Y - 1)
	w := uint32((r.Max.X - r.Min.X + 1) * 32)
	h := uint32(r.Max.Y - r.Min.Y + 1)

	ctx.buf.addByte(bUpdateDisplay)
	ctx.buf.addUint32(x)
	ctx.buf.addUint32(y)
	ctx.buf.addUint32(w)
	ctx.buf.addUint32(h)

	for line := r.Max.Y; line >= r.Min.Y; line-- {
		lineStart := line * (cw / 32)
		for col := r.Min.X; col <= r.Max.X; col++ {
			pixels := fb.Pix[lineStart+col]
			for range 32 {
				var color uint32
				if pixels&1 > 0 {
					color = colorWhite
				} else {
					color = colorBlack
				}
				ctx.buf.addUint32(color)
				pixels >>= 1
			}
		}
	}
	ctx.Flush()
}

func (ctx *Context) ClipboardWriteText(text string) {
	ctx.buf.addByte(bClipboardWriteText)
	ctx.buf.addString(text)
	ctx.Flush()
}

func (ctx *Context) Flush() {
	ctx.draws <- ctx.buf.bytes
	ctx.buf.reset()
}
