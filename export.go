package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	termutil "github.com/japanoise/tcell-util"
)

type outputFormat uint8

type outputType uint8

const (
	OUTPUT_FORMAT_HTML outputFormat = iota
	OUTPUT_FORMAT_MARKDOWN
	OUTPUT_FORMAT_TIDDLY
)

const (
	OUTPUT_TO_FILE outputType = iota
	OUTPUT_TO_DIRECTORY
)

type exportTags struct {
	start string
	end   string
}

// stack is a stack of capacity 3 with no bounds checking
type stack struct {
	data []string
	at   int
}

func newStack() *stack {
	return &stack{make([]string, 3), -1}
}

func (s *stack) hasNext() bool {
	return s.at >= 0
}

func (s *stack) push(str string) {
	s.at++
	s.data[s.at] = str
}

func (s *stack) pop() string {
	ret := s.data[s.at]
	s.at--
	return ret
}

func (b *Buffer) simpleExport(fi io.Writer, paragraph, rule,
	italic, underline, bold exportTags, header []exportTags) {
	for _, elem := range b.Elems {
		switch elem := elem.(type) {
		case *Header:
			fmt.Fprint(fi, header[elem.Level].start)
			fmt.Fprint(fi, string(elem.Data))
			fmt.Fprint(fi, header[elem.Level].end)
		case *Rule:
			fmt.Fprint(fi, rule.start)
			fmt.Fprint(fi, rule.end)
		case *Paragraph:
			fmt.Fprint(fi, paragraph.start)

			for idx, word := range elem.Words {
				if idx != 0 {
					fmt.Fprint(fi, " ")
				}
				st := newStack()
				sty := StyleNormal
				for jdx, ru := range word.Data {
					if word.Styles[jdx] != sty {
						sty = word.Styles[jdx]
						for st.hasNext() {
							fmt.Fprint(fi, st.pop())
						}
						if word.Styles[jdx]&StyleItalic != 0 {
							fmt.Fprint(fi, italic.start)
							st.push(italic.end)
						}
						if word.Styles[jdx]&StyleBold != 0 {
							fmt.Fprint(fi, bold.start)
							st.push(bold.end)
						}
						if word.Styles[jdx]&StyleUnderline != 0 {
							fmt.Fprint(fi, underline.start)
							st.push(underline.end)
						}
					}
					fmt.Fprintf(fi, "%c", ru)
				}
				for st.hasNext() {
					fmt.Fprint(fi, st.pop())
				}
			}

			fmt.Fprint(fi, paragraph.end)
		}
	}
}

// Picks a local file or directory and returns its path
func (state *State) pickFileOrDirectory(s tcell.Screen, prompt string) string {
	dirpath := []*dired{state.root}
	curdir := state.root
	for {
		path := ""
		for _, dir := range dirpath {
			path += dir.name
			path += "/"
		}
		var options []string
		options = []string{"<Go up>", "<This directory>"}
		for _, dir := range curdir.subdirs {
			options = append(options, dir.name+"/")
		}
		for _, fi := range curdir.files {
			idx := strings.LastIndexByte(state.Bufs[fi].Name, '/')
			name := state.Bufs[fi].Name
			options = append(options, name[idx+1:])
		}
		choice := termutil.ChoiceIndex(s, prompt+" "+path, options, 0)
		if choice == 0 {
			if curdir != state.root {
				dirpath = dirpath[:len(dirpath)-1]
				curdir = dirpath[len(dirpath)-1]
			}
		} else if choice == 1 {
			return path
		} else if len(curdir.subdirs) > 0 &&
			(choice-2) < len(curdir.subdirs) {
			curdir = curdir.subdirs[choice-2]
			dirpath = append(dirpath, curdir)
		} else {
			return state.Bufs[curdir.files[(choice-2)-len(
				curdir.subdirs)]].Name
		}
	}
}

func (state *State) outputFilename(s tcell.Screen, prompt string) string {
	return termutil.Prompt(s, prompt, nil)
}

