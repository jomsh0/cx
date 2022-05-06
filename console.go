package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

type Control struct {
	*tview.Box
	C    *Console
	prop cprop8
}

func (c *Control) Init(C *Console, prop cprop8) *Control {
	b := tview.NewBox().
		SetBorder(true).
		SetTitle("[ [::b]" + prop.String() + "[::-] ]").
		SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
			return x + 1, y + 1, w - 2, h - 2
		})

	*c = Control{b, C, prop}
	return c
}

func (c *Control) drawGrid(screen tcell.Screen) (int, int, int, int) {
	if !c.HasFocus() {
		c.SetBorder(false)
	} else {
		c.SetBorder(true)
	}
	c.DrawForSubclass(screen, c)

	x, y, w, h := c.GetInnerRect()
	w -= 1
	x_mid := x + w/2
	y_mid := y + h/2

	midline := strings.Repeat(string(tview.BoxDrawingsLightDoubleDashHorizontal), w)
	tview.PrintSimple(screen, "[::d]"+midline, x, y_mid)

	if !c.HasFocus() {
		tview.PrintSimple(screen, "[::d]"+c.prop.Abbr(), x_mid, y_mid-2)
	}

	mark_dn := string(tview.BoxDrawingsLightDownAndHorizontal)
	mark_up := string(tview.BoxDrawingsLightUpAndHorizontal)

	for i, x_off := 0, 0; ; i++ {
		if i%2 == 0 {
			tview.PrintSimple(screen, "[::d]"+mark_dn, x+x_off, y_mid)
		} else {
			tview.PrintSimple(screen, "[::d]"+mark_up, x+x_off, y_mid)
		}

		if i%2 == 0 && c.HasFocus() {
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

func (c *Control) Draw(screen tcell.Screen) {
	x, y, w, h := c.drawGrid(screen)
	//x_mid := x + w/2
	y_mid := y + h/2
	S := c.C.S

	model := make([]tcell.Color, w)
	for _, color := range S.sel {
		b := S.Theme[color]
		v := b.Access(c.prop)

		off := int(v) * w / 255
		if off == w {
			off--
		}
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
		if b.RGBA == S.Theme[Background].RGBA {
			ch = '▒' //▒ ░
			color = tcell.ColorDefault
		}
		if c.HasFocus() {
			st = st.Background(color)
		} else {
			ch = '░'
			st = st.Dim(true).Foreground(color)
		}
		screen.SetContent(x+off, y_mid, ch, nil, st)
	}
}

func (c *Control) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {

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

		c.adjust(p_adj)
	})
}

func (c *Control) adjust(adj int) {
	c.C.S.Adjust(c.prop, adj)
	c.C.S.SetTheme()
}

type Console struct {
	*tview.Flex
	S       Selection
	reset   Theme
	focus_i int
}

var props = []cprop8{Y, U, V, R, G, B}

type Selector struct {
	*tview.Table
	C *Console
}

func (s *Selector) Init(C *Console) *Selector {
	T := tview.NewTable().SetSelectable(true, true)
	for c := Black; c <= BrightWhite; c++ {
		T.SetCell(0, int(c-Black), tview.NewTableCell(cNames[c].abbr).SetExpansion(1).SetAlign(tview.AlignCenter))
		T.SetCell(1, int(c-Black), &tview.TableCell{Text: "", BackgroundColor: c})
	}
	T.SetTitle("Selection").
		SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
			return x + 1, y + 1, w - 2, h - 2
		})

	*s = Selector{T, C}
	return s
}

func (s *Selector) Draw(screen tcell.Screen) {
	if !s.HasFocus() {
		s.SetBorder(false)
	} else {
		s.SetBorder(true)
	}
	s.DrawForSubclass(screen, s)

	s.Table.Draw(screen)

	// x, y, w, h := s.GetInnerRect()
	// y_mid := y + h/2

	// for i, c := range s.C.S.sel {
	// 	if i > w {
	// 		break
	// 	}
	// 	st := tcell.StyleDefault
	// 	switch c {
	// 	case Foreground:
	// 		st = st.Foreground(tcell.ColorDefault).Reverse(true)
	// 	case Background:
	// 		st = st.Background(tcell.ColorDefault)
	// 	default:
	// 		st = st.Background(c)
	// 	}
	// 	for j := 0; j < 3; j++ {
	// 		screen.SetContent(x+3*i+j, y_mid, ' ', nil, st)
	// 	}

	// 	str := fmt.Sprintf(" %.3s", cNames[c].abbr)
	// 	tview.PrintSimple(screen, str, x+3*i, y_mid-1)
	// }
}

func (C *Console) Init(t Theme) *Console {
	C.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow)
	C.reset = t
	C.Reset()
	C.AddItem(new(Selector).Init(C), 0, 1, true)
	for _, p := range props {
		C.AddItem(new(Control).Init(C, p), 0, 1, true)
	}
	return C
}

func (C *Console) Reset() *Console {
	C.S = C.reset.Copy().Select(^uint64(0))
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
			C.Reset()
			C.S.SetTheme()
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
