package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	//"log"
	"strings"
)

type Padding struct {
	t int
	b int
	l int
	r int
}

type Control struct {
	*tview.Box
	C        *Console
	padding  Padding
	title    string
	b_title  string
	b_off    int
	b_border bool
}

type CtlTweaker struct {
	*Control
	prop cprop8
}

func (ctl *Control) Init(C *Console, title string) *Control {
	if ctl == nil {
		ctl = new(Control)
	}
	ctl.C = C
	ctl.title = title
	ctl.Box = tview.NewBox()
	return ctl
}

func (ctl *Control) setTitle() {
	if !ctl.HasFocus() {
		ctl.Box.SetTitle(" " + ctl.title + " ")
	} else {
		ctl.Box.SetTitle("[ [::b]" + ctl.title + "[::-] ]")
	}
}

func (ctl *Control) SetBlurBorder(b_border bool) *Control {
	ctl.b_border = b_border
	return ctl
}

func (ctl *Control) SetBlurTitle(b_title string, b_off int) *Control {
	ctl.b_title = b_title
	ctl.b_off = b_off
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

func (ctl *Control) drawBlurTitle(screen tcell.Screen) {
	roos := []rune(ctl.b_title)
	if len(roos) == 0 {
		return
	}

	x, y, w, h := ctl.GetRect()
	y_mid := y + h/2
	y_off := ctl.b_off
	if y_off < 0 {
		y_off += y_mid
	} else {
		y_off += y
	}

	lpad := (w - len(roos)) / 2
	for i, ru := range roos {
		screen.SetContent(x+lpad+i, y_off, ru, nil, tcell.StyleDefault.Underline(false))
	}
}

// Only draw borders when in focus
func (ctl *Control) DrawForSubclass(screen tcell.Screen, p tview.Primitive) {
	var b_attr tcell.AttrMask
	if !ctl.HasFocus() {
		b_attr = tcell.AttrDim
	}
	ctl.SetBorder(ctl.b_border || ctl.HasFocus()).
		SetBorderAttributes(b_attr)
	ctl.setTitle()

	ctl.Box.DrawForSubclass(screen, p)
	if !ctl.HasFocus() {
		ctl.drawBlurTitle(screen)
	}
}

func (tw *CtlTweaker) Init(C *Console, prop cprop8) *CtlTweaker {
	tw.Control = new(Control).
		Init(C, prop.String()).
		SetBlurTitle("[ "+prop.Abbr()+" ]", -2).
		SetBorderPadding(0, 0, 1, 1)
	tw.prop = prop
	return tw
}

func (tw *CtlTweaker) drawGrid(screen tcell.Screen) (int, int, int, int) {
	x, y, w, h := tw.GetInnerRect()
	y_mid := y + h/2

	midline := strings.Repeat(string(tview.BoxDrawingsLightDoubleDashHorizontal), w)
	tview.PrintSimple(screen, "[::d]"+midline, x, y_mid)

	mark_dn := string(tview.BoxDrawingsLightDownAndHorizontal)
	mark_up := string(tview.BoxDrawingsLightUpAndHorizontal)

	for i, x_off := 0, 0; ; i++ {
		if i%2 == 0 {
			tview.PrintSimple(screen, "[::d]"+mark_dn, x+x_off, y_mid)
		} else {
			tview.PrintSimple(screen, "[::d]"+mark_up, x+x_off, y_mid)
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
			tview.PrintSimple(screen, "[::d]"+label, xpos, y_mid+1)
		}

		if i == 8 {
			break
		}
		x_off += (w - x_off) / (8 - i)
	}
	return x, y, w, h
}

func (tw *CtlTweaker) Draw(screen tcell.Screen) {
	tw.Control.DrawForSubclass(screen, tw)
	x, y, w, h := tw.drawGrid(screen)
	y_mid := y + h/2

	// keep track of positions of already-drawn markers to detect collisions
	model := make([]tcell.Color, w)
	theme := tw.C.Theme()

	for _, color := range tw.C.Selection() {
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

		ch, _, st, _ := screen.GetContent(x+off, y_mid)
		// background color requires special handling
		if b.RGBA == theme[Background].RGBA {
			ch = '▒' //▒ ░ █
			color = tcell.ColorDefault
		} else {
			ch = '█'
		}

		st = st.Foreground(color).Dim(!tw.HasFocus())
		screen.SetContent(x+off, y_mid, ch, nil, st)
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

		tw.C.Adjust(tw.prop, p_adj)
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

func (cs *CtlSelector) Init(C *Console) *CtlSelector {
	cs.Control = new(Control).
		Init(C, "selection").
		SetBlurTitle("[ selection ]", 0).
		//SetBlurBorder(true).
		SetBorderPadding(1, 1, 1, 1)

	cols := len(C.theme)
	cs.model = make([]Lmodel, cols)

	for color := range C.theme {
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
	cols := len(cs.model)

	var csize int
	for csize = 6; csize > 2; csize-- {
		if cols*csize <= w {
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

func (C *Console) Init(t Theme) *Console {
	C.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow)

	C.reset = t
	C.Reset()

	C.csel = new(CtlSelector).Init(C)
	C.AddItem(C.csel, 0, 1, true)

	for _, p := range []cprop8{Y, Cb, Cr, R, G, B} {
		C.AddItem(new(CtlTweaker).Init(C, p), 0, 1, true)
	}
	return C
}

func (C *Console) Adjust(prop cprop8, adj int) Theme {
	return C.theme.Adjust(C.csel.mask, prop, adj).Apply()
}
func (C *Console) Selection() []tcell.Color {
	return C.csel.mask.Iter()
}
func (C *Console) Theme() Theme {
	return C.theme
}
func (C *Console) Reset() *Console {
	C.theme = C.reset.Copy()
	return C
}

func (C *Console) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return C.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {

		f_i := C.focus_i
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
			C.Reset().Theme().Apply()
			return
		default:
			C.GetItem(C.focus_i).InputHandler()(event, setFocus)
			return
		}

		if f_i < 0 {
			f_i += C.GetItemCount()
		}
		f_i %= C.GetItemCount()
		C.focus_i = f_i
		setFocus(C.GetItem(f_i))
	})
}
