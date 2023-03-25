package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	termutil "github.com/japanoise/tcell-util"
)

var renderCallback func(scr tcell.Screen, sx, sy int)

const EDGE_MIN = 10
const EDGE_MAX = 40

func attemptRender(s tcell.Screen,
	startx, sx, sy int,
	state *State, config *Config) (int, int) {
	cbuf := state.CurBuf()
	y := 0
	idx := cbuf.Sy
	// Cursor position; this is how we handle scrolling
	// Initial cx is high, so we know if the focused component was rendered
	cx, cy := sx*2, y
	for y < sy-2 {
		if idx >= cbuf.Nelems {
			break
		}
		y += cbuf.Elems[idx].render(
			config, s, cbuf, startx, y, idx == cbuf.Cy, &cx, &cy)
		y++
		idx++
	}
	if cbuf.Cy == cbuf.Nelems {
		cx = startx
		cy = y
	}
	return cx, cy
}

func renderModeline(s tcell.Screen, sx, sy int, state *State, config *Config, cbuf *Buffer) {
	for i := 0; i < sx; i++ {
		// Clear the penultimate line for the modeline
		termutil.PrintRuneStyle(s, i, sy-2, ' ', config.Modeline)
		// Clear the last line for the minibuffer
		termutil.PrintRuneStyle(s, i, sy-1, ' ', tcell.StyleDefault)
	}

	// Style indicators
	if cbuf.Sty&StyleItalic != 0 {
		termutil.PrintRuneStyle(s, sx-1, sy-2, 'I', config.Modeline)
	}
	if cbuf.Sty&StyleBold != 0 {
		termutil.PrintRuneStyle(s, sx-2, sy-2, 'B', config.Modeline)
	}
	if cbuf.Sty&StyleUnderline != 0 {
		termutil.PrintRuneStyle(s, sx-3, sy-2, 'U', config.Modeline)
	}

	// Filenames
	termutil.PrintStringStyle(s, 3, sy-2,
		fmt.Sprintf("%s:%s", state.Filename, cbuf.Name),
		config.Modeline)

	// Current buffer modified
	if cbuf.Dirty {
		termutil.PrintRuneStyle(s, 1, sy-2, '*', config.Modeline)
	}

	// Wordcount
	termutil.PrintStringStyle(s, sx-10, sy-2,
		fmt.Sprintf("%d", cbuf.Words),
		config.Modeline)

	// Message in the minibuffer
	termutil.PrintString(s, 0, sy-1, state.msg)

	if sx < config.Width {
		termutil.PrintStringStyle(
			s, 0, sy-1, "Term too narrow", config.Warning)
	}
}

func drawStringIn(s tcell.Screen, x, y, edge int, str string, sty tcell.Style) {
	width := 0
	for _, ru := range str {
		rw := termutil.Runewidth(ru)
		if x+width+rw >= edge {
			break
		}
		termutil.PrintRuneStyle(s, x+width, y, ru, sty)
		width += rw
	}
}

func (d *dired) render(s tcell.Screen, state *State,
	starty, edge, nest int, sty tcell.Style) int {
	y := starty
	tab := nest * 3
	if nest >= 0 {
		for i := 0; i <= nest; i++ {
			termutil.PrintRune(
				s, i*3, y, tcell.RuneVLine)
		}
		termutil.PrintRune(s, tab, y, tcell.RuneLTee)
		termutil.PrintRune(s, tab+1, y, tcell.RuneHLine)
		drawStringIn(s, tab+3, y, edge, d.name, sty)
		y++
		for i := 0; i <= nest; i++ {
			termutil.PrintRune(
				s, i*3, y, tcell.RuneVLine)
		}
	}
	for _, dir := range d.subdirs {
		if dir.open {
			y = dir.render(s, state, y, edge, nest+1, sty)
		} else {
			termutil.PrintRune(s, tab+3, y, tcell.RuneLTee)
			termutil.PrintRune(s, tab+4, y, tcell.RuneHLine)
			drawStringIn(s, tab+6, y, edge, dir.name, sty)
			if nest >= 0 {
				for i := 0; i <= nest; i++ {
					termutil.PrintRune(
						s, i*3, y, tcell.RuneVLine)
				}
			}
			y++
		}
	}
	for _, buf := range d.files {
		termutil.PrintRune(s, tab+3, y, tcell.RuneLTee)
		termutil.PrintRune(s, tab+4, y, tcell.RuneHLine)
		if nest >= 0 {
			idx := strings.LastIndexByte(state.Bufs[buf].Name, '/')
			drawStringIn(s, tab+6, y, edge,
				state.Bufs[buf].Name[idx+1:],
				tcell.StyleDefault)
			for i := 0; i <= nest; i++ {
				termutil.PrintRune(
					s, i*3, y, tcell.RuneVLine)
			}
		} else {
			drawStringIn(s, tab+6, y, edge,
				state.Bufs[buf].Name, tcell.StyleDefault)
		}
		y++
	}
	termutil.PrintRune(s, tab+3, y-1, tcell.RuneLLCorner)

	return y
}

