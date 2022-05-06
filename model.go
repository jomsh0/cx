package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"image/color"
	"os"
	"strconv"
)

type cprop8 uint8

const (
	Y cprop8 = 1 + iota
	U
	V
	R
	G
	B
)

type DispName struct {
	abbr string
	string
}

var propNames = map[cprop8]DispName{
	Y: {"Y", "Luma"},
	U: {"Cb", "Chroma-Blue"},
	V: {"Cr", "Chroma-Red"},
	R: {"R", "Red"},
	G: {"G", "Green"},
	B: {"B", "Blue"},
}

// classic nourishing ANSI names.
const (
	Black = tcell.ColorBlack + iota
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
)
const (
	Foreground = tcell.ColorReset + 1 + iota
	Background
)

var cNames = map[tcell.Color]DispName{
	Black:         {"k", "black"},
	Red:           {"r", "red"},
	Green:         {"g", "green"},
	Yellow:        {"y", "yellow"},
	Blue:          {"b", "blue"},
	Magenta:       {"m", "magenta"},
	Cyan:          {"c", "cyan"},
	White:         {"w", "white"},
	BrightBlack:   {"K", "bright-black"},
	BrightRed:     {"R", "bright-red"},
	BrightGreen:   {"G", "bright-green"},
	BrightYellow:  {"Y", "bright-yellow"},
	BrightBlue:    {"B", "bright-blue"},
	BrightMagenta: {"M", "bright-magenta"},
	BrightCyan:    {"C", "bright-cyan"},
	BrightWhite:   {"W", "bright-white"},

	Foreground: {"fg", "foreground"},
	Background: {"bg", "background"},
}

func (p cprop8) Abbr() string {
	return propNames[p].abbr
}
func (p cprop8) String() string {
	return propNames[p].string
}

type bColor struct {
	color.RGBA
	color.YCbCr
}

type Theme map[tcell.Color]bColor
type Selection struct {
	Theme
	sel []tcell.Color
}

func toYUV(c color.Color) color.YCbCr {
	return color.YCbCrModel.Convert(c).(color.YCbCr)
}

func toRGB(c color.Color) color.RGBA {
	return color.RGBAModel.Convert(c).(color.RGBA)
}

func (p Theme) Init(t []color.Color, fg, bg color.Color) Theme {
	for i, c := range t {
		p[tcell.ColorValid+tcell.Color(i)] = Bcolor(c)
	}
	p[Foreground] = Bcolor(fg)
	p[Background] = Bcolor(bg)

	return p
}

func (t Theme) Copy() Theme {
	tc := make(Theme)
	for c, bc := range t {
		tc[c] = bc
	}
	return tc
}

func (p Theme) From(hex []string, fg, bg string) Theme {
	var t []color.Color
	for _, h := range hex {
		t = append(t, FromHex(h))
	}

	fg_c := FromHex(fg)
	bg_c := FromHex(bg)

	return p.Init(t, fg_c, bg_c)
}

func FromHex(hex string) color.Color {
	var runes []rune
	for _, r := range hex {
		if (r >= '0' && r <= '9') ||
			(r >= 'A' && r <= 'F') ||
			(r >= 'a' && r <= 'f') {
			runes = append(runes, r)
		}
	}
	if len(runes) != 6 {
		panic(fmt.Errorf("bad hex color %s (%v)", hex, runes))
	}
	val, err := strconv.ParseUint(string(runes), 16, 32)
	if err != nil {
		panic("bad hex color")
	}
	return color.RGBA{uint8(val >> 16), uint8(val >> 8), uint8(val), 0xff}
}

func initc(code string, color color.RGBA) {
	os.Stderr.WriteString(
		fmt.Sprintf(
			"\x1b]%s;rgb:%02x/%02x/%02x\x1b\\",
			code, color.R, color.G, color.B,
		),
	)
}

func (t Theme) SetTheme() {
	for c, bc := range t {
		switch c {
		case Foreground:
			initc(fmt.Sprintf("%d", 10), bc.RGBA)
		case Background:
			initc(fmt.Sprintf("%d", 11), bc.RGBA)
		default:
			index := c - tcell.ColorValid
			initc(fmt.Sprintf("4;%d", index), bc.RGBA)
		}
	}
}

func (b *bColor) Access(prop cprop8) uint8 {
	switch prop {
	case Y:
		return b.Y
	case U:
		return b.Cb
	case V:
		return b.Cr
	case R:
		return b.R
	case G:
		return b.G
	case B:
		return b.B
	}
	panic("bad access")
}

func (b *bColor) Set(c color.Color) *bColor {
	*b = Bcolor(c)
	return b
}

func Bcolor(c color.Color) bColor {
	switch c0 := c.(type) {
	case color.YCbCr:
		return bColor{toRGB(c0), c0}
	case color.RGBA:
		return bColor{c0, toYUV(c0)}
	default:
		return bColor{toRGB(c0), toYUV(c0)}
	}
}

func first[T any](first T, _ ...any) T {
	return first
}

func (t Theme) Select(mask uint64) Selection {
	var sel []tcell.Color

	fgMask := uint64(1 << 63)
	bgMask := fgMask >> 1
	restMask := ^uint64(0)

	for c := Black; ; c++ {
		if _, found := t[c]; !found {
			break
		}

		off := uint64(c - Black)
		if mask&(1<<off) != 0 {
			sel = append(sel, c)
		}

		restMask <<= 1
		if (mask&restMask)%bgMask == 0 {
			break
		}
	}

	if mask&fgMask != 0 {
		sel = append(sel, Foreground)
	}
	if mask&bgMask != 0 {
		sel = append(sel, Background)
	}

	return Selection{t, sel}
}

func (s Selection) Iter() []bColor {
	var v []bColor
	for _, c := range s.sel {
		v = append(v, s.Theme[c])
	}
	return v
}

func (s Selection) Adjust(prop cprop8, adj int) Selection {
	if adj == 0 || s.sel == nil {
		return s
	}

	for _, c := range s.sel {
		b := s.Theme[c]

		var ptr *uint8
		switch prop {
		case Y:
			ptr = &b.Y
		case U:
			ptr = &b.Cb
		case V:
			ptr = &b.Cr
		case R:
			ptr = &b.R
		case G:
			ptr = &b.G
		case B:
			ptr = &b.B
		}

		// stupid simple
		val := int(*ptr) + adj
		if val < 0 {
			*ptr = 0
		} else if val > int(uint8(val)) {
			*ptr = ^uint8(0)
		} else {
			*ptr = uint8(val)
		}

		switch prop {
		case Y, U, V:
			b.Set(b.YCbCr)
		case R, G, B:
			b.Set(b.RGBA)
		}

		s.Theme[c] = b
	}

	return s
}
