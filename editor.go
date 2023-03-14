package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

type Editor struct {
	lines            []string
	writer           io.Writer
	reader           *bufio.Reader
	w, h             int // Width and height of terminal
	col, row         int // Width of current row and max number of rows
	cx, cy           int // Cursor position
	offset, tabWidth int
	prompt, mode     string
	lineNumWidth     int
	debug            []string
	cmd              []string
	ansiCodes        map[string][]byte
	keywords         []string
}

func (e *Editor) initEditor() {

	e.keywords = []string{"for", "func", "if", "else", "return", "package", "import", "switch", "case", "var"}
	e.ansiCodes = map[string][]byte{
		"escape":    {'\033'},
		"return":    {'\x0D'},
		"backspace": {'\u007F'},
		"clear":     {'\033', '[', '2', 'J'},
		"move":      {'\033', '[', ' ', ';', ' ', 'H'},
	}
	e.lines = []string{}
	e.cy = 1
	e.updatePrompt()
	e.lineNumWidth = len(e.prompt) + 1
	e.writer = os.Stdout
	e.reader = bufio.NewReader(os.Stdin)
	e.mode = "move"
	e.cx = e.lineNumWidth
	e.row = e.cy - 1
	e.col = e.cx - e.lineNumWidth
	e.w, e.h, _ = term.GetSize(0)
	e.offset = 0
	e.cmd = []string{"", ""}
	e.debug = []string{""}
	e.tabWidth = 4
}

func (e *Editor) moveCursor(row, col int) {
	fmt.Fprintf(e.writer, "\033[%d;%dH", row, col)
	e.cx = col
	e.cy = row
	e.row = e.cy - 1 + e.offset
	e.col = e.cx - e.lineNumWidth
}

func (e *Editor) updatePrompt() {
	p := ""
	if e.row < len(e.lines) {
		p = strconv.Itoa(e.row + 1)
	} else {
		p = "~"
	}
	for len(p) < 4 {
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
	e.moveCursor(e.cy, e.cx-e.col)
}

func (e *Editor) clearScreen() {
	// Clear screen
	fmt.Fprintf(e.writer, string(e.ansiCodes["clear"]))
	e.moveCursor(1, 1)
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
	for i := 0; i < e.h; i++ {
		e.moveCursor(i+1, 1)
		e.updatePrompt()
		fmt.Print(e.prompt)
	}
}

func contains(words []string, word string) bool {
	for _, i := range words {
		if i == word {
			return true
		}
	}
	return false
}

func (e *Editor) drawDocument() {
	drawHeight := 0
	if len(e.lines) > e.h {
		drawHeight = e.h - 1
	} else {
		drawHeight = len(e.lines)
	}
	for i := 0; i < drawHeight; i++ {
		e.moveCursor(i+1, e.lineNumWidth)
		line := strings.Split(e.lines[e.row], " ")
		for _, s := range line {
			if contains(e.keywords, s) {
				fmt.Fprintf(e.writer, "\x1b[34m%s\x1b[m ", s)
			} else {
				fmt.Print(s, " ")
			}
		}
	}
}

func (e *Editor) drawBottomInfo(x, y int) {
	e.moveCursor(y, x)
	btm := ""
	mode := " " + e.mode
	cmd := e.cmd[0] + e.cmd[1]
	coord := strconv.Itoa(e.col) + ":" + strconv.Itoa(e.cy)
	// Reverse bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[7m")
	// Move cursor to the bottom left of the screen
	fmt.Fprintf(e.writer, "\x1b[%d;%dH", e.h, 1)
	// Clear the line
	fmt.Fprintf(e.writer, "\x1b[2K")
	btm += strings.Join(e.debug, " ")
	btm += cmd
	btm += mode
	for i := len(btm); i < e.w-len(coord); i++ {
		btm += " "
	}
	btm += coord
	fmt.Print(btm)
	// Reset the bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[m")
	// e.debug = []string{""}
}

func (e *Editor) save(fn string) {
	file, err := os.Create(fn)
	defer file.Close()
	check(err)
	writer := bufio.NewWriter(file)
	for _, v := range e.lines {
		_, err := writer.WriteString(v + "\n")
		check(err)
	}
	writer.Flush()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func open(fn string, e *Editor) {
	tab := ""
	for i := 0; i < e.tabWidth; i++ {
		tab += " "
	}
	file, err := os.Open(fn)
	defer file.Close()
	check(err)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()
		if len(s) > 0 {
			s = strings.ReplaceAll(scanner.Text(), string('\t'), tab)
		}
		e.lines = append(e.lines, s)
	}
}
