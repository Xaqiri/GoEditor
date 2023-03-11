package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"golang.org/x/term"
)

// TODO
// Too many variables are changing in too many places, fix that
// Split this up into multiple files
// Put the ANSI escape codes into the Editor struct
// Look into ropes later
// Add saving and loading files
// Add command mode after pressing :

type Editor struct {
	lines                  []string
	writer                 io.Writer
	reader                 *bufio.Reader
	w, h, col, row, cx, cy int
	prompt, mode           string
}

func (e *Editor) initEditor() {
	e.prompt = strconv.Itoa(prompt) + "  "
	e.lines = []string{""}
	e.writer = os.Stdout
	e.reader = bufio.NewReader(os.Stdin)
	e.mode = "input"
	e.cx = 4
	e.cy = 1
	e.row = len(e.lines)
	e.col = len(e.lines[e.row-1])
	y = len(e.lines)
	x = len(e.lines[y-1])
	e.w, e.h, _ = term.GetSize(0)
}

func (e *Editor) updateEditor() {
	e.updatePrompt()
	e.cx = 4
	e.cy++
	y = len(e.lines)
	x = len(e.lines[y-1])
}

func (e *Editor) updatePrompt() {
	padding := ""
	if prompt < 10 {
		padding = "  "
	} else {
		padding = " "
	}
	e.prompt = strconv.Itoa(prompt) + padding
}

func (e *Editor) setCursor() {
	if e.mode == "input" {
		fmt.Fprintf(e.writer, "\x1b[6 q")
	} else {
		fmt.Fprintf(e.writer, "\x1b[2 q")
	}
	e.row = e.cy
	e.col = len(e.lines[e.row-1])
}

func (e *Editor) insertLine(lineNumber int, line string) {
	temp := make([]string, len(e.lines)+1)
	copy(temp, e.lines[:lineNumber])
	temp[lineNumber] = line
	copy(temp[lineNumber+1:], e.lines[lineNumber:])
	e.lines = temp
}

func (e *Editor) refreshScreen() {
	// Reverse bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[7m")
	// Move cursor to the bottom left of the screen
	fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.h, 1)
	// Clear the line
	fmt.Fprintf(e.writer, "\x1b[2K")
	for i := 0; i < e.w; i++ {
		fmt.Print(" ")
	}
	// Reset the bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[m")
	// Draw placeholder icons on the left of the screen
	for i := len(e.lines) + 1; i < e.h; i++ {
		fmt.Fprintf(e.writer, "\x1b[%d;%dH", i, 1)
		fmt.Print("~")
	}
	// Move the cursor to the far right minus five spaces
	fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.h, e.w-5)
	// Reverse bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[7m")
	// Display the cursor's x and y positions
	fmt.Printf("%d:%d ", e.cx, e.cy)
	// Reset the bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[m")
	// Reset the cursor to the top left of the screen
	fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.cy, e.cx)
	e.setCursor()
}

var x, y = 0, 1
var prompt, height = 1, 1

func main() {
	var e Editor
	e.initEditor()

	fmt.Fprintf(e.writer, "\x1b[2J")
	// Move cursor to top left corner
	fmt.Fprintf(e.writer, "\x1b[1;1H")

	fmt.Print(e.prompt)
	line := ""
	e.lines[0] = line
	s, _ := term.MakeRaw(0)
	defer term.Restore(0, s)

	for {

		e.refreshScreen()
		if e.mode != "input" {
			inp, _, _ := e.reader.ReadRune()
			if inp == 'q' {
				fmt.Printf("\033c")
				fmt.Fprintf(e.writer, "\x1b[1;1H")
				break
			} else if inp == 'k' {
				Up(1, &e)
			} else if inp == 'j' {
				Down(1, &e)
			} else if inp == 'h' {
				Left(1, &e)
			} else if inp == 'l' {
				Right(1, &e)
			} else if inp == 'i' {
				e.mode = "input"
				e.updatePrompt()
				e.setCursor()
			}
		}
		if e.mode == "input" {
			line = e.lines[y-1]
			if x > len(line) {
				x = len(line)
				e.cx = x + 4
			}
			inp, _, _ := e.reader.ReadRune()
			if inp == '\033' { // Pressing escape
				e.mode = "move"
				e.setCursor()
			} else if inp == '\x0D' { // Pressing return
				prompt++
				// Split the line at the cursor
				// Part of the line after the cursor
				line = e.lines[e.row-1][e.cx-len(e.prompt)-1:]
				// Part of the line up to the cursor
				e.lines[e.row-1] = e.lines[e.row-1][:e.cx-len(e.prompt)-1]

				e.updateEditor()

				// Inserts the portion of the previous line after the cursor
				// onto the new line
				e.insertLine(e.row, line)

				// Redraw the screen
				fmt.Fprintf(e.writer, "\x1b[2J")
				for i := 1; i <= len(e.lines); i++ {
					y = i
					fmt.Fprintf(e.writer, "\x1b[%d;%dH", i, 1)
					if i < 10 {
						fmt.Print(strconv.Itoa(i) + "  ")
					} else {
						fmt.Print(strconv.Itoa(i) + " ")
					}
					fmt.Print(e.lines[i-1])
				}
				y = prompt
				fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.cy, 1)
				fmt.Print(e.prompt)
				fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.cy, e.cx)

			} else if inp == '\u007F' {
				if len(line) > 0 && x > 0 {
					if x < len(line) {
						left := line[:x-1]
						right := line[x:]
						line = left + right
					} else {
						line = line[:len(line)-1]
					}
					e.lines[y-1] = line
					Left(1, &e)
					fmt.Fprintf(e.writer, "\x1b[2K")
					fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.cy, 1)
					fmt.Print(e.prompt)
					fmt.Print(line)
				}
			} else {
				left := line[:x]
				right := line[x:]
				left += string(inp)
				line = left + right
				e.lines[y-1] = line
				fmt.Fprintf(e.writer, "\x1b[2K")
				fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.cy, 1)
				fmt.Print(e.prompt)
				fmt.Print(line)
				Right(1, &e)
			}
		}
	}
}

func Up(n int, e *Editor) {
	fmt.Fprintf(e.writer, "\x1b[%dA", n)
	if e.cy-n <= 0 {
		y = 1
		e.cy = 1
	} else {
		y -= n
		e.cy -= n
		prompt -= n
	}
	if x > len(e.lines[e.cy-1]) {
		x = len(e.lines[e.cy-1])
		e.cx = x + 4
	}
}

func Down(n int, e *Editor) {
	fmt.Fprintf(e.writer, "\x1b[%dB", n)
	if y+n > len(e.lines) {
		y = len(e.lines)
	} else {
		y += n
		e.cy += n
		prompt += n
	}
	if x > len(e.lines[y-1]) {
		x = len(e.lines[y-1])
		e.cx = x + 4
	}
}

func Left(n int, e *Editor) {
	fmt.Fprintf(e.writer, "\x1b[%dD", n)
	if x-n <= 0 {
		x = 0
		e.cx = 4
	} else {
		x -= n
		e.cx -= n
	}
}

func Right(n int, e *Editor) {
	fmt.Fprintf(e.writer, "\x1b[%dC", n)
	x += n
	e.cx += n
	if x > len(e.lines[y-1]) {
		x = len(e.lines[y-1])
		e.cx = x + 4
	}
}
