package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	termutil "github.com/japanoise/tcell-util"
	"golang.org/x/exp/slices"
)

type Element interface {
	// render draws the element to the screen at startx, starty and returns the lines needed
	render(c *Config, s tcell.Screen, b *Buffer, startx, starty int,
		focus bool, cx, cy *int) int
	// click is called when the screen is clicked. It returns the lines needed to draw the element and true if the element was clicked
	click(c *Config, b *Buffer, startx, starty, idxOf int,
		ev *tcell.EventMouse) (int, bool)
	keyEvent(c *Config, b *Buffer, key string)
	startOf(c *Config, b *Buffer, columnHint int)
	endOf(c *Config, b *Buffer, columnHint int)
	serialize() string
}

// Header is an HTML-style header
type Header struct {
	Level   int
	Data    []rune
	Datalen int
}

func (h *Header) click(c *Config, b *Buffer, startx, starty, idxOf int,
	ev *tcell.EventMouse) (int, bool) {
	anchorx := startx + ((c.Width) / 2) - ((h.Datalen - 1) / 2)
	if anchorx < startx+2 {
		anchorx = startx + 2
	}

	ret := false
	cx, cy := ev.Position()

	if starty == cy {
		ret = true
		b.Cy = idxOf
		b.Zeroes()
	}

	sx := 0
	for i := 0; sx < c.Width && i < h.Datalen; i++ {
		if cx == anchorx+sx {
			b.Cex = i
		}
		sx += termutil.Runewidth(h.Data[i])
	}
	if cx >= anchorx+sx {
		b.Cex = h.Datalen
	}

	return 1, ret
}

func (h *Header) render(c *Config, s tcell.Screen, b *Buffer,
	startx, starty int, focus bool, cx, cy *int) int {

	termutil.PrintRuneStyle(s, startx, starty, '0'+rune(h.Level),
		c.Header)

	anchorx := startx + ((c.Width) / 2) - ((h.Datalen - 1) / 2)
	if anchorx < startx+2 {
		anchorx = startx + 2
	}

	if focus {
		*cx = anchorx + h.Datalen
		*cy = starty
	}

	sx := 0
	for i := 0; sx < c.Width && i < h.Datalen; i++ {
		termutil.PrintRuneStyle(
			s, anchorx+sx, starty, h.Data[i], c.Header)
		if focus && b.Cex == i {
			*cx = anchorx + sx
		}
		sx += termutil.Runewidth(h.Data[i])
	}

	return 1
}

func (h *Header) startOf(c *Config, b *Buffer, columnHint int) {
	b.Zeroes()
	if columnHint <= h.Datalen {
		b.Cex = columnHint
	} else {
		b.Cex = h.Datalen
	}
}

func (h *Header) endOf(c *Config, b *Buffer, columnHint int) {
	b.Zeroes()
	if columnHint <= h.Datalen {
		b.Cex = columnHint
	} else {
		b.Cex = h.Datalen
	}
}

