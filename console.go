package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	//"log"
	"strings"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

const (
	LEFT   uint8 = 0x00
	CENTER       = 0x10
	RIGHT        = 0x20
	TOP          = 0x00
	MIDDLE       = 0x01
	BOTTOM       = 0x02
)

type Padding struct{ t, b, l, r int }

type Rec struct {
	x, y int
	w, h int
	R    *Rec
}

func (r Rec) Rec(x, y, w, h int) Rec {
	return Rec{x: x, y: y, w: w, h: h}
}

func (r Rec) GetRect() (int, int, int, int) {
	return r.x, r.y, r.w, r.h
}

func (r Rec) Pos(R Rec, align uint8) Rec {

	switch align & (LEFT | CENTER | RIGHT) {
	case CENTER:
		r.x += (R.w-r.w)/2 - 1
	case RIGHT:
		r.x += R.w - r.w
	}

	switch align & (TOP | MIDDLE | BOTTOM) {
	case MIDDLE:
		r.y += (R.h-r.h)/2 - 1
	case BOTTOM:
		r.y += R.h - r.h
	}

	r.R = &R
	return r
}

func (r Rec) Ab() Rec {
	for _r := r.R; _r != nil; _r = _r.R {
		r.x += _r.x
		r.y += _r.y
	}
	r.R = nil
	return r
}

func (r Rec) Rl(R Rec) Rec {
	ra, Ra := r.Ab(), R.Ab()
	x := ra.x - Ra.x
	y := ra.y - Ra.y
	return Rec{x, y, r.w, r.h, &R}
}

func (r Rec) Ix(R Rec) Rec {
	r, R = r.Ab(), R.Ab()
	x, y := max(r.x, R.x), max(r.y, R.y)

	return Rec{
		x: x,
		y: y,
		w: min(r.x+r.w, R.x+R.w) - x,
		h: min(r.y+r.h, R.y+R.h) - y,
	}
}

func (r Rec) Text(txt string) Rec {
	return Rec{w: len([]rune(txt)), h: 1}
}

func (r Rec) Each(cb func(int, int, Rec)) Rec {
	rA := r.Ab()
	for x := rA.x; x < rA.x+rA.w; x++ {
		for y := rA.y; y < rA.y+rA.h; y++ {
			cb(x, y, r)
		}
	}
	return r
}

type Control struct {
	*tview.Box
	Con     *Console
	padding Padding
	title   string
}

func (ctl *Control) Init(Con *Console, title string) *Control {
	if ctl == nil {
		ctl = new(Control)
	}
	ctl.Con = Con
	ctl.Box = tview.NewBox().SetTitle(strings.ToUpper(title))
	return ctl
}

func (ctl *Control) SetBorderPadding(t, b, l, r int) *Control {
	ctl.padding = Padding{t, b, l, r}
	return ctl
}

// Always return the same InnerRect whether or not borders are drawn
func (ctl *Control) GetInnerRect() (int, int, int, int) {
	x, y, w, h := ctl.Box.GetRect()
	x += ctl.padding.l + 1
	y += ctl.padding.t + 1
	w -= ctl.padding.l + ctl.padding.r + 2
	h -= ctl.padding.t + ctl.padding.b + 2

	return x, y, w, h
}

func (ctl *Control) DrawForSubclass(screen tcell.Screen, p tview.Primitive) {
	// ctl.SetBorder(false).SetBorderAttributes(tcell.AttrReverse)
	ctl.Box.DrawForSubclass(screen, p)
	Rec{}.Rec(ctl.GetRect())

	if ctl.HasFocus() {

		_, _, _, h := txt.GetRect()
		y1, y2 := h/2-2, h/2+3
		if y1 < 0 {
			y1 = 0
		}
		if y2 > h {
			y2 = h
		}
		for j := y1; j < y2; j++ {
			txt.FillRow(screen, ' ', 0, j)
		}
	}
	// ru = tview.BoxDrawingsLightHorizontal
}

// func (txt *Text) Print(screen tcell.Screen, x0, y0 int) *Text {
// 	roos := []rune(txt.string)
// 	if len(roos) == 0 {
// 		return txt
// 	}
// 	x, y, w, h := txt.GetRect()
// 	txt.x0, txt.y0 = x0, y0
// 	switch txt.align {
// 	case TopCenter, MidCenter, BottomCenter:
// 		x += w/2 - len(roos)/2 - 1
// 	case TopRight, MidRight, BottomRight:
// 		x += w - len(roos)
// 	}
// 	switch txt.align {
// 	case MidLeft, MidCenter, MidRight:
// 		y += h/2 - 1
// 	case BottomLeft, BottomCenter, BottomRight:
// 		y += h - 1
// 	}
// 	x += x0
// 	y += y0
// 	for i, ru := range roos {
// 		screen.SetContent(x+i, y, ru, nil, txt.Style)
// 	}
// 	return txt
// }

type CtlTweaker struct {
	*Control
	prop cprop8
}

