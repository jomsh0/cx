package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	//"log"
	"sort"
)

type Themer struct {
	*tview.Flex

	tm       ThemeMap
	cb       func(string)
	list     *tview.List
	refList  *tview.List
	search   *tview.InputField
	Width    int
	doneFunc func(string)
}

func (th *Themer) Init(tm ThemeMap, cb func(string)) *Themer {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetMainTextColor(tcell.ColorGray).
		SetHighlightFullLine(true).
		SetWrapAround(false)
	refList := tview.NewList()

	search := tview.NewInputField().
		SetFieldBackgroundColor(list.GetBackgroundColor())

	flex := tview.NewFlex().
		AddItem(list, 0, 1, false).
		AddItem(search, 2, 0, true).
		SetDirection(tview.FlexRow)

	search.SetBorderPadding(1, 0, 0, 0)
	flex.SetBorderPadding(1, 1, 1, 1)

	*th = Themer{Flex: flex, list: list, refList: refList, search: search, cb: cb}

	search.SetChangedFunc(func(text string) {
		if text == "" {
			th.resetList(nil)
		} else if matches := refList.FindItems(text, "", false, true); matches != nil {
			th.resetList(matches)
		} else {
			th.resetList([]int{})
		}
	})

	return th.init(tm)
}

func simpleAdd(list *tview.List, theme string) {
	list.AddItem(theme, "", 0, nil)
}

func (t *Themer) init(tm ThemeMap) *Themer {
	if tm == nil {
		tm = allThemes()
	}
	themes := make([]string, 0, len(tm))
	for tname := range tm {
		themes = append(themes, tname)
	}
	sort.Strings(themes)

	for _, tname := range themes {
		simpleAdd(t.refList, tname)
		if len(tname) > t.Width {
			t.Width = len(tname)
		}
	}

	t.tm = tm
	t.resetList(nil)
	return t
}

func (t *Themer) resetList(inds []int) {
	t.list.Clear()

	if inds == nil {
		for i := 0; i < t.refList.GetItemCount(); i++ {
			tname, _ := t.refList.GetItemText(i)
			simpleAdd(t.list, tname)
		}
		return
	}
	for _, i := range inds {
		tname, _ := t.refList.GetItemText(i)
		simpleAdd(t.list, tname)
	}
}

func (t *Themer) GetTheme() (string, bool) {
	if t.list.GetItemCount() == 0 {
		return "", false
	}
	theme, _ := t.list.GetItemText(t.list.GetCurrentItem())
	return theme, true
}

func (t *Themer) done() {
	theme, ok := t.GetTheme()
	if !ok {
		return
	}
	t.doneFunc(theme)
}

func (t *Themer) Done(doneFunc func(string)) *Themer {
	t.doneFunc = doneFunc
	return t
}

func (t *Themer) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	clearOr := func(f func()) {
		if t.search.GetText() == "" {
			f()
		} else {
			t.search.SetText("")
		}
	}

	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		trigger := func(key tcell.Key) {
			if key == 0 {
				t.list.InputHandler()(event, setFocus)
			} else {
				t.list.InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), setFocus)
			}

			if theme, ok := t.GetTheme(); ok {
				t.tm.Apply(theme)
				if t.cb != nil {
					t.cb(theme)
				}
			}
		}

		switch event.Key() {
		case tcell.KeyCtrlJ, tcell.KeyCtrlN:
			trigger(tcell.KeyDown)
		case tcell.KeyCtrlK, tcell.KeyCtrlP:
			trigger(tcell.KeyUp)
		case tcell.KeyDown, tcell.KeyUp:
			trigger(0)
		case tcell.KeyEnter:
			trigger(0)
			t.done()
		case tcell.KeyEsc:
			clearOr(t.done)
		default:
			t.search.InputHandler()(event, setFocus)
		}
	})
}