func (h *Header) keyEvent(c *Config, b *Buffer, key string) {
	if utf8.RuneCountInString(key) == 1 {
		ru, _ := utf8.DecodeRuneInString(key)
		if b.Cex == h.Datalen {
			h.Data = append(h.Data, ru)
		} else {
			h.Data = slices.Insert(h.Data, b.Cex, ru)
		}
		b.Cex++
		h.Datalen++
	} else {
		switch key {
		case "C-p", "UP":
			b.PrevElem(c, b.Cex)
		case "C-n", "DOWN":
			b.NextElem(c, b.Cex)
		case "C-f", "RIGHT":
			if b.Cex < h.Datalen {
				b.Cex++
			} else {
				b.NextElem(c, 0)
			}
		case "C-b", "LEFT":
			if b.Cex > 0 {
				b.Cex--
			} else {
				b.PrevElem(c, c.Width)
			}
		case "M-b":
			if b.Cex == 0 {
				b.PrevElem(c, c.Width)
				break
			}
			for b.Cex > 0 && h.Data[b.Cex-1] == ' ' {
				b.Cex--
			}
			if b.Cex == 0 {
				b.PrevElem(c, c.Width)
				break
			}
			for b.Cex > 0 && h.Data[b.Cex-1] != ' ' {
				b.Cex--
			}
		case "M-f":
			if b.Cex >= h.Datalen {
				b.NextElem(c, 0)
			}
			for b.Cex < h.Datalen && h.Data[b.Cex] == ' ' {
				b.Cex++
			}
			if b.Cex >= h.Datalen {
				b.NextElem(c, 0)
			}
			for b.Cex < h.Datalen && h.Data[b.Cex] != ' ' {
				b.Cex++
			}
		case "C-d", "deletechar":
			if b.Cex < h.Datalen {
				b.Dirty = true
				h.Data = slices.Delete(h.Data, b.Cex, b.Cex+1)
				h.Datalen--
			}
		case "C-h", "DEL":
			if b.Cex > 0 {
				b.Dirty = true
				h.Data = slices.Delete(h.Data, b.Cex-1, b.Cex)
				h.Datalen--
				b.Cex--
			}
		case "C-u":
			if b.Cex > 0 {
				b.Dirty = true
				h.Data = slices.Delete(h.Data, 0, b.Cex)
				h.Datalen = len(h.Data)
				b.Cex = 0
			}
		case "C-k":
			if b.Cex < h.Datalen {
				b.Dirty = true
				h.Data = slices.Delete(h.Data, b.Cex, h.Datalen)
				h.Datalen = len(h.Data)
			}
		case "C-a", "Home":
			b.Cex = 0
		case "C-e", "M-e", "End":
			b.Cex = h.Datalen
		case "RET":
			b.Cy++
			b.InsertElem(&Paragraph{})
		}
	}
}

func (h *Header) serialize() string {
	var b strings.Builder
	b.WriteRune('h')
	fmt.Fprintf(&b, "%d", h.Level)
	b.WriteString(string(h.Data))
	return b.String()
}

func deserializeHeader(s string) *Header {
	ret := Header{}
	ret.Level = int(s[1] - '0')
	ret.Data = []rune(s[2:])
	ret.Datalen = len(ret.Data)
	return &ret
}

// Rule is a horizontal rule
type Rule struct{}

func (r *Rule) click(c *Config, b *Buffer, startx, starty, idxOf int,
	ev *tcell.EventMouse) (int, bool) {
	cx, cy := ev.Position()
	if cy == starty {
		b.Zeroes()
		b.Cy = idxOf
		b.Cex = cx - startx
		if b.Cex < 0 {
			b.Cex = 0
		} else if b.Cex > c.Width {
			b.Cex = c.Width
		}
		return 1, true
	}
	return 1, false
}

func (r *Rule) render(c *Config, s tcell.Screen, b *Buffer, startx, starty int,
	focus bool, cx, cy *int) int {
	if focus {
		*cx = startx + b.Cex
		*cy = starty
	}
	for i := 0; i < c.Width; i++ {
		termutil.PrintRuneStyle(
			s, startx+i, starty, tcell.RuneS7, c.Rule)
	}
	return 1
}

func (r *Rule) keyEvent(c *Config, b *Buffer, key string) {
	switch key {
	case "C-a":
		b.Cex = 0
	case "C-e", "M-e":
		b.Cex = c.Width
	case "C-p", "UP", "M-b":
		b.PrevElem(c, b.Cex)
	case "C-n", "DOWN", "M-f":
		b.NextElem(c, b.Cex)
	case "C-d", "deletechar", "C-h", "DEL":
		b.DeleteElem(c)
	case "RET":
		b.Cy++
		b.InsertElem(&Paragraph{})
	}
}

func (r *Rule) startOf(c *Config, b *Buffer, columnHint int) {
	b.Zeroes()
	b.Cex = columnHint
}

func (r *Rule) endOf(c *Config, b *Buffer, columnHint int) {
	b.Zeroes()
	b.Cex = columnHint
}

func (r *Rule) serialize() string {
	return "r"
}

func deserializeRule(s string) *Rule {
	return &Rule{}
}

// Paragraph is a list of words
type Paragraph struct {
	Words  []Word
	Nwords int
	// Cached value
	Column int
}

// A word is exactly what it sounds like
type Word struct {
	Data   []rune
	Styles []Style
	Len    int
	Width  int
	// Cached
	Line int
}