func (state *State) obit(dir *dired) []*Buffer {
	ret := []*Buffer{}
	for _, subdir := range dir.subdirs {
		ret = append(ret, state.obit(subdir)...)
	}
	for _, fi := range dir.files {
		if !strings.Contains(state.Bufs[fi].Name, "#") {
			ret = append(ret, &state.Bufs[fi])
		}
	}
	return ret
}

func (state *State) lastDir(subdir *dired, paths []string) *dired {
	if len(paths) == 0 || paths[0] == "" {
		return subdir
	}
	for _, dir := range subdir.subdirs {
		if dir.name == paths[0] {
			return state.lastDir(dir, paths[1:])
		}
	}
	return nil
}

func (state *State) outputBuffersInTree(path string) []*Buffer {
	// Learning how to do FP was a mistake
	root := state.lastDir(state.root, strings.Split(path, "/")[1:])
	if root == nil {
		return []*Buffer{}
	}
	return state.obit(root)
}

func (state *State) doExport(frmt outputFormat, typ outputType,
	input, output string) error {
	var outputBuffers []*Buffer
	if r, _ := utf8.DecodeLastRuneInString(input); r == '/' {
		outputBuffers = state.outputBuffersInTree(input)
	} else {
		for idx := range state.Bufs {
			if state.Bufs[idx].Name == input {
				outputBuffers = append(outputBuffers,
					&state.Bufs[idx])
			}
		}
	}
	if len(outputBuffers) == 0 {
		return errors.New("no buffers selected for output")
	}

	var export func(b *Buffer, fi io.Writer)

	switch frmt {
	case OUTPUT_FORMAT_HTML:
		export = func(b *Buffer, fi io.Writer) {
			b.simpleExport(fi,
				exportTags{"<p>", "</p>\n"},
				exportTags{"<hr/>", "\n"},
				exportTags{"<em>", "</em>"},
				exportTags{"<u>", "</u>"},
				exportTags{"<b>", "</b>"},
				[]exportTags{
					{"<h0>", "</h0>\n"},
					{"<h1>", "</h1>\n"},
					{"<h2>", "</h2>\n"},
					{"<h3>", "</h3>\n"},
					{"<h4>", "</h4>\n"},
					{"<h5>", "</h5>\n"},
					{"<h6>", "</h6>\n"},
				},
			)
		}

	case OUTPUT_FORMAT_MARKDOWN:
		export = func(b *Buffer, fi io.Writer) {
			b.simpleExport(fi,
				exportTags{"", "\n\n"},
				exportTags{"------", "\n\n"},
				exportTags{"*", "*"},
				exportTags{"<u>", "</u>"},
				exportTags{"**", "**"},
				[]exportTags{
					{"<h0>", "</h0>\n"},
					{"# ", "\n\n"},
					{"## ", "\n\n"},
					{"### ", "\n\n"},
					{"#### ", "\n\n"},
					{"##### ", "\n\n"},
					{"###### ", "\n\n"},
				},
			)
		}

	case OUTPUT_FORMAT_TIDDLY:
		export = func(b *Buffer, fi io.Writer) {
			b.simpleExport(fi,
				exportTags{"", "\n\n"},
				exportTags{"------", "\n\n"},
				exportTags{"//", "//"},
				exportTags{"__", "__"},
				exportTags{"''", "''"},
				[]exportTags{
					{"<h0>", "</h0>\n"},
					{"! ", "\n\n"},
					{"!! ", "\n\n"},
					{"!!! ", "\n\n"},
					{"!!!! ", "\n\n"},
					{"!!!!! ", "\n\n"},
					{"!!!!!! ", "\n\n"},
				},
			)
		}
	}

	if typ == OUTPUT_TO_FILE {
		if err := os.MkdirAll(
			filepath.Dir(output), 0755); err != nil {
			return err
		}

		fi, err := os.Create(output)
		if err != nil {
			return err
		}

		for _, buf := range outputBuffers {
			export(buf, fi)
		}

		fi.Close()
	} else if typ == OUTPUT_TO_DIRECTORY {
		if err := os.MkdirAll(output, 0755); err != nil {
			return err
		}

		for _, buf := range outputBuffers {
			realpath := strings.TrimPrefix(buf.Name, input)
			if output != "" {
				realpath = output + "/" + realpath
			}
			if err := os.MkdirAll(
				filepath.Dir(realpath), 0755); err != nil {
				return err
			}

			fi, err := os.Create(realpath)
			if err != nil {
				return err
			}
			export(buf, fi)

			fi.Close()
		}
	}
	return nil
}

