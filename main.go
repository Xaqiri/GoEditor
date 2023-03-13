package main

import (
	"golang.org/x/term"
)

// TODO
// Look into ropes later
// Add saving and loading files

func main() {
	var e Editor
	e.initEditor()
	line := e.lines[0]
	open("main.go", &e)
	s, _ := term.MakeRaw(0)
	defer term.Restore(0, s)

	for {

		e.refreshScreen()
		line = e.lines[e.row]

		inp, _ := e.reader.ReadByte()
		switch e.mode {
		case "move":
			i := handleMoveInput(inp, &e)
			if i < 1 {
				return
			}
		case "command":
			if inp == 'w' {
				e.initEditor()
			}
			e.mode = "move"
		case "input":
			if inp == e.ansiCodes["escape"][0] { // Pressing escape
				e.mode = "move"
				e.setCursorStyle()
			} else if inp == e.ansiCodes["return"][0] { // Pressing return
				// Split the line at the cursor
				// Part of the line after the cursor
				line = e.lines[e.row][e.col:]
				// Part of the line up to the cursor
				e.lines[e.row] = e.lines[e.row][:e.col]
				// Inserts the portion of the previous line after the cursor
				// onto the new line
				e.insertLine(e.cy, line)
				e.moveCursor(e.cy+1, e.lineNumWidth)
			} else if inp == e.ansiCodes["backspace"][0] {
				if len(line) > 0 && e.col > 0 {
					Left(1, &e)
					if e.col < len(line) {
						line = line[:e.col] + line[e.col+1:]
					}
					e.lines[e.row] = line
				}
			} else {
				e.lines[e.row] =
					line[:e.col] + // Get the line up to the cursor
						string(inp) + // Add the new letter
						line[e.col:] // Append the rest of the line
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
	if e.col-n > 0 {
		e.moveCursor(e.cy, e.cx-n)
	} else {
		e.moveCursor(e.cy, e.lineNumWidth)
	}
}

func Right(n int, e *Editor) {
	if e.col < len(e.lines[e.row]) {
		e.moveCursor(e.cy, e.cx+n)
	}
}

func handleMoveInput(inp byte, e *Editor) int {
	if inp == 'q' {
		e.clearScreen()
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
	} else if inp == 'o' {
		e.insertLine(e.cy, "")
		Down(1, e)
		e.mode = "input"
	} else if inp == 'O' {
		Left(e.col, e)
		e.insertLine(e.cy, e.lines[e.row])
		e.lines[e.row] = ""
		e.mode = "input"
	} else if inp == 'x' {
		line := e.lines[e.row]
		if len(line) > 0 {
			if e.col < len(line) {
				line = line[:e.col] + line[e.col+1:]
			}
			e.lines[e.row] = line
		}
	}
	return 1
}
