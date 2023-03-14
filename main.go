package main

import (
	"os"
	// "strconv"

	"golang.org/x/term"
)

// TODO
// Look into ropes later
// Move cursor related stuff to its own file
// Move file related stuff to its own file

func main() {
	args := os.Args
	var fn string
	var e Editor
	e.initEditor()

	if len(args) > 1 {
		fn = args[1]
		open(fn, &e)
		e.debug = append(e.debug, fn)
	}
	if len(e.lines) == 0 {
		e.lines = append(e.lines, "")
	}
	s, _ := term.MakeRaw(0)
	defer term.Restore(0, s)

	for {
		// e.debug = append(e.debug, strconv.Itoa(e.offset))
		// e.debug = append(e.debug, strconv.Itoa(e.cx))
		// e.debug = append(e.debug, strconv.Itoa(e.col))

		e.refreshScreen()
		line := e.lines[e.row]

		inp, _ := e.reader.ReadByte()
		switch e.mode {
		case "move":
			i := handleMoveInput(inp, &e)
			if i < 1 {
				return
			}
		case "command":
			switch inp {
			case e.ansiCodes["escape"][0]:
				e.cmd = []string{"", ""}
				e.mode = "move"
			case e.ansiCodes["return"][0]:
				if e.cmd[1] == "q" {
					e.clearScreen()
					os.Exit(0)
				} else if e.cmd[1] == "w" {
					if fn != "" {
						e.save(fn)
					} else {
						e.cmd[0] = ""
						e.debug = append(e.debug, "No file to save")
					}
				}
				e.cmd = []string{"", ""}
			case e.ansiCodes["backspace"][0]:
				if len(e.cmd[1]) > 0 {
					e.cmd[1] = e.cmd[1][:len(e.cmd[1])-1]
				}
			default:
				e.debug = []string{""}
				e.cmd[0] = ":"
				e.cmd[1] += string(inp)
			}

		case "input":
			if inp == e.ansiCodes["escape"][0] { // Pressing escape
				e.mode = "move"
				e.setCursorStyle()
			} else if inp == e.ansiCodes["return"][0] { // Pressing return
				// Split the line at the cursor
				left := e.lines[e.row][e.col:]  // Part of the line up to the cursor
				right := e.lines[e.row][:e.col] // Part of the line after the cursor
				e.lines[e.row] = left           // Current row will contain characters up to the cursor
				e.insertLine(e.row, right)      // Add a new line below with the rest of the characters
				Down(1, &e)                     // Move down to the new line
			} else if inp == e.ansiCodes["backspace"][0] {
				if len(line) > 0 && e.col > 0 {
					Left(1, &e)
					if e.col < len(line) {
						line = line[:e.col] + line[e.col+1:]
					}
					e.lines[e.row] = line
				}
			} else {
				// Typing new characters
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
	if e.cy < 1 {
		e.offset--
		e.cy = 1
	}
}

func Down(n int, e *Editor) {
	if e.row < len(e.lines)-1 {
		e.moveCursor(e.cy+1, e.cx)
	}
	if e.col > len(e.lines[e.row]) {
		e.moveCursor(e.cy, len(e.lines[e.row])+e.lineNumWidth)
	}
	if e.cy > e.h-1 && len(e.lines) >= e.h-1 {
		e.offset++
	}
	if e.cy > e.h-1 {
		e.cy = e.h - 1
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
	if inp == 'k' {
		Up(1, e)
	} else if inp == 'j' {
		Down(1, e)
	} else if inp == 'h' {
		Left(1, e)
	} else if inp == 'l' {
		Right(1, e)
	} else if inp == ':' {
		e.cmd[0] = ":"
		e.mode = "command"
	} else if inp == 'i' {
		e.mode = "input"
		e.setCursorStyle()
	} else if inp == 'o' {
		e.insertLine(e.row+1, "")
		Down(1, e)
		e.mode = "input"
	} else if inp == 'O' {
		Left(e.col, e)
		e.insertLine(e.row, e.lines[e.row])
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
	} else if inp == 'D' {
		if e.row != len(e.lines)-1 {
			temp := make([]string, len(e.lines)-1)
			copy(temp, e.lines[:e.row])
			copy(temp[e.row:], e.lines[e.row+1:])
			e.lines = temp
			dif := e.col - len(e.lines[e.row])
			if dif > 0 {
				Left(dif, e)
			}
		}
	}
	return 1
}
