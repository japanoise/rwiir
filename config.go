package main

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	termutil "github.com/japanoise/tcell-util"
)

func customizeOneStyle(s tcell.Screen, config *Config,
	sty *tcell.Style, def tcell.Style, name string) {
	type settingChoices struct {
		name string
		fn   func()
	}

	choices := []settingChoices{
		{"Foreground", func() {
			*sty = sty.Foreground(
				termutil.PickColor(s, "Foreground Color"))
		}},
		{"Background", func() {
			*sty = sty.Background(
				termutil.PickColor(s, "Background Color"))
		}},
		{"Bold", func() {
			_, _, attr := sty.Decompose()
			*sty = sty.Attributes(attr ^ tcell.AttrBold)
		}},
		{"Blink", func() {
			_, _, attr := sty.Decompose()
			*sty = sty.Attributes(attr ^ tcell.AttrBlink)
		}},
		{"Reverse", func() {
			_, _, attr := sty.Decompose()
			*sty = sty.Attributes(attr ^ tcell.AttrReverse)
		}},
		{"Underline", func() {
			_, _, attr := sty.Decompose()
			*sty = sty.Attributes(attr ^ tcell.AttrUnderline)
		}},
		{"Dim", func() {
			_, _, attr := sty.Decompose()
			*sty = sty.Attributes(attr ^ tcell.AttrDim)
		}},
		{"Italic", func() {
			_, _, attr := sty.Decompose()
			*sty = sty.Attributes(attr ^ tcell.AttrItalic)
		}}}

	options := make([]string, 0, len(choices)+2)
	for _, choice := range choices {
		options = append(options, choice.name)
	}
	options = append(options, "Reset to default")
	options = append(options, "Done")

	for {
		choice := termutil.ChoiceIndexCallback(s, name, options, 0,
			func(sc tcell.Screen, ch, sx, sy int) {
				termutil.PrintStringStyle(sc, sx-43, 2,
					"Current Preview", *sty)
				termutil.PrintStringStyle(sc, sx-43, 4,
					"Default Preview", def)
			})
		if choice < len(choices) {
			choices[choice].fn()
		} else if choice-len(choices) == 0 {
			*sty = def
		} else {
			return
		}
	}
}

func customizeColors(s tcell.Screen, config *Config) {
	type colorChoices struct {
		name string
		desc string
		ptr  *tcell.Style
		def  tcell.Style
	}

	choices := []colorChoices{
		{"Warning",
			"Used for a warning or error message",
			&config.Warning,
			tcell.StyleDefault.Foreground(
				tcell.ColorRed).Reverse(true)},
		{"Italic",
			"Used for italic body text",
			&config.Italic,
			tcell.StyleDefault.Attributes(tcell.AttrItalic)},
		{"Bold",
			"Used for bold body text",
			&config.Bold,
			tcell.StyleDefault.Attributes(tcell.AttrBold)},
		{"Underline",
			"Used for underlined body text",
			&config.Underline,
			tcell.StyleDefault.Attributes(tcell.AttrUnderline)},
		{"ItUnd",
			"Used for italic underlined body text",
			&config.ItUnd,
			tcell.StyleDefault.Attributes(
				tcell.AttrItalic | tcell.AttrUnderline)},
		{"UndBold",
			"Used for bold underlined body text",
			&config.UndBold,
			tcell.StyleDefault.Attributes(
				tcell.AttrUnderline | tcell.AttrBold)},
		{"BoldIt",
			"Used for bold italic body text",
			&config.BoldIt,
			tcell.StyleDefault.Attributes(
				tcell.AttrItalic | tcell.AttrBold)},
		{"BoldUndIt",
			"Used for bold italic underlined body text",
			&config.BoldUndIt,
			tcell.StyleDefault.Attributes(
				tcell.AttrUnderline | tcell.AttrItalic |
					tcell.AttrBold)},
		{"Header",
			"Used for headers",
			&config.Header,
			tcell.StyleDefault.Foreground(
				tcell.ColorGreen).Attributes(
				tcell.AttrBold | tcell.AttrUnderline)},
		{"Rule",
			"Used for horizontal rules",
			&config.Rule,
			tcell.StyleDefault.Foreground(tcell.ColorBlue)},
		{"Modeline",
			"Used for the modeline/status bar",
			&config.Modeline,
			tcell.StyleDefault.Reverse(true)},
		{"Dired",
			"Used to signify a directory in the sidebar",
			&config.Dired,
			tcell.StyleDefault.Foreground(
				tcell.ColorFuchsia).Attributes(tcell.AttrBold)}}

	options := make([]string, 0, len(choices)+1)
	for _, choice := range choices {
		options = append(options, choice.name)
	}
	options = append(options, "Done")

	for {
		choice := termutil.ChoiceIndexCallback(s, "Customize", options,
			0, func(sc tcell.Screen, choice, sx, sy int) {
				if choice < len(choices) {
					termutil.PrintString(sc, sx-43, 3,
						choices[choice].desc)
					termutil.PrintStringStyle(sc, sx-43, 5,
						"Current Preview",
						*choices[choice].ptr)
					termutil.PrintStringStyle(sc, sx-43, 7,
						"Default Preview",
						choices[choice].def)
				} else {
					termutil.PrintString(sc, sx-43, 3,
						"Return to the config screen")
				}
			})
		if choice < len(choices) {
			customizeOneStyle(s, config,
				choices[choice].ptr,
				choices[choice].def,
				choices[choice].name)
		} else {
			return
		}
	}
}

func configure(s tcell.Screen, config *Config) {
	type configChoices struct {
		choice string
		msg    string
		f      func() bool
	}

	choices := []configChoices{
		{"Screen Width", "Set the width of the screen", func() bool {
			choice := termutil.Prompt(s,
				"Width of screen in characters (default 79)",
				nil)
			w, err := strconv.Atoi(choice)
			sx, _ := s.Size()
			if err == nil && 30 < w && w < sx {
				config.Width = w
			}
			return true
		}},
		{"Bindings", "Choose which keybindings to use", func() bool {
			config.CUA = !config.CUA
			return true
		}},
		{"Customize", "Customize the colors & appearance", func() bool {
			customizeColors(s, config)
			return true
		}},
		{"Done", "Return to the main screen",
			func() bool { return false }},
	}

	running := true
	for running {
		options := make([]string, 4)
		for i, c := range choices {
			options[i] = c.choice
		}

		options[0] = fmt.Sprintf(
			"Screen width (currently %d)", config.Width)

		if config.CUA {
			options[1] = "Use Emacs bindings"
		} else {
			options[1] = "Use CUA-type bindings"
		}

		choice := termutil.ChoiceIndexCallback(s, "Configure rwiir",
			options, 0, func(sc tcell.Screen, choice, sx, sy int) {
				termutil.PrintString(sc, sx-43, 2,
					choices[choice].msg)
			})

		running = choices[choice].f()
	}
	config.save()
}
