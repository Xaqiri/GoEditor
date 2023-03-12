package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"golang.org/x/term"
)

type Editor struct {
	lines        []string
	writer       io.Writer
	reader       *bufio.Reader
	w, h         int // Width and height of terminal
	col, row     int // Width of current row and max number of rows
	cx, cy       int // Cursor position
	prompt, mode string
	lineNumWidth int
	debug        []string
	ansiCodes    map[string][]rune
}

func (e *Editor) initEditor() {
	e.ansiCodes = map[string][]rune{
		"escape":    {'\033'},
		"return":    {'\x0D'},
		"backspace": {'\u007F'},
		"clear":     {'\033', '[', '2', 'J'},
		"move":      {'\033', '[', ' ', ';', ' ', 'H'},
	}
	e.lines = []string{""}
	e.cy = 1
	e.updatePrompt()
	e.lineNumWidth = len(e.prompt) + 1
	e.writer = os.Stdout
	e.reader = bufio.NewReader(os.Stdin)
	e.mode = "input"
	e.cx = e.lineNumWidth
	e.row = e.cy - 1
	e.col = e.cx - e.lineNumWidth
	e.w, e.h, _ = term.GetSize(0)
	e.debug = e.lines
}

func (e *Editor) moveCursor(row, col int) {
	fmt.Fprintf(e.writer, "\033[%d;%dH", row, col)
	e.cx = col
	e.cy = row
	e.row = e.cy - 1
	e.col = e.cx - e.lineNumWidth
}

func (e *Editor) updatePrompt() {
	p := ""
	if e.cy <= len(e.lines) {
		p = strconv.Itoa(e.cy)
	} else {
		p = "~"
	}
	for len(p) < 3 {
		p += " "
	}
	e.prompt = p
	e.lineNumWidth = len(e.prompt) + 1
}

func (e *Editor) setCursorStyle() {
	if e.mode == "input" {
		// Line
		fmt.Fprintf(e.writer, "\x1b[6 q")
	} else {
		// Block
		fmt.Fprintf(e.writer, "\x1b[2 q")
	}
}

func (e *Editor) insertLine(lineNumber int, line string) {
	temp := make([]string, len(e.lines)+1)
	copy(temp, e.lines[:lineNumber])
	temp[lineNumber] = line
	copy(temp[lineNumber+1:], e.lines[lineNumber:])
	e.lines = temp
}

func (e *Editor) clearScreen() {
	// Clear screen
	fmt.Fprintf(e.writer, string(e.ansiCodes["clear"]))
}

func (e *Editor) refreshScreen() {
	x, y := e.cx, e.cy
	e.clearScreen()
	e.drawLineNums()
	e.drawDocument()
	e.drawBottomInfo(x, y)
	e.setCursorStyle()
	e.moveCursor(y, x)
}

func (e *Editor) drawLineNums() {
	for i := 1; i < e.h; i++ {
		e.moveCursor(i, 1)
		e.updatePrompt()
		fmt.Print(e.prompt)
	}
}

func (e *Editor) drawDocument() {
	for i := 1; i <= len(e.lines); i++ {
		e.moveCursor(i, e.lineNumWidth)
		fmt.Print(e.lines[e.row])
	}
}

func (e *Editor) drawBottomInfo(x, y int) {
	// strings.Join(e.debug, ",")

	e.moveCursor(y, x)
	btm := ""
	mode := " " + e.mode
	coord := strconv.Itoa(e.col) + ":" + strconv.Itoa(e.cy)
	// Reverse bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[7m")
	// Move cursor to the bottom left of the screen
	fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.h, 1)
	// Clear the line
	fmt.Fprintf(e.writer, "\x1b[2K")
	btm += mode
	for i := len(mode); i < e.w-len(coord); i++ {
		btm += " "
	}
	btm += coord
	fmt.Print(btm)
	// Reset the bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[m")
}