func (w *Word) Insert(at int, ru rune, sty Style) {
	if at == w.Len {
		w.Data = append(w.Data, ru)
		w.Styles = append(w.Styles, sty)
	} else {
		w.Data = slices.Insert(w.Data, at, ru)
		w.Styles = slices.Insert(w.Styles, at, sty)
	}
	w.Len++
	w.Width += termutil.Runewidth(ru)
}

func (w *Word) Set(data []rune, styles []Style) {
	w.Data = data
	w.Styles = styles
	w.Len = len(w.Data)
	w.Width = termutil.RunewidthStr(string(w.Data))
}

func (w *Word) Append(aw *Word) {
	w.Data = append(w.Data, aw.Data...)
	w.Styles = append(w.Styles, aw.Styles...)
	w.Len = len(w.Data)
	w.Width = termutil.RunewidthStr(string(w.Data))
}

func (w *Word) Delete(i, j int) {
	rewidth := true
	if (j - i) == 1 {
		rewidth = false
		w.Width -= termutil.Runewidth(w.Data[i])
	}
	w.Data = slices.Delete(w.Data, i, j)
	w.Styles = slices.Delete(w.Styles, i, j)
	w.Len -= j - i
	if rewidth {
		w.Width = termutil.RunewidthStr(string(w.Data))
	}
}

func (p *Paragraph) click(c *Config, b *Buffer, startx, starty, idxOf int,
	ev *tcell.EventMouse) (int, bool) {
	cx, cy := ev.Position()
	sx, sy := 0, 0-b.Sey
	ret := false
	for idx, word := range p.Words {
		if sx+word.Width > c.Width {
			if sy == cy && !ret {
				sey := b.Sey
				b.Zeroes()
				b.Sey = sey
				b.Cei = idx - 1
				b.Cey = sy
				b.Cex = p.Words[b.Cei].Len
				ret = true
			}
			sy++
			sx = 0
		}
		if sy == cy && cx < startx && !ret {
			sey := b.Sey
			b.Zeroes()
			b.Sey = sey
			b.Cei = idx
			b.Cey = sy
			b.Cex = 0
			ret = true
		}
		for jdx, ru := range word.Data {
			if sy == cy && cx == startx+sx {
				sey := b.Sey
				b.Zeroes()
				b.Sey = sey
				b.Cei = idx
				b.Cey = sy
				b.Cex = jdx
				ret = true
			}
			sx += termutil.Runewidth(ru)
		}
		if idx != p.Nwords-1 {
			if sx+1 > c.Width {
				if cy == sy && !ret {
					sey := b.Sey
					b.Zeroes()
					b.Sey = sey
					b.Cei = idx
					b.Cey = sy
					b.Cex = p.Words[b.Cei].Len
					ret = true
				}
				sy++
				sx = 0
			} else {
				if cy == sy && cx == startx+sx && !ret {
					sey := b.Sey
					b.Zeroes()
					b.Sey = sey
					b.Cei = idx
					b.Cey = sy
					b.Cex = p.Words[b.Cei].Len
					ret = true
				}
				sx++
			}
		}
	}
	if cy == sy && !ret {
		sey := b.Sey
		b.Zeroes()
		b.Sey = sey
		b.Cei = p.Nwords - 1
		b.Cey = sy
		b.Cex = p.Words[b.Cei].Len
		ret = true
	}
	return 1 + sy, ret
}

func (p *Paragraph) render(c *Config, s tcell.Screen, b *Buffer,
	startx, starty int,
	focus bool, cx, cy *int) int {
	if focus {
		*cx = startx
		*cy = starty
	}
	sx, sy := 0, 0-b.Sey
	for idx, word := range p.Words {
		if sx+word.Width > c.Width {
			sy++
			sx = 0
		}
		p.Words[idx].Line = sy + b.Sey
		if focus && idx == b.Cei {
			*cx = startx + sx
			*cy = starty + sy
		}
		for jdx, ru := range word.Data {
			if focus && idx == b.Cei && jdx == b.Cex {
				*cx = startx + sx
				*cy = starty + sy
				b.Cey = sy
			}
			if sy >= 0 {
				termutil.PrintRuneStyle(s, startx+sx, starty+sy, ru,
					c.style2style(word.Styles[jdx]))
			}
			sx += termutil.Runewidth(ru)
		}
		if focus && idx == b.Cei && b.Cex == word.Len {
			*cx = startx + sx
			*cy = starty + sy
		}
		if idx != p.Nwords-1 {
			if sx+1 > c.Width {
				sy++
				sx = 0
			} else {
				sx++
			}
		}
	}
	if focus {
		p.Column = *cx - startx
	}
	return 1 + sy
}