func (tw *CtlTweaker) Init(Con *Console, prop cprop8) *CtlTweaker {
	tw.Control = new(Control).
		Init(Con, prop.String()).
		SetBorderPadding(0, 0, 0, 0)
	tw.prop = prop
	return tw
}

func (tw *CtlTweaker) drawGrid(screen tcell.Screen) {
	x, y, w, h := tw.GetInnerRect()
	y_mid := y + h/2

	mid_ln := tview.BoxDrawingsLightDoubleDashHorizontal
	mark_dn := tview.BoxDrawingsLightDownAndHorizontal
	mark_up := tview.BoxDrawingsLightUpAndHorizontal

	for i, ru := range []rune(tw.prop.Abbr()) {
		screen.SetContent(x+i, y_mid-1, ru, nil, tw.dStyle().Bold(true))
	}
	for i := 0; i < w; i++ {
		screen.SetContent(x+i, y_mid, mid_ln, nil, tw.dStyle())
	}
	for i, x_off := 0, 0; ; i++ {

		if i%2 == 0 {
			screen.SetContent(x+x_off, y_mid, mark_dn, nil, tw.dStyle())
		} else {
			screen.SetContent(x+x_off, y_mid, mark_up, nil, tw.dStyle())
		}
		if i%2 == 0 && tw.HasFocus() {
			xpos := x + x_off - 1
			label := fmt.Sprintf("%d", i*32)
			switch i {
			case 0, 2:
				xpos += 1
			case 8:
				label = "255"
				xpos -= 1
			}
			for i, ru := range []rune(label) {
				screen.SetContent(xpos+i, y_mid+1, ru, nil, tw.dStyle())
			}
		}
		if i == 8 {
			break
		}
		x_off += (w - x_off) / (8 - i)
	}
}

func (tw *CtlTweaker) dStyle() tcell.Style {
	color := tcell.ColorReset
	if tw.HasFocus() {
		color = tcell.ColorGray
	}
	return tcell.StyleDefault.Foreground(color).Reverse(tw.HasFocus()).Dim(!tw.HasFocus())
}

func (tw *CtlTweaker) Draw(screen tcell.Screen) {
	tw.Control.DrawForSubclass(screen, tw)
	tw.drawGrid(screen)

	x, y, w, h := tw.GetInnerRect()
	y_mid := y + h/2

	// keep track of positions of already-drawn markers to detect collisions
	model := make([]tcell.Color, w)
	theme := tw.Con.Theme()

	for _, color := range tw.Con.Selection() {
		b := theme[color]
		v := b.Access(tw.prop)

		off := int(v) * w / 255
		if off == w {
			off--
		}
		// attempt to spread out collisions into adjacent cells
		for i := 0; i <= 3; i++ {
			if off-i >= 0 && model[off-i] == 0 {
				off -= i
				break
			}
			if off+i < w && model[off+i] == 0 {
				off += i
				break
			}
		}
		model[off] = color

		ch, _, _, _ := screen.GetContent(x+off, y_mid)
		// background color requires special handling
		if b.RGBA == theme[Background].RGBA {
			ch = '▒' //▒ ░ █
			color = tcell.ColorDefault
		} else {
			ch = '█'
		}

		screen.SetContent(x+off, y_mid, ch, nil, tw.dStyle().Reverse(false).Foreground(color))
	}
}

func (tw *CtlTweaker) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return tw.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {

		mod := event.Modifiers()
		p_adj := 0 //property adjust
		key := event.Key()

		if key == tcell.KeyRune {
			rune := event.Rune()
			if rune < 'a' {
				mod |= tcell.ModShift
			}
			switch rune {
			case 'H', 'h':
				key = tcell.KeyLeft
			case 'L', 'l':
				key = tcell.KeyRight
			}
		}

		switch key {
		case tcell.KeyCtrlL, tcell.KeyRight:
			p_adj = 1
		case tcell.KeyCtrlH, tcell.KeyLeft:
			p_adj = -1
		}

		if p_adj == 0 {
			return
		}

		if mod&tcell.ModAlt == 0 {
			p_adj *= 3
		}
		if mod&tcell.ModShift != 0 {
			p_adj *= 3
		}

		tw.Con.Adjust(tw.prop, p_adj)
	})
}

type Lmodel struct {
	tcell.Color
	string
}

type CtlSelector struct {
	*Control
	mask   CMask
	model  []Lmodel
	curcol int
}

func (cs *CtlSelector) Init(Con *Console) *CtlSelector {
	cs.Control = new(Control).
		Init(Con, "selection").
		//SetBlurBorder(true).
		SetBorderPadding(1, 1, 1, 1)

	cols := len(Con.theme)
	cs.model = make([]Lmodel, cols)

	for color := range Con.theme {
		name, found := cNames[color]
		if !found {
			name.abbr = fmt.Sprintf("%v", CMask(0).Index(color))
		}

		col := int(CMask(0).IndexMod(cols, color))
		cs.model[col] = Lmodel{color, name.abbr}
	}

	return cs.Reset()
}

