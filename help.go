package main

import (
	"github.com/gdamore/tcell/v2"
	termutil "github.com/japanoise/tcell-util"
)

func helpscreen(s tcell.Screen) {
	inverse := tcell.StyleDefault.Reverse(true)
	for {
		s.Clear()
		s.HideCursor()
		sx, sy := s.Size()

		anc := (sx - 30) / 2
		termutil.PrintString(s, anc, 0,
			"               .__.__        ")
		termutil.PrintString(s, anc, 1,
			"_________  _  _|__|__|______ ")
		termutil.PrintString(s, anc, 2,
			"\\_  __ \\ \\/ \\/ /  |  \\_  __ \\")
		termutil.PrintString(s, anc, 3,
			" |  | \\/\\     /|  |  ||  | \\/")
		termutil.PrintString(s, anc, 4,
			" |__|    \\/\\_/ |__|__||__|   ")
		termutil.PrintString(s, anc, 5, "")
		termutil.PrintString(s, anc, 6, "rwiir - alleged prose editor")
		termutil.PrintString(s, anc-3, 7,
			"unleashed on an unsuspecting world")
		termutil.PrintString(s, anc+8, 8, "by japanoise")
		termutil.PrintString(s, anc+2, sy-1,
			"Press any key to close...")

		colW := sx / 3
		colAnc := (colW - 26) / 2

		termutil.PrintStringStyle(s, colAnc+9, 11, "Movement", inverse)
		termutil.PrintString(s, colAnc, 12, "Arrows, etc work")
		termutil.PrintString(s, colAnc, 13, "^F Forward     M-f Word")
		termutil.PrintStringStyle(s, colAnc, 13, "^F", inverse)
		termutil.PrintStringStyle(s, colAnc+15, 13, "M-f", inverse)
		termutil.PrintString(s, colAnc, 14, "^B Backward    M-b Word")
		termutil.PrintStringStyle(s, colAnc, 14, "^B", inverse)
		termutil.PrintStringStyle(s, colAnc+15, 14, "M-b", inverse)
		termutil.PrintString(s, colAnc, 15, "^P Up (Prev)   M-p Para")
		termutil.PrintStringStyle(s, colAnc, 15, "^P", inverse)
		termutil.PrintStringStyle(s, colAnc+15, 15, "M-p", inverse)
		termutil.PrintString(s, colAnc, 16, "^N Down (Next) M-n Para")
		termutil.PrintStringStyle(s, colAnc, 16, "^N", inverse)
		termutil.PrintStringStyle(s, colAnc+15, 16, "M-n", inverse)
		termutil.PrintString(s, colAnc, 17, "^A Beg of Line M-a Para")
		termutil.PrintStringStyle(s, colAnc, 17, "^A", inverse)
		termutil.PrintStringStyle(s, colAnc+15, 17, "M-a", inverse)
		termutil.PrintString(s, colAnc, 18, "^E End of Line M-e Para")
		termutil.PrintStringStyle(s, colAnc, 18, "^E", inverse)
		termutil.PrintStringStyle(s, colAnc+15, 18, "M-e", inverse)
		termutil.PrintString(s, colAnc, 19, "M-< Beginning of File")
		termutil.PrintStringStyle(s, colAnc, 19, "M-<", inverse)
		termutil.PrintString(s, colAnc, 20, "M-> End of File")
		termutil.PrintStringStyle(s, colAnc, 20, "M->", inverse)

		colAnc += colW
		termutil.PrintStringStyle(s, colAnc+10, 11, "Editing", inverse)
		termutil.PrintString(s, colAnc, 12, "Type etc. to add text")
		termutil.PrintString(s, colAnc, 13, "M-d Delete word forward")
		termutil.PrintStringStyle(s, colAnc, 13, "M-d", inverse)
		termutil.PrintString(s, colAnc, 14, "M-^H, M-Backspace del word")
		termutil.PrintStringStyle(s, colAnc, 14, "M-^H", inverse)
		termutil.PrintStringStyle(s, colAnc+6, 14, "M-Backspace", inverse)
		termutil.PrintString(s, colAnc, 15, "^U Delete to BoL")
		termutil.PrintStringStyle(s, colAnc, 15, "^U", inverse)
		termutil.PrintString(s, colAnc, 16, "^K Delete to EoL")
		termutil.PrintStringStyle(s, colAnc, 16, "^K", inverse)
		termutil.PrintString(s, colAnc, 18, "M-1 thru M-6 add header")
		termutil.PrintStringStyle(s, colAnc, 18, "M-1", inverse)
		termutil.PrintStringStyle(s, colAnc+9, 18, "M-6", inverse)
		termutil.PrintString(s, colAnc, 19, "M-- Add horizontal line")
		termutil.PrintStringStyle(s, colAnc, 19, "M--", inverse)
		termutil.PrintString(s, colAnc, 21, "M-D Del this paragraph")
		termutil.PrintStringStyle(s, colAnc, 21, "M-D", inverse)

		colAnc += colW + 1
		termutil.PrintStringStyle(s, colAnc+8, 11, "Interface", inverse)
		termutil.PrintString(s, colAnc, 12, "F1 Help")
		termutil.PrintStringStyle(s, colAnc, 12, "F1", inverse)
		termutil.PrintString(s, colAnc, 13, "F2 Sidebar")
		termutil.PrintStringStyle(s, colAnc, 13, "F2", inverse)
		termutil.PrintString(s, colAnc, 14, "F6 Configuration")
		termutil.PrintStringStyle(s, colAnc, 14, "F6", inverse)
		termutil.PrintString(s, colAnc, 21, "^S Save ^C New file")
		termutil.PrintStringStyle(s, colAnc, 21, "^S", inverse)
		termutil.PrintStringStyle(s, colAnc+8, 21, "^C", inverse)
		termutil.PrintString(s, colAnc, 22, "^Q Quit ^X Export")
		termutil.PrintStringStyle(s, colAnc, 22, "^Q", inverse)
		termutil.PrintStringStyle(s, colAnc+8, 22, "^X", inverse)

		s.Show()

		ev := s.PollEvent()
		switch ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			return
		}
	}
}
