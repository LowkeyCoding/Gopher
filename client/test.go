package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell"
)

func main() {

	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e = s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	s.Clear()

	quit := make(chan struct{})

	cnt := 0
	dur := time.Duration(0)
loop:
	for {
		select {
		case <-quit:
			break loop
		case <-time.After(time.Millisecond):
		}
		start := time.Now()
		style := tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorBlack)
		str := "--- Fun, games, and other neat things ----------------------------"
		for x, char := range str {
			s.SetCell(x, 0, style, char)
		}
		s.Show()
		cnt++
		dur += time.Now().Sub(start)
	}

	s.Fini()
}