func (p *Paragraph) tryColumn(b *Buffer, target int) {
	if p.Nwords == 0 {
		return
	}
	line := p.Words[b.Cei].Line
	// Find the start of the line first
	candidate := b.Cei
	for i := b.Cei; i >= 0; i-- {
		if p.Words[i].Line != line {
			break
		} else {
			candidate = i
		}
	}
	if target == 0 {
		b.Cei = candidate
		b.Cex = 0
		return
	}
	// Now march words until either we leave the line,
	// reach the end of the words array, or reach our target.
	column := 0
	b.Cei = candidate
	for candidate < p.Nwords {
		if p.Words[candidate].Line != line {
			// We've left the line
			break
		} else if column+p.Words[candidate].Width >= target {
			b.Cei = candidate
			b.Cex = 0
			for column < target {
				column += termutil.Runewidth(
					p.Words[candidate].Data[b.Cex])
				b.Cex++
			}
			break
		} else {
			b.Cei = candidate
			b.Cex = p.Words[candidate].Len
			column += b.Cex
		}
		candidate++
		column++
	}
}

func (p *Paragraph) keyEvent(c *Config, b *Buffer, key string) {
	switch key {
	case "C-a", "HOME":
		p.tryColumn(b, 0)

	case "C-e", "END":
		p.tryColumn(b, c.Width)

	case "C-k":
		if p.Nwords == 0 || b.Cei == p.Nwords {
			break
		}

		cei := b.Cei
		cex := b.Cex
		p.tryColumn(b, c.Width)
		if cei == b.Cei && cex == b.Cex {
			// Nothing to do, we're already at EoL
			break
		} else if cei == b.Cei {
			// Just kill to end of word
			p.Words[b.Cei].Delete(cex, b.Cex)
			b.Cex = cex
			b.Dirty = true
		} else {
			// Kill to end of word first
			p.Words[cei].Delete(cex, p.Words[cei].Len)
			// Now delete other words
			p.Words = slices.Delete(p.Words, cei+1, b.Cei+1)
			p.Nwords = len(p.Words)
			// And reset position
			b.Cei = cei
			b.Cex = cex
			b.Dirty = true
		}

	case "C-u":
		if p.Nwords == 0 || (b.Cei == 0 && b.Cex == 0) {
			break
		}

		cei := b.Cei
		cex := b.Cex
		p.tryColumn(b, 0)
		if cei == b.Cei && cex == b.Cex {
			// Nothing to do, we're already at BoL
			break
		} else if cei == b.Cei {
			// Kill to beg of word
			p.Words[b.Cei].Delete(0, cex)
			b.Cex = 0
			b.Dirty = true
		} else {
			// Kill to beg of word first
			p.Words[cei].Delete(0, cex)
			// Clear words
			p.Words = slices.Delete(p.Words, b.Cei, cei)
			p.Nwords = len(p.Words)
			// Set position
			b.Cei = 0
			b.Cex = 0
			b.Dirty = true
		}

	case "M-e":
		if p.Nwords != 0 {
			b.Cei = p.Nwords - 1
			b.Cex = p.Words[b.Cei].Len
		}

	case "C-p", "UP":
		if p.Nwords == 0 {
			b.PrevElem(c, 0)
			return
		} else if p.Nwords <= b.Cei {
			b.Cei = p.Nwords - 1
		}

		line := p.Words[b.Cei].Line
		if line == 0 {
			b.PrevElem(c, p.Column)
			return
		}
		for p.Words[b.Cei].Line == line {
			b.Cei--
		}
		p.tryColumn(b, p.Column)

	case "C-n", "DOWN":
		if p.Nwords == 0 {
			b.NextElem(c, 0)
		} else if b.Cei >= p.Nwords {
			b.NextElem(c, p.Column)
		} else {
			line := p.Words[b.Cei].Line
			b.Cei++
			for b.Cei < p.Nwords {
				if p.Words[b.Cei].Line > line {
					p.tryColumn(b, p.Column)
					return
				}
				b.Cei++
			}
			b.NextElem(c, p.Column)
		}

	case "C-b", "LEFT":
		if b.Cex == 0 {
			if b.Cei == 0 {
				b.PrevElem(c, c.Width)
			} else {
				b.Cei--
				b.Cex = p.Words[b.Cei].Len
			}
		} else {
			b.Cex--
		}

	case "C-f", "RIGHT":
		if p.Nwords == 0 {
			b.NextElem(c, 0)
		} else if b.Cex == p.Words[b.Cei].Len {
			b.Cei++
			b.Cex = 0
			if b.Cei == p.Nwords {
				b.NextElem(c, 0)
			}
		} else {
			b.Cex++
		}

	case "M-f":
		if p.Nwords == 0 || b.Cei == p.Nwords {
			b.NextElem(c, 0)
		} else if b.Cex < p.Words[b.Cei].Len {
			b.Cex = p.Words[b.Cei].Len
		} else if b.Cei == p.Nwords-1 {
			b.NextElem(c, 0)
		} else {
			b.Cei++
			b.Cex = p.Words[b.Cei].Len
		}

	case "M-b":
		if p.Nwords == 0 || (b.Cei == 0 && b.Cex == 0) {
			b.PrevElem(c, c.Width)
		} else if b.Cex == 0 {
			b.Cei--
			b.Cex = 0
		} else {
			b.Cex = 0
		}

	case "C-d", "deletechar":
		if p.Nwords == 0 {
			b.DeleteElem(c)
		} else if (b.Cei == p.Nwords ||
			(b.Cei == p.Nwords-1 && b.Cex == p.Words[b.Cei].Len)) &&
			b.Cy != b.Nelems-1 {
			oldcei := b.Cei
			oldcey := b.Cey
			oldsey := b.Sey
			oldcex := b.Cex
			switch elem := b.Elems[b.Cy+1].(type) {
			case *Paragraph:
				if elem.Nwords == 0 {
					break
				}
				p.Words = append(p.Words, elem.Words...)
				p.Nwords += elem.Nwords
			case *Rule:
				break
			default:
				return
			}
			b.NextElem(c, 0)
			b.DeleteElem(c)
			b.Cei = oldcei
			b.Cex = oldcex
			b.Cey = oldcey
			b.Sey = oldsey
			b.Cy--
		} else if b.Cei < p.Nwords-1 && b.Cex == p.Words[b.Cei].Len {
			b.Dirty = true
			p.Words[b.Cei].Append(&p.Words[b.Cei+1])
			p.Words = slices.Delete(p.Words, b.Cei+1, b.Cei+2)
			p.Nwords--
			b.Words--
		} else if b.Cei < p.Nwords && p.Words[b.Cei].Len > 0 {
			b.Dirty = true
			p.Words[b.Cei].Delete(b.Cex, b.Cex+1)
		}

	case "C-h", "DEL":
		if b.Cei == 0 && b.Cex == 0 {
			if b.Cy == 0 {
				// Beginning of buffer
				return
			}
			switch elem := b.Elems[b.Cy-1].(type) {
			case *Paragraph:
				if elem.Nwords == 0 {
					b.PrevElem(c, 0)
					b.DeleteElem(c)
					return
				} else if p.Nwords == 0 {
					b.DeleteElem(c)
					b.Cei = elem.Nwords - 1
					b.Cex = elem.Words[b.Cei].Len
					b.Cy--
					return
				}
				newcei := elem.Nwords
				elem.Words = append(elem.Words, p.Words...)
				elem.Nwords += p.Nwords
				b.DeleteElem(c)
				b.Cei = newcei
				b.Cex = 0
				b.Cy--
			case *Rule:
				b.Cy--
				b.DeleteElem(c)
			}
		} else if b.Cex == 0 {
			b.Dirty = true
			b.Cex = p.Words[b.Cei-1].Len
			p.Words[b.Cei-1].Append(&p.Words[b.Cei])
			p.Words = slices.Delete(p.Words, b.Cei, b.Cei+1)
			p.Nwords--
			b.Words--
			b.Cei--
		} else {
			b.Dirty = true
			p.Words[b.Cei].Delete(b.Cex-1, b.Cex)
			b.Cex--
		}

	case "RET":
		b.Cy++
		b.InsertElem(&Paragraph{})
		b.Zeroes()

	case " ":
		if b.Cei < p.Nwords && p.Words[b.Cei].Len > 0 {
			b.Dirty = true
			p.Words = slices.Insert(p.Words, b.Cei+1, Word{})
			p.Nwords++
			b.Words++
			if b.Cex < p.Words[b.Cei].Len {
				p.Words[b.Cei+1].Set(p.Words[b.Cei].Data[b.Cex:],
					p.Words[b.Cei].Styles[b.Cex:])
				p.Words[b.Cei].Delete(b.Cex, p.Words[b.Cei].Len)
			}
			b.Cei++
			b.Cex = 0
		}

	default:
		if utf8.RuneCountInString(key) == 1 {
			b.Dirty = true
			ru, _ := utf8.DecodeRuneInString(key)
			if b.Cei >= p.Nwords {
				p.Nwords++
				p.Words = append(p.Words, Word{})
			}
			p.Words[b.Cei].Insert(b.Cex, ru, b.Sty)
			b.Cex++
		}
	}
}

