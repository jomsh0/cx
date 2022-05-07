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
	Cb
	Cr
	R
	G
	B
)

type DispName struct {
	abbr string
	string
}

var propNames = map[cprop8]DispName{
	Y:  {"Y′", "luma"},
	Cb: {"Cb", "chroma-blue"},
	Cr: {"Cr", "chroma-red"},
	R:  {"R", "red"},
	G:  {"G", "green"},
	B:  {"B", "blue"},
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
type CMask uint64

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

func (t Theme) Apply() Theme {
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
	return t
}

func (b bColor) Access(prop cprop8) uint8 {
	switch prop {
	case Y:
		return b.Y
	case Cb:
		return b.Cb
	case Cr:
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

func (b bColor) Set(c color.Color) bColor {
	return Bcolor(c)
}

func (b bColor) SetProp(prop cprop8, val uint8) bColor {
	switch prop {
	case Y:
		b.Y = val
	case Cb:
		b.Cb = val
	case Cr:
		b.Cr = val
	case R:
		b.R = val
	case G:
		b.G = val
	case B:
		b.B = val
	}
	switch prop {
	case Y, Cb, Cr:
		return Bcolor(b.YCbCr)
	case R, G, B:
		return Bcolor(b.RGBA)
	}
	panic("invalid property")
}

func (b bColor) Adjust(prop cprop8, adj int) bColor {
	val := int(b.Access(prop)) + adj
	if val < 0 {
		val = 0
	} else if val > 0xff {
		val = 0xff
	}
	return b.SetProp(prop, uint8(val))
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

func (cm CMask) Iter() []tcell.Color {
	var sel []tcell.Color

	for c := Black; ; c++ {
		if cm.Has(c) {
			sel = append(sel, c)
		}

		restMask := cm.Mask(Background) - cm.Mask(c)<<1
		if cm&restMask == 0 {
			break
		}
	}

	if cm.Has(Foreground) {
		sel = append(sel, Foreground)
	}
	if cm.Has(Background) {
		sel = append(sel, Background)
	}

	return sel
}

func (cm CMask) Has(color tcell.Color) bool {
	return cm&cm.Mask(color) != 0
}
func (cm CMask) Index(color tcell.Color) int {
	indx := int(color % (1 << 8))
	if color&tcell.ColorSpecial != 0 {
		return -indx
	}
	return indx
}
func (cm CMask) IndexMod(mod int, color tcell.Color) int {
	return (cm.Index(color) + mod) % mod
}
func (cm CMask) Mask(color tcell.Color) CMask {
	return 1 << cm.IndexMod(64, color)
}

func (cm CMask) Interval(c1, c2 tcell.Color) CMask {
	cL, cH := c1, c2
	if c2 < c1 {
		cL, cH = c2, c1
	}
	return cm.Mask(cH)<<1 - cm.Mask(cL)
}

func (cm CMask) _Colors() CMask {
	return cm.Interval(Red, Cyan)
}
func (cm CMask) BrightColors() CMask {
	return cm.Interval(BrightRed, BrightCyan)
}
func (cm CMask) Colors() CMask {
	return cm._Colors() | cm.BrightColors()
}

func (cm CMask) _Grays() CMask {
	return cm.Mask(Black) | cm.Mask(White)
}
func (cm CMask) BrightGrays() CMask {
	return cm.Mask(BrightBlack) | cm.Mask(BrightWhite)
}
func (cm CMask) Grays() CMask {
	return cm._Grays() | cm.BrightGrays() | cm.Mask(Foreground) | cm.Mask(Background)
}

func (t Theme) Adjust(cm CMask, prop cprop8, adj int) Theme {
	if adj == 0 || cm == 0 {
		return t
	}

	for _, color := range cm.Iter() {
		t[color] = t[color].Adjust(prop, adj)
	}

	return t
}
