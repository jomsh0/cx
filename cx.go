package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	"os/exec"
)

func bat() ([]byte, error) {
	cmd := exec.Command("bat", "-f", "--theme", "base16", "cx.go")
	return cmd.Output()
}

func borders() {
	B := &tview.Borders
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
}

func main() {
	borders()
	themeMap := allThemes()
	app := tview.NewApplication()
	themer := NewThemer().init(themeMap)
	preview := NewPreview().Start()
	log.SetOutput(preview)
	go func() {
		txt, err := bat()
		if err != nil {
			log.Fatalln("bat error", err)
		}
		preview.sample <- txt
		app.Draw()
	}()

	flex := tview.NewFlex().
		AddItem(themer, themer.Width+5, 0, true).
		AddItem(preview, 0, 3, false)

	themer.Done(func(theme string) {
		preview.Cancel()
		model := Model{}.New(themeMap[theme])
		cons := new(Console).Init(model)
		flex.Clear().
			AddItem(cons, 0, 1, true).
			AddItem(preview, 0, 1, false)
		app.SetFocus(cons)
	})

	app.SetAfterDrawFunc(func(screen tcell.Screen) {
		screen.SetCursorStyle(tcell.CursorStyleSteadyBlock)
		app.SetAfterDrawFunc(nil)
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
