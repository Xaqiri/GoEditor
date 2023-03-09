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
// Put the ANSI escape codes into a struct or consts
// Look into ropes later

type Editor struct {
	prompt string
	lines  map[int]string
	writer io.Writer
	reader *bufio.Reader
	mode   string
	w      int
	h      int
	cx     int
	cy     int
}

func (e *Editor) initEditor() {
	e.prompt = strconv.Itoa(prompt) + "  "
	e.lines = map[int]string{}
	e.writer = os.Stdout
	e.reader = bufio.NewReader(os.Stdin)
	e.mode = "input"
	e.cx = 4
	e.cy = 1
	e.w, e.h, _ = term.GetSize(0)
}

func (e *Editor) updateEditor() {
	e.prompt = strconv.Itoa(prompt) + "  "
	e.cx = 4
	e.cy++
	y++
	x = 0

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
	e.lines[1] = line
	s, _ := term.MakeRaw(0)
	defer term.Restore(0, s)

	for {
		// Reverse bg and fg colors
		fmt.Fprintf(e.writer, "\x1b[7m")
		// Move cursor to the bottom left of the screen
		fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.h, 1)
		// Clear the line
		fmt.Fprintf(e.writer, "\x1b[2K")
		for i := 0; i < e.w; i++ {
			fmt.Print(" ")
		}
		fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.h, 1)
		fmt.Printf("%v", e.lines)
		// fmt.Printf("%d:%d", x, y)

		// Move the cursor to the far right minus five spaces
		fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.h, e.w-5)
		// Display the cursor's x and y positions
		fmt.Printf("%d:%d", e.cx, e.cy)
		// Reset the bg and fg colors
		fmt.Fprintf(e.writer, "\x1b[m")
		// Reset the cursor to the top left of the screen
		fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.cy, e.cx)
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
				Right(1, line, &e)
			} else if inp == 'i' {
				e.mode = "input"
				e.prompt = strconv.Itoa(prompt) + "  "
			}
		}
		if e.mode == "input" {

			line = e.lines[y]
			if x > len(line) {
				x = len(line)
				e.cx = x + 4
			}
			inp, _, _ := e.reader.ReadRune()
			if inp == '\033' {
				e.mode = "move"
			} else if inp == '\x0D' {
				prompt++
				e.updateEditor()
				line = ""
				e.lines[y] = line
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
					e.lines[y] = line
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
				e.lines[y] = line
				fmt.Fprintf(e.writer, "\x1b[2K")
				fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.cy, 1)
				fmt.Print(e.prompt)
				fmt.Print(line)
				Right(1, line, &e)
			}
		}
	}
}

func Up(n int, e *Editor) {
	fmt.Fprintf(e.writer, "\x1b[%dA", n)
	if y-n <= 0 {
		y = 1
		e.cy = 1
	} else {
		y -= n
		e.cy -= n
		prompt -= n
	}
	if x > len(e.lines[y]) {
		x = len(e.lines[y])
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
	if x > len(e.lines[y]) {
		x = len(e.lines[y])
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

func Right(n int, line string, e *Editor) {
	fmt.Fprintf(e.writer, "\x1b[%dC", n)
	x += n
	e.cx += n
	if x > len(e.lines[y]) {
		x = len(e.lines[y])
		e.cx = x + 4
	}
}
