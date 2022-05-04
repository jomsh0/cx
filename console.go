package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	//"log"
	"strings"
)

type Control struct {
	*tview.Box
	C     *Console
	model *Model
	p     Cprop
	vect  Pvect
}

func (c *Control) Init(C *Console, M *Model, p Cprop) *Control {
	b := tview.NewBox().
		SetBorder(true).
		SetTitle("( [::b]" + p.String() + "[::-] )").
		SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
			return x + 1, y + 1, w - 2, h - 2
		})

	*c = Control{b, C, M, p, Pvect{}}
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
		tview.PrintSimple(screen, "[::d]"+string(c.p.Rune()), x_mid, y_mid-2)
	}

	mark_dn := string(tview.BoxDrawingsLightDownAndHorizontal)
	mark_up := string(tview.BoxDrawingsLightUpAndHorizontal)
	for _, x_off := range []int{x, x + w/2, x + w} {
		tview.PrintSimple(screen, "[::d]"+mark_up, x_off, y_mid)
	}

	for i, x_off := 1, 0; i < 8; i++ {
		// not sure I understand why this works but it does
		x_off += (w - x_off) / (8 - i + 1)
		if i == 4 {
			continue
		} else if i%2 == 0 {
			tview.PrintSimple(screen, "[::d]"+mark_up, x+x_off, y_mid)
		} else {
			tview.PrintSimple(screen, "[::d]"+mark_dn, x+x_off, y_mid)
			if c.HasFocus() {
				tview.PrintSimple(screen, "[::d]0x"+fmt.Sprintf("%X", i*0x2), x+x_off-1, y_mid+1)
			}
		}
	}
	return x, y, w, h
}

func (c *Control) Draw(screen tcell.Screen) {
	x, y, w, h := c.drawGrid(screen)
	//x_mid := x + w/2
	y_mid := y + h/2

	c.model.Update(c.p, &c.vect)
	for i, v := range c.vect {
		off := int(v) * w / 0xff
		ch, _, st, _ := screen.GetContent(x+off, y_mid)
		if i%8 == 0 {
			ch = '▒' //░
		}
		if c.HasFocus() {
			st = st.Background(Cindex(i).Color())
		} else {
			ch = '░'
			st = st.Dim(true).Foreground(Cindex(i).Color())
		}
		screen.SetContent(x+off, y_mid, ch, nil, st)
	}
}

func (c *Control) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		mod := event.Modifiers()
		p_adj := 0 //property adjust

		switch event.Key() {
		case tcell.KeyCtrlL, tcell.KeyRight:
			p_adj = 1
		case tcell.KeyCtrlH, tcell.KeyLeft:
			p_adj = -1
		case tcell.KeyRune:
			rune := event.Rune()
			if rune < 'a' {
				mod |= tcell.ModShift
			}
			switch rune {
			case 'L', 'l':
				p_adj = 1
			case 'H', 'h':
				p_adj = -1
			}
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
	c.model.Adjust(c.p, nil, adj)
	c.C.update()
}

func (c *Control) update() {
	c.model.Update(c.p, &c.vect)
}

type Console struct {
	*tview.Flex
	model   *Model
	focus_x int
	props   []Cprop
}

func (C *Console) Init(M *Model) *Console {
	C.props = []Cprop{Y, U, V, R, G, B}
	C.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow)

	for _, p := range C.props {
		C.AddItem(new(Control).Init(C, M, p), 0, 1, true)
	}
	C.focus_x = -1
	C.model = M
	return C
}

func (C *Console) update() {
	for i := 0; i < C.GetItemCount(); i++ {
		ctl := C.GetItem(i).(*Control)
		ctl.update()
	}
	C.model.Theme().SetTheme()
}

func (C *Console) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return C.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		f_x := C.focus_x

		switch event.Key() {
		case tcell.KeyTab, tcell.KeyCtrlJ, tcell.KeyDown:
			f_x++
		case tcell.KeyBacktab, tcell.KeyCtrlK, tcell.KeyUp:
			f_x--
		case tcell.KeyRune:
			switch event.Rune() {
			case 'J', 'j':
				f_x++
			case 'K', 'k':
				f_x--
			}
		}

		if f_x == C.focus_x {
			C.GetItem(f_x).InputHandler()(event, setFocus)
			return
		}
		if f_x < 0 {
			f_x += C.GetItemCount()
		}
		f_x %= C.GetItemCount()
		C.focus_x = f_x
		setFocus(C.GetItem(f_x))
	})
}
