package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	termutil "github.com/japanoise/tcell-util"
)

var sendMessage func(msg string)

func main() {
	var state *State
	if len(os.Args) > 1 {
		var err error
		state, err = loadOperation(os.Args[1])
		if err != nil {
			log.Fatalf("%+v", err)
		}
	} else {
		state = &State{}
		state.Bufs = []Buffer{{}}
		state.Bufs[0].Name = "untitled"
		state.NBufs = 1
	}
	state.regenerateDired()

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(tcell.StyleDefault)
	s.EnableMouse()
	s.Clear()

	defer func() {
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			filename := fmt.Sprintf("%d-crash.rwiir",
				time.Now().Unix())
			state.saveOperation(filename)
			fmt.Fprintf(
				os.Stderr,
				"rwiir has crashed. Data saved to %s\n",
				filename)
			panic(maybePanic)
		}
	}()

	config := &Config{}

	config.load()

	renderCallback = func(scr tcell.Screen, sx, sy int) {
		render(scr, sx, sy, state, config)
	}

	msgtime := time.Now()
	sendMessage = func(msg string) {
		state.msg = msg
		msgtime = time.Now()
	}

	sendMessage("rwiir - a prose editor. Press F1 for help.")

	for {
		if state.msg != "" && time.Since(msgtime).Seconds() > 5.0 {
			state.msg = ""
		}
		cbuf := state.CurBuf()
		s.Clear()
		sx, sy := s.Size()
		render(s, sx, sy, state, config)

		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventMouse:
			if ev.Buttons() != tcell.ButtonNone {
				click(state, config, sx, sy, ev)
			}
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyF1:
				helpscreen(s)
				continue
			case tcell.KeyF2:
				state.sidebar = !state.sidebar
				continue
			case tcell.KeyF6:
				configure(s, config)
				continue
			}

			key := termutil.ParseTcellEvent(ev)
			if config.CUA {
				switch key {
				case "TAB": // C-i
					cbuf.Sty ^= StyleItalic
					continue
				case "C-b":
					cbuf.Sty ^= StyleBold
					continue
				case "C-u":
					cbuf.Sty ^= StyleUnderline
					continue
				}
			}
			switch key {
			case "C-c":
				newName := strings.Trim(termutil.Prompt(
					s, "Filename", renderCallback), "/")
				if newName != "" {
					state.newBuffer(newName)
					state.Current = state.NBufs - 1
				}
			case "C-s":
				if state.Filename == "" {
					state.Filename = termutil.Prompt(
						s, "Filename", renderCallback)
					if state.Filename == "" {
						break
					}
				}
				err := state.saveOperation(state.Filename)
				if err == nil {
					sendMessage(fmt.Sprintf("Saved to %s",
						state.Filename))
				} else {
					sendMessage(fmt.Sprintf(
						"Save failed: %+v", err))
				}
			case "C-x":
				exportMenu(s, state, config)
			case "M-P":
				panic("Panic key")
			case "C-q":
				dirty := false
				for _, buf := range state.Bufs {
					dirty = buf.Dirty || dirty
				}
				if dirty {
					choice := termutil.YesNo(s,
						"Unsaved changes, really quit?",
						renderCallback)
					if !choice {
						continue
					}
				}
				return
			case "M-$":
				cbuf.Sty ^= StyleUnderline
			case "M-%":
				cbuf.Sty ^= StyleBold
			case "M-^":
				cbuf.Sty ^= StyleItalic
			case "M--":
				e := cbuf.SaveExcursion()
				if cbuf.Cy < cbuf.Nelems {
					cbuf.Cy++
				}
				cbuf.InsertElem(&Rule{})
				cbuf.LoadExcursion(e)
			case "M-D":
				cbuf.DeleteElem(config)
			case "M-n":
				if cbuf.Cy < cbuf.Nelems {
					cbuf.NextElem(config, 0)
				}
			case "M-p":
				if cbuf.Cy > 0 {
					cbuf.PrevElem(config, 0)
				}
			case "M-a":
				if cbuf.Cy < cbuf.Nelems {
					cbuf.Zeroes()
				}
			case "M-<":
				cbuf.Zeroes()
				cbuf.Cy = 0
				cbuf.Sy = 0
			case "M->":
				cbuf.Zeroes()
				cbuf.Cy = cbuf.Nelems
			default:
				if len(key) == 3 && key[0] == 'M' &&
					('1' <= key[2] && key[2] <= '6') {
					h := Header{}
					h.Level = int(key[2] - '0')
					h.Data = make([]rune, 0, config.Width)
					h.Datalen = 0
					cbuf.InsertElem(&h)
					cbuf.Zeroes()
				} else if cbuf.Nelems > cbuf.Cy {
					cbuf.Elems[cbuf.Cy].keyEvent(
						config, cbuf, key)
				} else {
					switch key {
					case "C-p", "UP":
						cbuf.PrevElem(config, 0)
					case "C-b", "LEFT":
						cbuf.PrevElem(config, config.Width)
					case "RET":
						cbuf.InsertElem(&Paragraph{})
						cbuf.Cy++
					case " ":
						cbuf.InsertElem(createParagraph(
							cbuf, []rune{},
							cbuf.Sty))
					default:
						if utf8.RuneCountInString(
							key) == 1 {
							ru, _ := utf8.
								DecodeRuneInString(
									key)
							cbuf.InsertElem(
								createParagraph(
									cbuf,
									[]rune{ru},
									cbuf.Sty))

						}
					}
				}
			}
		}
	}
}