func renderSidebar(s tcell.Screen, startx, sy int,
	state *State, config *Config) {
	edge := EDGE_MIN
	if startx > EDGE_MIN {
		edge = startx - 1
	}
	if edge > EDGE_MAX {
		edge = EDGE_MAX
	}
	for y := 0; y < sy-2; y++ {
		for x := 0; x < edge; x++ {
			termutil.PrintRune(s, x, y, ' ')
		}
		termutil.PrintRune(s, edge, y, tcell.RuneVLine)
	}
	state.root.render(s, state, 0, edge, -1, config.Dired)
}

func render(s tcell.Screen, sx, sy int, state *State, config *Config) {
	cbuf := state.CurBuf()
	var cx, cy int

	if cbuf.Sy > cbuf.Cy {
		cbuf.Sy = cbuf.Cy
	}

	startx := (sx - config.Width) / 2

	// Render the buffer. Uses the cursor pos to judge scrolling
	cx, cy = attemptRender(s, startx, sx, sy, state, config)
	for cx > sx || cy > sy-3 {
		s.Clear()
		if cbuf.Sy == cbuf.Cy {
			cbuf.Sey++
		} else {
			cbuf.Sy += 5
		}
		if cbuf.Sy > cbuf.Cy {
			cbuf.Sy = cbuf.Cy
		}
		cx, cy = attemptRender(s, startx, sx, sy, state, config)
	}
	for cy < 0 {
		s.Clear()
		cbuf.Sey--
		cx, cy = attemptRender(s, startx, sx, sy, state, config)
	}

	if state.sidebar {
		renderSidebar(s, startx, sy, state, config)
	}

	renderModeline(s, sx, sy, state, config, cbuf)

	s.ShowCursor(cx, cy)
}

func (d *dired) click(state *State, starty, edge, nest int,
	ev *tcell.EventMouse) (int, bool) {
	y := starty
	_, cy := ev.Position()
	if nest >= 0 {
		if cy == y {
			d.open = !d.open
			return y, true
		}
		y++
	}
	for _, dir := range d.subdirs {
		if dir.open {
			dy, found := dir.click(state, y, edge, nest+1, ev)
			if found {
				return y, true
			}
			y = dy
		} else {
			if cy == y {
				dir.open = true
				return y, true
			}
			y++
		}
	}
	for _, buf := range d.files {
		if cy == y {
			state.changeBuffer(buf)
			return y, true
		}
		y++
	}

	return y, false
}

func clickSidebar(state *State, sy, edge int, ev *tcell.EventMouse) bool {
	_, found := state.root.click(state, 0, edge, -1, ev)
	return found
}

// It's rather like a render in reverse, hence why it's here
func click(state *State, config *Config,
	sx, sy int, ev *tcell.EventMouse) {
	cbuf := state.CurBuf()
	cx, cy := ev.Position()

	if cbuf.Sy > cbuf.Cy {
		cbuf.Sy = cbuf.Cy
	}

	startx := (sx - config.Width) / 2

	if state.sidebar {
		edge := EDGE_MIN
		if startx > EDGE_MIN {
			edge = startx - 1
		}
		if edge > EDGE_MAX {
			edge = EDGE_MAX
		}
		if edge > cx {
			if clickSidebar(state, sy, edge, ev) {
				return
			}
		}
	}

	y := 0
	idx := cbuf.Sy

	for y < sy-2 {
		if idx >= cbuf.Nelems {
			break
		}
		dy, found := cbuf.Elems[idx].click(
			config, cbuf, startx, y, idx, ev)
		if found {
			return
		}
		y += dy
		if cy == y {
			cbuf.Cy = idx
			cbuf.Elems[cbuf.Cy].endOf(config, cbuf, config.Width)
			return
		}
		y++
		idx++
	}
	if cbuf.Cy == cbuf.Nelems {
		cx = startx
		cy = y
	}
}