func (cs *CtlSelector) Left() *CtlSelector {
	cs.curcol--
	if cs.curcol < 0 {
		cs.curcol = len(cs.model) + cs.curcol
	}
	return cs
}

func (cs *CtlSelector) Right() *CtlSelector {
	cs.curcol = (cs.curcol + 1) % len(cs.model)
	return cs
}

func (cs *CtlSelector) Pick() *CtlSelector {
	return cs.Toggle(CMask(0).Mask(cs.model[cs.curcol].Color))
}

func (cs *CtlSelector) Reset() *CtlSelector {
	cs.mask = CMask(0).Colors()
	return cs
}

func (cs *CtlSelector) Toggle(mask CMask) *CtlSelector {
	if cs.mask&mask == mask {
		cs.mask ^= mask
	} else {
		cs.mask |= mask
	}
	return cs
}

func (cs *CtlSelector) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return cs.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		//mod := event.Modifiers()
		key := event.Key()

		if key == tcell.KeyRune {
			rune := event.Rune()
			switch rune {
			case 'H', 'h':
				cs.Left()
			case 'L', 'l':
				cs.Right()
			case ' ':
				cs.Pick()
			case 'c':
				cs.Toggle(CMask(0).Colors())
			}
			return
		}

		switch key {
		case tcell.KeyCtrlL, tcell.KeyRight:
			cs.Right()
		case tcell.KeyCtrlH, tcell.KeyLeft:
			cs.Left()
		case tcell.KeyEnter:
			cs.Pick()
		}
	})
}

func (cs *CtlSelector) Draw(screen tcell.Screen) {
	cs.DrawForSubclass(screen, cs)

	x, y, w, h := cs.GetInnerRect()
	y_c := y + h/2
	y_l := y_c - 1

	var csize int
	for csize = 6; csize > 2; csize-- {
		if len(cs.model)*csize <= w {
			break
		}
	}
	for col, mod := range cs.model {
		selected := cs.mask.Has(mod.Color)

		block := '█' //░

		x_c := x + col*csize

		roos := []rune(mod.string)
		lpad := (csize - len(roos)) / 2

		st_l := tcell.StyleDefault.Dim(!selected).Reverse(selected).Bold(selected)
		st_c := tcell.StyleDefault.Foreground(mod.Color).Dim(!selected)

		if mod.Color == Background {
			block = ' '
		}
		if cs.HasFocus() && col == cs.curcol {
			st_l = st_l.Foreground(BrightBlue).Background(BrightWhite).Reverse(true).Bold(true).Dim(false)
		}

		for i := 0; i < csize; i++ {
			ru := ' '
			pos := i - lpad
			if pos >= 0 && pos < len(roos) {
				ru = roos[pos]
			}

			screen.SetContent(x_c+i, y_l, ru, nil, st_l)
			screen.SetContent(x_c+i, y_c, block, nil, st_c)
		}
	}
}

type Console struct {
	*tview.Flex
	csel    *CtlSelector
	theme   Theme
	reset   Theme
	focus_i int
}

func (Con *Console) Init(t Theme) *Console {
	Con.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow)

	Con.reset = t
	Con.Reset()

	Con.csel = new(CtlSelector).Init(Con)
	Con.AddItem(Con.csel, 0, 1, true)

	for _, p := range []cprop8{Y, Cb, Cr, R, G, B} {
		Con.AddItem(new(CtlTweaker).Init(Con, p), 0, 1, true)
	}
	return Con
}

func (Con *Console) Adjust(prop cprop8, adj int) Theme {
	return Con.theme.Adjust(Con.csel.mask, prop, adj).Apply()
}
func (Con *Console) Selection() []tcell.Color {
	return Con.csel.mask.Iter()
}
func (Con *Console) Theme() Theme {
	return Con.theme
}
func (Con *Console) Reset() *Console {
	Con.theme = Con.reset.Copy()
	return Con
}

func (Con *Console) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return Con.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {

		f_i := Con.focus_i
		key := event.Key()

		if key == tcell.KeyRune {
			switch event.Rune() {
			case 'J', 'j':
				key = tcell.KeyDown
			case 'K', 'k':
				key = tcell.KeyUp
			}
		}

		switch key {

		case tcell.KeyTab, tcell.KeyCtrlJ, tcell.KeyDown:
			f_i++

		case tcell.KeyBacktab, tcell.KeyCtrlK, tcell.KeyUp:
			f_i--

		case tcell.KeyCtrlR:
			Con.Reset().Theme().Apply()
			return

		default:
			Con.GetItem(Con.focus_i).InputHandler()(event, setFocus)
			return
		}

		if f_i < 0 {
			f_i += Con.GetItemCount()
		}
		f_i %= Con.GetItemCount()
		Con.focus_i = f_i
		setFocus(Con.GetItem(f_i))
	})
}
