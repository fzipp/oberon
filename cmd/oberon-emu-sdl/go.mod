module github.com/fzipp/oberon/cmd/oberon-emu-sdl

go 1.16

require (
	github.com/fzipp/oberon v0.0.0-00010101000000-000000000000
	github.com/veandco/go-sdl2 v0.4.8
)

replace github.com/fzipp/oberon => ../..
