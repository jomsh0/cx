package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	//"log"
	"time"
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
}

type Preview struct {
	*tview.TextView
	samples [][]byte
	sample  chan []byte
	raster  *Raster
	active  bool
}

func NewPreview() *Preview {
	prv := tview.NewTextView().
		SetDynamicColors(true)
		//SetScrollable(false)
		//SetMaxLines(200)

	prv.SetBorder(true).
		SetBorderPadding(0, 1, 2, 1).
		SetBorderAttributes(tcell.AttrDim).
		SetTitle("Preview")

	return &Preview{
		TextView: prv,
		samples:  nil,
		sample:   make(chan []byte, 1),
		raster:   nil,
		active:   false,
	}
}

func NewRaster(screen tcell.Screen, view *tview.Box) *Raster {
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

	return &Raster{
		rows:    rows,
		cancel:  make(chan bool, 1),
		done:    make(chan bool, 1),
		pending: false,
	}
}

func (r *Raster) Draw(screen tcell.Screen, view *tview.Box, unitsz, msec int) {
	x, y, _, _ := view.GetInnerRect()
	var tick *time.Ticker
	stRows := 10

	_draw := func(rows [][]Cell) {
	outer:
		for i := 0; i < len(rows); i++ {
			row := rows[i]
			for j := 0; j < len(row); j++ {
				screen.SetContent(x+j, y+i, row[j].ch, nil, row[j].style)
				if unitsz == 0 || j%unitsz != 0 || tick == nil {
					continue
				}
				screen.Show()
				select {
				case <-r.cancel:
					break outer
				case <-tick.C:
				}
			}
		}
		screen.Show()
	}

	if len(r.rows) < stRows {
		stRows = len(r.rows)
	}

	_draw(r.rows[:stRows])
	tick = time.NewTicker(time.Duration(msec) * time.Millisecond)
	y += stRows
	_draw(r.rows[stRows:])
	tick.Stop()

	r.done <- true
}

func (p *Preview) Draw(screen tcell.Screen) {
	select {
	case s := <-p.sample:
		p.samples = append(p.samples, s)
	default:
	}

	if !p.active {
		p.TextView.Draw(screen)
		return
	}

	if len(p.samples) > 0 && p.raster == nil {
		p.TextView.SetText(tview.TranslateANSI(string(p.samples[0])))
		p.TextView.Draw(screen)
		p.raster = NewRaster(screen, p.TextView.Box)
		p.TextView.Clear()
	}

	if p.raster == nil {
		return
	}

	select {
	case <-p.raster.done:
		p.raster.pending = false
	default:
	}

	if !p.raster.pending {
		// blanks out content
		p.TextView.DrawForSubclass(screen, p)
		p.raster.pending = true
		go p.raster.Draw(screen, p.TextView.Box, 30, 20)
	}
}

func (p *Preview) Cancel() {
	p.active = false
	if p.raster == nil {
		return
	}
	if p.raster.pending {
		p.raster.cancel <- true
		<-p.raster.done
	}
	p.raster = nil
}

func (p *Preview) Start() *Preview {
	p.active = true
	return p
}