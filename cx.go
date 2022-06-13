package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os/exec"
)

func defstyles() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorReset
	tview.Styles.BorderColor = tcell.ColorReset
	tview.Styles.PrimaryTextColor = tcell.ColorReset
	tview.Styles.SecondaryTextColor = tcell.ColorReset
	tview.Styles.TertiaryTextColor = tcell.ColorReset
	tview.Styles.TitleColor = tcell.ColorReset
	tview.Styles.InverseTextColor = tcell.ColorReset
}

func borders(r rune) {
	B := &tview.Borders

	if r == 0 {
		B.TopRight = tview.BoxDrawingsLightArcDownAndLeft
		B.TopLeft = tview.BoxDrawingsLightArcDownAndRight
		B.BottomRight = tview.BoxDrawingsLightArcUpAndLeft
		B.BottomLeft = tview.BoxDrawingsLightArcUpAndRight
		B.TopRightFocus = B.TopRight
		B.TopLeftFocus = B.TopLeft
		B.BottomRightFocus = B.BottomRight
		B.BottomLeftFocus = B.BottomLeft
		B.HorizontalFocus = B.Horizontal
		B.VerticalFocus = B.Vertical
		return
	}

	B.TopRight = r
	B.TopLeft = r
	B.BottomRight = r
	B.BottomLeft = r
	B.TopRightFocus = r
	B.TopLeftFocus = r
	B.BottomRightFocus = r
	B.BottomLeftFocus = r
	B.HorizontalFocus = r
	B.VerticalFocus = r
}

func main() {
	defstyles()
	borders(0)
	var tname string
	themeMap := allThemes()
	app := tview.NewApplication()
	themer := new(Themer).Init(themeMap, func(t string) { tname = t })
	preview := new(Preview).Init() //.Start()
	log.SetOutput(preview.log)
	go func() {
		cmd := exec.Command("bat", "-f", "--theme", "base16", "cx.go")
		txt, err := cmd.Output()
		if err != nil {
			log.Fatalln("bat error", err)
		}
		preview.sample <- txt
		app.Draw()
	}()

	flex := tview.NewFlex().
		AddItem(themer, themer.Width+5, 0, true)
		//AddItem(preview, 0, 3, false)

	themer.Done(func(tname string) {
		preview.Stop()
		theme := themeMap[tname]
		cons := new(Console).Init(theme)
		flex.Clear().
			AddItem(cons, 0, 1, true).
			AddItem(preview, 0, 1, false)

		app.SetFocus(flex)
		preview.Start()
	})

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {

		case tcell.KeyCtrlSpace:
			preview.Rotate()

		case tcell.KeyCtrlL:
			preview.Log()

		case tcell.KeyCtrlB:
			if theme, found := themeMap[tname]; found {
				preview.Bars(theme)
			} else {
				preview.Bars(nil)
			}

		case tcell.KeyCtrlT:
			preview.Start()

		default:
			return event
		}
		return nil
	})

	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
		app.SetAfterDrawFunc(nil)
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