func (p *Paragraph) startOf(c *Config, b *Buffer, columnHint int) {
	b.Zeroes()
	if p.Nwords == 0 {
		return
	}
	p.tryColumn(b, columnHint)
}

func (p *Paragraph) endOf(c *Config, b *Buffer, columnHint int) {
	b.Zeroes()
	if p.Nwords == 0 {
		return
	}
	b.Cei = p.Nwords - 1
	p.tryColumn(b, columnHint)
}

const (
	CTRL_B = 002
	CTRL_I = 011
	CTRL_U = 025
)

func (p *Paragraph) serialize() string {
	var b strings.Builder

	b.WriteRune('p')

	for _, word := range p.Words {
		sty := StyleNormal
		for idx, ru := range word.Data {
			if (sty^word.Styles[idx])&StyleItalic != 0 {
				b.WriteByte(CTRL_I)
			}
			if (sty^word.Styles[idx])&StyleBold != 0 {
				b.WriteByte(CTRL_B)
			}
			if (sty^word.Styles[idx])&StyleUnderline != 0 {
				b.WriteByte(CTRL_U)
			}
			sty = word.Styles[idx]
			b.WriteRune(ru)
		}
		b.WriteRune(' ')
	}

	return b.String()
}

func deserializeParagraph(s string) *Paragraph {
	ret := Paragraph{}

	sty := StyleNormal
	buildword := Word{}
	for _, ru := range s[1:] {
		switch ru {
		case ' ':
			ret.Words = append(ret.Words, buildword)
			ret.Nwords++
			sty = StyleNormal
			buildword = Word{}
		case CTRL_B:
			sty ^= StyleBold
		case CTRL_I:
			sty ^= StyleItalic
		case CTRL_U:
			sty ^= StyleUnderline
		default:
			buildword.Data = append(buildword.Data, ru)
			buildword.Styles = append(buildword.Styles, sty)
			buildword.Len++
			buildword.Width += termutil.Runewidth(ru)
		}
	}

	return &ret
}

func createParagraph(b *Buffer, data []rune, sty Style) *Paragraph {
	b.Dirty = true
	buildpara := Paragraph{}
	buildword := Word{}
	for _, ru := range data {
		if ru == ' ' {
			if buildword.Len != 0 {
				buildpara.Words =
					append(buildpara.Words, buildword)
				buildpara.Nwords++
				buildword = Word{}
			}
		}
		buildword.Data = append(buildword.Data, ru)
		buildword.Styles = append(buildword.Styles, sty)
		buildword.Len++
		buildword.Width += termutil.Runewidth(ru)
	}
	if buildword.Len != 0 {
		buildpara.Words = append(buildpara.Words, buildword)
		buildpara.Nwords++
	}
	if buildpara.Nwords == 0 {
		b.Cei = 0
		b.Cex = 0
	} else {
		b.Cei = buildpara.Nwords - 1
		b.Cex = buildpara.Words[b.Cei].Len
	}
	return &buildpara
}
