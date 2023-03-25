package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/exp/slices"
)

type State struct {
	Bufs     []Buffer
	Trash    []Buffer
	Current  int
	NBufs    int
	Filename string
	msg      string
	sidebar  bool
	root     *dired
}

func (s *State) newBuffer(name string) {
	newBuf := Buffer{}
	newBuf.Dirty = true
	newBuf.Name = name
	s.Bufs = append(s.Bufs, newBuf)
	s.NBufs++
	s.regenerateDired()
}

func (s *State) changeBuffer(idx int) {
	s.Current = idx
	// We don't have to manipulate the position - we may in future.
	// i.e. If adding some way to mutate the buffer outside of it.
}

func (s *State) regenerateDired() {
	s.root = &dired{}
	for idx, buf := range s.Bufs {
		s.root.insert(idx, &buf, buf.Name)
	}
	s.root.sort(s)
}

func (s *State) saveOperation(fn string) error {
	fi, err := os.Create(fn)
	if err != nil {
		return err
	}

	defer fi.Close()

	fmt.Fprintf(fi, "%s\n", s.Filename)
	fmt.Fprintf(fi, "%d\n", s.Current)
	for idx := range s.Bufs {
		s.Bufs[idx].Serialize(fi)
		s.Bufs[idx].Dirty = false
	}
	fmt.Fprintln(fi, "EOF")

	return nil
}

func loadOperation(fn string) (*State, error) {
	fi, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	defer fi.Close()
	ret := &State{}

	// This allows lines of up to 1mb; beyond which the user is on their own
	scanner := bufio.NewScanner(fi)
	const maxCapacity int = 100000
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	if !scanner.Scan() {
		return nil, errors.New("eof reached before filename")
	}
	ret.Filename = scanner.Text()

	if !scanner.Scan() {
		return nil, errors.New("eof reached before current buffer idx")
	}
	ret.Current, err = strconv.Atoi(scanner.Text())

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		} else if line[0] == 'B' {
			buf, err := DeserializeBuffer(scanner, line[1:])
			if err != nil {
				return nil, err
			}
			ret.Bufs = append(ret.Bufs, *buf)
			ret.NBufs++
		} else if line == "EOF" {
			break
		}
	}

	return ret, nil
}

type dired struct {
	open    bool
	name    string
	subdirs []*dired
	files   []int
}

func (d *dired) insert(idx int, buf *Buffer, partial string) {
	dirs := strings.SplitN(partial, "/", 2)
	if len(dirs) == 1 {
		d.files = append(d.files, idx)
		return
	}
	for _, dir := range d.subdirs {
		if dir.name == dirs[0] {
			dir.insert(idx, buf, dirs[1])
			return
		}
	}
	newdir := &dired{}
	newdir.name = dirs[0]
	d.subdirs = append(d.subdirs, newdir)
	newdir.insert(idx, buf, dirs[1])
}

func (d *dired) sort(s *State) {
	for _, dir := range d.subdirs {
		dir.sort(s)
	}
	sort.Slice(d.subdirs, func(i, j int) bool {
		return d.subdirs[i].name < d.subdirs[j].name
	})
	sort.Slice(d.files, func(i, j int) bool {
		return s.Bufs[d.files[i]].Name < s.Bufs[d.files[j]].Name
	})
}

type Config struct {
	Width     int
	Warning   tcell.Style
	Underline tcell.Style
	Italic    tcell.Style
	ItUnd     tcell.Style
	Bold      tcell.Style
	BoldIt    tcell.Style
	UndBold   tcell.Style
	BoldUndIt tcell.Style
	Header    tcell.Style
	Rule      tcell.Style
	Modeline  tcell.Style
	Dired     tcell.Style
	CUA       bool
}

type Style uint8

const (
	StyleNormal Style = iota
	StyleUnderline
	StyleItalic
	StyleItalicUnderline
	StyleBold
	StyleBoldUnderline
	StyleBoldIt
	StyleBoldUnderlineItalic
)

