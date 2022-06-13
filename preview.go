package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"log"
	//"strings"
	"time"
)

const (
	PM_LOG int = 0 + iota
	PM_BARS
	PM_RASTER
	PM_SIZE
)

type Cell struct {
	ch    rune
	style tcell.Style
}

type Raster struct {
	rows    [][]Cell
	cancel  chan bool
	done    chan bool
	pending bool
	reset   bool
}

type Preview struct {
	*tview.Box
	log     *tview.TextView
	text    *tview.TextView
	theme   Theme
	samples [][]byte
	sample  chan []byte
	raster  *Raster
	mode    int
}

func (p *Preview) Init() *Preview {
	p.Box = tview.NewBox().
		SetBorder(true).
		SetBorderAttributes(tcell.AttrDim).
		SetTitle("[::d][ preview ]")

	p.log = tview.NewTextView()
	p.log.Box = p.Box

	p.text = tview.NewTextView().SetDynamicColors(true)
	p.text.Box = p.Box

	p.raster = new(Raster).Init()
	p.sample = make(chan []byte, 1)
	return p
}

func (rs *Raster) Init() *Raster {
	rs.cancel = make(chan bool, 1)
	rs.done = make(chan bool, 1)
	return rs
}

func (rs *Raster) Capture(screen tcell.Screen, view *tview.Box) *Raster {
	x, y, width, height := view.GetInnerRect()
	rows := make([][]Cell, height)

	for i := 0; i < height; i++ {
		row := make([]Cell, width)
		lim := 0
		for j := 0; j < width; j++ {
			ch, _, style, _ := screen.GetContent(x+j, y+i)
			if ch != ' ' {
				lim = j
			}
			row[j] = Cell{ch, style}
		}
		rows[i] = row[:lim+1]
	}
	rs.rows = rows
	return rs
}

// the first X rows appear immediately to get space filled
const START_ROWS = 10

func (rs *Raster) Draw(screen tcell.Screen, view *tview.Box, unitsz, msec int) {
	x, y, _, _ := view.GetInnerRect()
	var tick *time.Ticker

	_draw := func(rows [][]Cell) {
		for i := 0; i < len(rows); i++ {
			row := rows[i]
			for j := 0; j < len(row); j++ {
				screen.SetContent(x+j, y+i, row[j].ch, nil, row[j].style)
				if unitsz == 0 || j%unitsz != 0 || tick == nil {
					continue
				}
				screen.Show()
				select {
				case <-rs.cancel:
					return
				case <-tick.C:
				}
			}
		}
		screen.Show()
	}

	defer func() { rs.done <- true }()
	tick = time.NewTicker(time.Duration(msec) * time.Millisecond)
	defer tick.Stop()

	st_rows := START_ROWS
	if len(rs.rows) < START_ROWS {
		st_rows = len(rs.rows)
	}

	_draw(rs.rows[:st_rows])
	y += st_rows
	_draw(rs.rows[st_rows:])
}

func (p *Preview) Draw(screen tcell.Screen) {
	p.SetBorderPadding(1, 1, 1, 1)

	switch p.mode {
	case PM_LOG:
		p.log.Draw(screen)
	case PM_BARS:
		p.drawBars(screen)
	case PM_RASTER:
	loop:
		for {
			select {
			case s := <-p.sample:
				p.samples = append(p.samples, s)
			case <-p.raster.done:
				p.raster.pending = false
			default:
				break loop
			}
		}
		if p.raster.pending {
			break
		}
		if p.raster.rows == nil && len(p.samples) > 0 {
			p.text.SetText(tview.TranslateANSI(string(p.samples[0])))
			p.text.Draw(screen)
			p.raster.Capture(screen, p.text.Box)
			p.text.Clear()
		}
		if p.raster.reset {
			p.raster.reset = false
			p.raster.pending = true
			p.DrawForSubclass(screen, p)
			go p.raster.Draw(screen, p.text.Box, 30, 20)
		}
	}
}

func (p *Preview) Stop() *Preview {
	p.mode = PM_SIZE
	p.raster.reset = false
	if p.raster.pending {
		p.raster.cancel <- true
		<-p.raster.done
		p.raster.pending = false
	}
	return p
}
func (p *Preview) Start() *Preview {
	p.mode = PM_RASTER
	p.raster.reset = true
	return p
}
func (p *Preview) Rotate() *Preview {
	p.Stop()
	p.mode = (p.mode + 1) % PM_SIZE
	return p
}
func (p *Preview) Bars(theme Theme) *Preview {
	p.Stop()
	p.mode = PM_BARS
	if theme != nil {
		p.theme = theme
	}
	return p
}
func (p *Preview) Log() *Preview {
	p.Stop()
	p.mode = PM_LOG
	log.Println("LOGGING")
	return p
}

var NAME_WIDTH = 15

func (p *Preview) drawBars(screen tcell.Screen) *Preview {
	x, y, w, h := p.GetRect()
	p.SetBorderPadding(0, 0, w/15, w/15)
	x, y, w, h = p.GetInnerRect()
	p.DrawForSubclass(screen, p)

	theme := p.theme
	if theme == nil {
		theme = make(Theme)
		basics := (CMask(0).Colors() | CMask(0).Grays())
		for _, color := range basics.Iter() {
			theme[color] = bColor{}
		}
	}

	count := len(theme)
	var size int
	for size = 5; size > 1; size-- {
		if count*size < h {
			break
		}
	}
	y_off := (h - count*size) / 2

	for color := range theme {
		indx := CMask(0).IndexMod(len(theme), color)
		y_ind := y + y_off + indx*size

		name := cNames[color].string
		if name == "" {
			name = fmt.Sprintf("color%d", indx)
		}

		tview.Print(screen, name, x, y_ind, NAME_WIDTH, tview.AlignRight, Foreground)
		for j := 0; j < size; j++ {
			for i := NAME_WIDTH + 2; i < w; i++ {
				st := tcell.StyleDefault.Background(color).Reverse(color == Foreground)
				screen.SetContent(x+i, y_ind+j, ' ', nil, st)
			}
		}
	}
	return p
}