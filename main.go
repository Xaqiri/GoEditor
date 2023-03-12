package main

import (
	"fmt"

	"golang.org/x/term"
)

// TODO
// Look into ropes later
// Add saving and loading files

func main() {
	var e Editor
	e.initEditor()
	line := e.lines[0]
	s, _ := term.MakeRaw(0)
	defer term.Restore(0, s)

	for {

		// e.debug = []string{strconv.Itoa(e.col), strconv.Itoa(e.row)}
		e.refreshScreen()
		line = e.lines[e.row]

		inp, _, _ := e.reader.ReadRune()
		switch e.mode {
		case "move", "command":
			i := handleMoveInput(inp, &e)
			if i < 1 {
				return
			}
		case "input":
			if inp == e.ansiCodes["escape"][0] { // Pressing escape
				e.mode = "move"
				e.setCursorStyle()
			} else if inp == e.ansiCodes["return"][0] { // Pressing return
				// Split the line at the cursor
				// Part of the line after the cursor
				line = e.lines[e.row][e.cx-e.lineNumWidth:]
				// Part of the line up to the cursor
				e.lines[e.row] = e.lines[e.row][:e.col]
				// Inserts the portion of the previous line after the cursor
				// onto the new line
				e.insertLine(e.cy, line)
				e.moveCursor(e.cy+1, e.lineNumWidth)
			} else if inp == e.ansiCodes["backspace"][0] {
				if len(line) > 0 && e.col > 0 {
					if e.col < len(line) {
						left := line[:e.col-1]
						right := line[e.col:]
						line = left + right
					} else {
						line = line[:len(line)-1]
					}
					e.lines[e.row] = line
					Left(1, &e)
				}
			} else {
				line = e.lines[e.row]
				left := line[:e.col]
				right := line[e.col:]
				left += string(inp)
				line = left + right
				e.lines[e.row] = line
				Right(1, &e)
			}
		}
	}
}

func Up(n int, e *Editor) {
	if e.row > 0 {
		e.moveCursor(e.cy-1, e.cx)
	}
	if e.col > len(e.lines[e.row]) {
		e.moveCursor(e.cy, len(e.lines[e.row])+e.lineNumWidth)
	}
}

func Down(n int, e *Editor) {
	if e.row < len(e.lines)-1 {
		e.moveCursor(e.cy+1, e.cx)
	}
	if e.col > len(e.lines[e.row]) {
		e.moveCursor(e.cy, len(e.lines[e.row])+e.lineNumWidth)
	}
}

func Left(n int, e *Editor) {
	if e.col > 0 {
		e.moveCursor(e.cy, e.cx-1)
	}
}

func Right(n int, e *Editor) {
	if e.col < len(e.lines[e.row]) {
		e.moveCursor(e.cy, e.cx+1)
	}
}

func handleMoveInput(inp rune, e *Editor) int {
	if inp == 'q' {
		fmt.Printf("\033c")
		fmt.Fprintf(e.writer, "\x1b[1;1H")
		return 0
	} else if inp == 'k' {
		Up(1, e)
	} else if inp == 'j' {
		Down(1, e)
	} else if inp == 'h' {
		Left(1, e)
	} else if inp == 'l' {
		Right(1, e)
	} else if inp == ':' {
		e.mode = "command"
	} else if inp == 'i' {
		e.mode = "input"
		e.setCursorStyle()
	}
	return 1
}
