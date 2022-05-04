package main

import (
	"github.com/gdamore/tcell/v2"
	"image/color"
)

type Cprop uint8

const (
	Y Cprop = 1 + iota
	U
	V
	R
	G
	B
)

var propNames = map[Cprop]struct {
	rune
	string
}{
	Y: {'Y', "Luma"},
	U: {'U', "Chroma-U"},
	V: {'V', "Chroma-V"},
	R: {'R', "Red"},
	G: {'G', "Green"},
	B: {'B', "Blue"},
}

type Cindex uint8

const (
	Black Cindex = 0 + iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White

	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite

	Cind_ENDPALETTE

	Foreground
	Background

	Cind_COUNT
)

var cmap = map[Cindex]struct {
	string
	tcell.Color
}{
	Black:         {"black", tcell.ColorBlack},
	Red:           {"red", tcell.ColorMaroon},
	Green:         {"green", tcell.ColorGreen},
	Yellow:        {"yellow", tcell.ColorOlive},
	Blue:          {"blue", tcell.ColorNavy},
	Magenta:       {"magenta", tcell.ColorPurple},
	Cyan:          {"cyan", tcell.ColorTeal},
	White:         {"white", tcell.ColorSilver},
	BrightBlack:   {"bright-black", tcell.ColorGray},
	BrightRed:     {"bright-red", tcell.ColorRed},
	BrightGreen:   {"bright-green", tcell.ColorLime},
	BrightYellow:  {"bright-yellow", tcell.ColorYellow},
	BrightBlue:    {"bright-blue", tcell.ColorBlue},
	BrightMagenta: {"bright-magenta", tcell.ColorFuchsia},
	BrightCyan:    {"bright-cyan", tcell.ColorAqua},
	BrightWhite:   {"bright-white", tcell.ColorWhite},
}

func (p Cprop) Rune() rune {
	return propNames[p].rune
}

func (p Cprop) String() string {
	return propNames[p].string
}

func (i Cindex) String() string {
	return cmap[i].string
}

func (i Cindex) Color() tcell.Color {
	return cmap[i].Color
}

var cind_all [Cind_ENDPALETTE]Cindex

func CindAll() []Cindex {
	if cind_all[1] != 0 {
		return cind_all[:]
	}
	for i := 0; i < len(cind_all); i++ {
		cind_all[i] = Cindex(i)
	}
	return cind_all[:]
}

type Model struct {
	rgb [Cind_COUNT]color.RGBA
	yuv [Cind_COUNT]color.YCbCr
}

type Pvect [Cind_COUNT]uint8

func toYUV(c color.Color) color.YCbCr {
	return color.YCbCrModel.Convert(c).(color.YCbCr)
}

func toRGB(c color.Color) color.RGBA {
	return color.RGBAModel.Convert(c).(color.RGBA)
}

func (m *Model) Theme() *Theme {
	var palette []color.RGBA
	for i, c := range m.rgb {
		if i >= int(Cind_ENDPALETTE) {
			break
		}
		palette = append(palette, c)
	}
	return &Theme{
		palette:    palette,
		foreground: m.rgb[Foreground],
		background: m.rgb[Background],
	}
}

func (m Model) New(t Theme) *Model {
	for i, c := range t.palette {
		if i >= int(Cind_ENDPALETTE) {
			break
		}
		m.rgb[i] = c
	}

	m.rgb[Foreground] = t.foreground
	m.rgb[Background] = t.background

	m.genYUV()

	return &m
}

func (m *Model) genRGB() {
	for i, c := range m.yuv {
		m.rgb[i] = toRGB(c)
	}
}

func (m *Model) genYUV() {
	for i, c := range m.rgb {
		m.yuv[i] = toYUV(c)
	}
}

func access(c color.Color, p Cprop) (uint8, bool) {
	switch c0 := c.(type) {
	case color.YCbCr:
		switch p {
		case Y:
			return c0.Y, true
		case U:
			return c0.Cb, true
		case V:
			return c0.Cr, true
		}
	case color.RGBA:
		switch p {
		case R:
			return c0.R, true
		case G:
			return c0.G, true
		case B:
			return c0.B, true
		}
	}
	return 0, false
}

func first[T any](first T, _ ...any) T {
	return first
}

func (m *Model) Access(p Cprop, i Cindex) (uint8, bool) {
	switch p {
	case Y, U, V:
		return access(m.yuv[i], p)
	case R, G, B:
		return access(m.rgb[i], p)
	}
	return 0, false
}

func (m *Model) Update(p Cprop, a *Pvect) {
	switch p {
	case Y, U, V:
		for i, c := range m.yuv {
			a[i], _ = access(c, p)
		}
	case R, G, B:
		for i, c := range m.rgb {
			a[i], _ = access(c, p)
		}
	}
}

func (m *Model) Adjust(p Cprop, sel []Cindex, adj int) {
	if adj == 0 {
		return
	}
	if sel == nil {
		sel = CindAll()
	}

	for _, i := range sel {
		if i >= Cind_ENDPALETTE {
			panic("bad selection")
		}

		var ptr *uint8
		switch p {
		case Y:
			ptr = &m.yuv[i].Y
		case U:
			ptr = &m.yuv[i].Cb
		case V:
			ptr = &m.yuv[i].Cr
		case R:
			ptr = &m.rgb[i].R
		case G:
			ptr = &m.rgb[i].G
		case B:
			ptr = &m.rgb[i].B
		}

		// stupid simple
		val := int(*ptr) + adj
		if val < 0 {
			val = 0
		}
		*ptr = uint8(val)
	}

	switch p {
	case Y, U, V:
		m.genRGB()

	case R, G, B:
		m.genYUV()
	}
}