func (c *Config) style2style(sty Style) tcell.Style {
	switch sty {
	case StyleNormal:
		return tcell.StyleDefault
	case StyleUnderline:
		return c.Underline
	case StyleItalic:
		return c.Italic
	case StyleItalicUnderline:
		return c.ItUnd
	case StyleBold:
		return c.Bold
	case StyleBoldUnderline:
		return c.UndBold
	case StyleBoldIt:
		return c.BoldIt
	case StyleBoldUnderlineItalic:
		return c.BoldUndIt
	}
	// Shouldn't happen
	return tcell.StyleDefault
}

type Buffer struct {
	Elems  []Element
	Nelems int
	Sty    Style
	Name   string
	Words  int
	Dirty  bool
	// Global positions - index of Element @ top of screen & current Element
	Sy int
	Cy int
	// Per-Element positions - it's up to Elements to set & use these
	Cex int
	Cey int
	Cei int
	// Pseudo per-Element position - main may update this
	Sey int
}

type Excursion struct {
	sy  int
	cy  int
	cex int
	cey int
	cei int
	sey int
}

func (s *State) CurBuf() *Buffer {
	return &(s.Bufs[s.Current])
}

func (b *Buffer) Serialize(fi io.Writer) {
	fmt.Fprintf(fi, "B%s\n", b.Name)
	for _, elem := range b.Elems {
		fmt.Fprintln(fi, elem.serialize())
	}
	fmt.Fprintf(fi, "EOB\n")
}

func DeserializeBuffer(scanner *bufio.Scanner, fn string) (*Buffer, error) {
	ret := Buffer{}
	ret.Name = fn
	for scanner.Scan() {
		line := scanner.Text()
		if line == "EOB" {
			return &ret, nil
		}
		switch line[0] {
		case 'h':
			ret.Elems = append(ret.Elems, deserializeHeader(line))
			ret.Nelems++
		case 'r':
			ret.Elems = append(ret.Elems, deserializeRule(line))
			ret.Nelems++
		case 'p':
			p := deserializeParagraph(line)
			ret.Elems = append(ret.Elems, p)
			ret.Nelems++
			ret.Words += p.Nwords
		case '#':
			// Comment! Do nothing for now.
			continue
		}
	}
	return nil, errors.New("reached end of file while reading a buffer")
}

func (b *Buffer) Zeroes() {
	b.Cex = 0
	b.Cei = 0
	b.Cey = 0
	b.Sey = 0
}

func (b *Buffer) NextElem(c *Config, columnHint int) {
	if b.Cy < b.Nelems {
		b.Cy++
	} else {
		return
	}
	if b.Cy < b.Nelems {
		b.Elems[b.Cy].startOf(c, b, columnHint)
	} else {
		b.Zeroes()
	}
}

func (b *Buffer) PrevElem(c *Config, columnHint int) {
	if b.Cy > 0 {
		b.Cy--
	} else {
		return
	}
	b.Elems[b.Cy].endOf(c, b, columnHint)
}

func (b *Buffer) InsertElem(e Element) {
	b.Dirty = true
	if b.Cy == b.Nelems {
		b.Elems = append(b.Elems, e)
		b.Nelems++
	} else {
		b.Elems = slices.Insert(b.Elems, b.Cy, e)
		b.Nelems++
	}
}

func (b *Buffer) DeleteElem(c *Config) {
	b.Dirty = true
	if b.Cy < b.Nelems {
		b.Elems = slices.Delete(b.Elems, b.Cy, b.Cy+1)
		b.Nelems--
		b.WordCount()
	}
	if b.Cy == b.Nelems {
		b.Zeroes()
	} else {
		b.Elems[b.Cy].startOf(c, b, 0)
	}
}

func (b *Buffer) WordCount() {
	wc := 0
	for _, elem := range b.Elems {
		switch el := elem.(type) {
		case *Paragraph:
			wc += el.Nwords
		}
	}
	b.Words = wc
}

func (b *Buffer) SaveExcursion() *Excursion {
	ret := Excursion{}

	ret.cei = b.Cei
	ret.cex = b.Cex
	ret.cey = b.Cey
	ret.sey = b.Sey
	ret.cy = b.Cy

	return &ret
}

func (b *Buffer) LoadExcursion(e *Excursion) {
	b.Cei = e.cei
	b.Cex = e.cex
	b.Cey = e.cey
	b.Sey = e.sey
	b.Cy = e.cy
}