func exportMenu(s tcell.Screen, state *State, c *Config) {
	xit := tcell.StyleDefault.Background(tcell.ColorRed)
	msgStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	reverse := tcell.StyleDefault.Reverse(true)
	focus := 0
	formats := []string{
		"HTML",
		"Markdown",
		"TiddlyWiki",
	}
	selFormat := 0
	types := []string{
		"Single File",
		"Directory",
	}
	selOutput := 0
	inputFilePath := ""
	inputLen := 0
	outputFilename := ""
	outputLen := 0
	message := ""
	formatFunc := func(s tcell.Screen, selection, sx, sy int) {
		termutil.PrintString(s, sx-40, 1, formats[selection])
		switch selection {
		case 0:
			termutil.PrintString(s, sx-40, 3,
				"HTML fragment. Suitable for AO3.")
		case 1:
			termutil.PrintString(s, sx-40, 3,
				"Relatively compliant Markdown.")
		case 2:
			termutil.PrintString(s, sx-40, 3,
				"WikiText suitable for TiddlyWiki.")
		}
	}
	outputFunc := func(s tcell.Screen, selection, sx, sy int) {
		termutil.PrintString(s, sx-40, 1, types[selection])
		switch selection {
		case 0:
			termutil.PrintString(s, sx-40, 3,
				"Output to a single, large file.")
		case 1:
			termutil.PrintString(s, sx-40, 3,
				"Output individual files to a directory.")
		}
	}

	formatAction := func() {
		selFormat = termutil.ChoiceIndexCallback(s, "Output Format",
			formats, selFormat, formatFunc)
	}
	outputAction := func() {
		selOutput = termutil.ChoiceIndexCallback(s, "Output Type",
			types, selOutput, outputFunc)
	}
	inputAction := func() {
		inputFilePath = state.pickFileOrDirectory(s,
			"Export which file/directory?")
		inputLen = termutil.RunewidthStr(inputFilePath)
	}
	filenameAction := func() {
		outputFilename = state.outputFilename(s, "Output to")
		outputLen = termutil.RunewidthStr(outputFilename)
	}
	exportAction := func() error {
		if outputLen == 0 || inputLen == 0 {
			message = "enter input/output filenames"
			return errors.New(message)
		}
		err := state.doExport(
			outputFormat(selFormat), outputType(selOutput),
			inputFilePath, outputFilename)
		if err != nil {
			message = err.Error()
		}
		return err
	}

	for {
		s.Clear()
		sx, sy := s.Size()
		topline := (sy / 2) - 6
		anchorx := (sx / 2) - 28
		termutil.PrintRuneStyle(s, sx-3, 0, ' ', xit)
		termutil.PrintRuneStyle(s, sx-2, 0, ' ', xit)
		termutil.PrintRuneStyle(s, sx-1, 0, ' ', xit)
		termutil.PrintRuneStyle(s, sx-3, 1, ' ', xit)
		termutil.PrintRuneStyle(s, sx-2, 1, 'X', xit)
		termutil.PrintRuneStyle(s, sx-1, 1, ' ', xit)
		termutil.PrintRuneStyle(s, sx-3, 2, ' ', xit)
		termutil.PrintRuneStyle(s, sx-2, 2, ' ', xit)
		termutil.PrintRuneStyle(s, sx-1, 2, ' ', xit)

		termutil.PrintString(s, anchorx+16, topline, "Format")
		termutil.PrintString(s, anchorx+36, topline, "Output")

		for i := 0; i < 16; i++ {
			termutil.PrintRuneStyle(
				s, i+anchorx+11, topline+1, ' ', reverse)
			termutil.PrintRuneStyle(
				s, i+anchorx+31, topline+1, ' ', reverse)
		}

		termutil.PrintStringStyle(s,
			anchorx+11, topline+1, formats[selFormat], reverse)
		termutil.PrintStringStyle(s,
			anchorx+31, topline+1, types[selOutput], reverse)

		for i := 0; i <= 56; i++ {
			termutil.PrintRuneStyle(
				s, anchorx+i, topline+4, ' ', reverse)
			termutil.PrintRuneStyle(
				s, anchorx+i, topline+7, ' ', reverse)
		}

		termutil.PrintString(s, anchorx, topline+3, "Input File(s)")
		termutil.PrintStringStyle(s,
			anchorx, topline+4, inputFilePath, reverse)

		termutil.PrintString(s, anchorx, topline+6, "Output Filename")
		termutil.PrintStringStyle(s,
			anchorx, topline+7, outputFilename, reverse)

		termutil.PrintStringStyle(s,
			anchorx+24, topline+9, "        ", reverse)
		termutil.PrintStringStyle(s,
			anchorx+24, topline+10, " Export ", reverse)
		termutil.PrintStringStyle(s,
			anchorx+24, topline+11, "        ", reverse)

		termutil.PrintStringStyle(s,
			anchorx, topline+14, message, msgStyle)

		switch focus {
		case 0: // Format
			s.ShowCursor(anchorx+11, topline+1)
		case 1: // Output
			s.ShowCursor(anchorx+31, topline+1)
		case 2: // Input File(s)
			s.ShowCursor(anchorx+inputLen, topline+4)
		case 3: // Output filename
			s.ShowCursor(anchorx+outputLen, topline+7)
		case 4: // Export button
			s.ShowCursor(anchorx+31, topline+10)
		}

		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventMouse:
			if ev.Buttons() != tcell.ButtonNone {
				cx, cy := ev.Position()
				if cx >= sx-3 && cy <= 2 {
					return
				} else if cx < anchorx || cx > anchorx+56 ||
					cy < topline || cy > topline+11 {
					// Not inside the UI box
					continue
				} else if cy == topline || cy == topline+1 {
					if anchorx+11 <= cx && cx <= anchorx+27 {
						formatAction()
					} else if anchorx+31 <= cx &&
						cx <= anchorx+47 {
						outputAction()
					}
				} else if cy == topline+3 || cy == topline+4 {
					inputAction()
				} else if cy == topline+6 || cy == topline+7 {
					filenameAction()
				} else if anchorx+24 <= cx && cx <= anchorx+32 &&
					topline+9 <= cy {
					err := exportAction()
					if err == nil {
						return
					}
					exportAction()
				}
			}
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEsc, tcell.KeyCtrlC, tcell.KeyCtrlG,
				tcell.KeyCtrlQ:
				return
			case tcell.KeyBacktab, tcell.KeyLeft, tcell.KeyCtrlB:
				focus--
				if focus < 0 {
					focus = 4
				}
			case tcell.KeyTab, tcell.KeyRight, tcell.KeyCtrlF:
				focus++
				if focus > 4 {
					focus = 0
				}
			case tcell.KeyDown, tcell.KeyCtrlN:
				if focus < 2 {
					focus = 2
					break
				}
				focus++
				if focus > 4 {
					focus = 0
				}
			case tcell.KeyUp, tcell.KeyCtrlP:
				if focus == 2 {
					focus = 0
					break
				}
				focus--
				if focus < 0 {
					focus = 4
				}
			case tcell.KeyEnter:
				switch focus {
				case 0: // Format
					formatAction()
				case 1: // Output
					outputAction()
				case 2: // Input File(s)
					inputAction()
				case 3: // Output filename
					filenameAction()
				case 4: // Export button
					err := exportAction()
					if err == nil {
						return
					}
				}
			}
		}
	}
}
