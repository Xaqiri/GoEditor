package main

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// TODO
// Look into ropes later
// Move cursor related stuff to its own file

func main() {
	args := os.Args
	var e Editor
	var k KeyCode
	var f File
	e.initEditor()
	k.initKeyCode()

	if len(args) > 1 {
		f.init(args[1])
		f.open(&e)
	}
	if len(e.lines) == 0 {
		e.lines = append(e.lines, "")
	}
	s, _ := term.MakeRaw(0)
	defer term.Restore(0, s)

	for {
		e.refreshScreen()
		line := e.lines[e.row]
		inp, _ := e.reader.ReadByte()
		switch e.mode {
		case move:
			handleMoveInput(inp, &e, k)
		case command:
			switch inp {
			case k.esc:
				e.cmd = []string{"", ""}
				e.mode = move
			case k.cr:
				if e.cmd[1] == "q" {
					e.clearScreen()
					os.Exit(0)
				} else if e.cmd[1] == "w" {
					if f.name != "" {
						f.save(&e)
					} else {
						e.cmd[0] = "No file to save"
					}
				}
				e.cmd[1] = ""
			case k.backspace:
				if len(e.cmd[1]) > 0 {
					e.cmd[1] = e.cmd[1][:len(e.cmd[1])-1]
				}
			default:
				e.debug = []string{""}
				e.cmd[0] = ":"
				e.cmd[1] += string(inp)
			}
		case input:
			if inp == k.esc { // Pressing escape
				e.mode = move
				Left(1, &e)
				e.setCursorStyle()
			} else if inp == k.cr { // Pressing return
				// Split the line at the cursor
				left := e.lines[e.row][:e.col]  // Part of the line up to the cursor
				right := e.lines[e.row][e.col:] // Part of the line after the cursor
				e.lines[e.row] = left           // Current row will contain characters up to the cursor
				e.insertLine(e.row+1, right)    // Add a new line below with the rest of the characters
				Down(1, &e)                     // Move down to the new line
				if len(left) > 0 && len(right) > 0 && k.matchingBrackets(left[len(left)-1], right[0]) {
					e.lines[e.row] = e.tab
					e.insertLine(e.row+1, right)
					Right(e.tabWidth, &e)
				}
			} else if inp == k.backspace {
				deleteChar(&e, k, 1, k.backspace)
			} else {
				dif := 1
				// Typing new characters
				left := line[:e.col]
				newChars := ""
				right := line[e.col:]
				switch inp {
				case k.parenthesis:
					newChars = "()"
				case k.bracket:
					newChars = "[]"
				case k.brace:
					newChars = "{}"
				case k.quote:
					newChars = "''"
				case k.dquote:
					newChars = "\"\""
				case k.backtick:
					newChars = "``"
				case k.tab:
					newChars = e.tab
					dif = e.tabWidth
				default:
					newChars = string(inp)
				}
				e.lines[e.row] = left + newChars + right
				Right(dif, &e)
			}
		}
	}
}

func deleteChar(e *Editor, k KeyCode, num int, key byte) {
	line := e.lines[e.row]
	if e.col-num >= 0 && key == k.backspace {
		Left(num, &*e)
		line = line[:e.col] + line[e.col+num:]
	} else if e.col+num <= len(line) && key == 'x' {
		line = line[:e.col] + line[e.col+num:]
	}
	e.lines[e.row] = line
}

func Up(n int, e *Editor) {
	if e.cy-n >= e.document.t {
		e.moveDocCursor(e.cx, e.cy-n)
	} else if e.cy-n < e.document.t && e.row >= e.document.t {
		e.offset -= n
		if e.offset < 0 {
			e.offset = 0
		}
		e.moveDocCursor(e.cx, e.cy)
	}
}

func Down(n int, e *Editor) {
	if e.cy+n <= e.document.h && e.cy+n < len(e.lines)-e.offset {
		e.moveDocCursor(e.cx, e.cy+n)
	} else if e.cy+n > e.document.h && e.row < len(e.lines)-1 {
		e.offset += n
		if e.offset+e.document.h > len(e.lines) {
			e.offset = len(e.lines) - e.document.h
		}
		e.moveDocCursor(e.cx, e.cy)
	}
}

func Left(n int, e *Editor) {
	if e.cx-n > e.document.l {
		e.moveDocCursor(e.cx-n, e.cy)
	} else {
		e.moveDocCursor(e.document.l, e.cy)
	}
}

func Right(n int, e *Editor) {
	if e.col < len(e.lines[e.row]) {
		e.moveDocCursor(e.cx+n, e.cy)
	}
}

func handleMoveInput(inp byte, e *Editor, k KeyCode) {
	if inp == k.up {
		Up(1, e)
	} else if inp == k.down {
		Down(1, e)
	} else if inp == k.left {
		Left(1, e)
	} else if inp == k.right {
		Right(1, e)
	} else if inp == 48 {
		Left(len(e.lines[e.row]), e)
	} else if inp == 36 {
		Right(len(e.lines[e.row]), e)
	} else if inp == 'w' {
		line := strings.Split(strings.Trim(e.lines[e.row][e.col:], " "), " ")
		Right(len(line[0])+1, e)
	} else if inp == 'b' {
		line := strings.Split(strings.Trim(e.lines[e.row][:e.col], " "), " ")
		Left(len(line[len(line)-1])+1, e)
	} else if inp == ':' {
		e.cmd[0] = ":"
		e.mode = command
	} else if inp == 'a' {
		e.mode = input
		e.setCursorStyle()
		Right(1, e)
	} else if inp == 'i' {
		e.mode = input
		e.setCursorStyle()
	} else if inp == 'o' {
		e.insertLine(e.row+1, "")
		Down(1, e)
		e.mode = input
	} else if inp == 'O' {
		Left(e.col, e)
		e.insertLine(e.row, e.lines[e.row])
		e.lines[e.row] = ""
		e.mode = input
	} else if inp == k.delete {
		deleteChar(e, k, 1, k.delete)
	} else if inp == 'd' {
		inp, _ = e.reader.ReadByte()
		if inp == 'd' {
			// Pressing d twice deletes the current line
			if e.row < len(e.lines)-1 {
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
	} else if inp == 'g' {
		Up(e.row, e)
		if e.row != 1 {
			Up(e.row, e)
		}
	} else if inp == 'G' {
		Down(len(e.lines), e)
		if e.row != len(e.lines)-1 {
			Down(len(e.lines)-e.row-1, e)
		}

	} else if inp == k.ctrlU {
		if e.offset > e.document.h {
			e.offset -= e.document.h
		} else {
			e.offset = 0
			e.moveCursor(e.document.l, 1)
		}
	} else if inp == k.ctrlD {
		if len(e.lines) > e.document.h {
			e.offset += e.document.h
			if e.offset > len(e.lines)-e.document.h {
				e.offset = len(e.lines) - e.document.h
				e.moveCursor(e.document.l, e.document.h)
			}
		}
	}
}
