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
	offset, tabWidth int // Number of lines offscreen, width of a tab in spaces
	prompt, mode     string
	lineNums         LineNumSection
	document         TextDocSection
	infoBar          InfoBarSection
	debug            []string
	cmd              []string
	ansiCodes        map[string][]byte
	keywords         []string
	fileName         string
}

type LineNumSection struct {
	w, h int
	l, t int
}

type InfoBarSection struct {
	w, h int
	l, t int
	cmd  string
	pos  string
	mode string
}

type TextDocSection struct {
	w, h int
	l, t int
}

func (e *Editor) initEditor() {
	e.w, e.h, _ = term.GetSize(0)
	e.infoBar = InfoBarSection{w: e.w, h: 2, l: 1, t: e.h - 1}
	e.lineNums = LineNumSection{w: 4, h: e.h - 1, l: 1, t: 1}
	e.document = TextDocSection{w: e.w - e.lineNums.w, h: e.h - e.infoBar.h, l: e.lineNums.w + 1, t: 1}

	e.keywords = []string{"for", "func", "if", "else", "return", "package", "import", "switch", "case", "var"}
	e.ansiCodes = map[string][]byte{
		"escape":    {'\033'},
		"return":    {'\x0D'},
		"backspace": {'\u007F'},
		"clear":     {'\033', '[', '2', 'J'},
		"move":      {'\033', '[', ' ', ';', ' ', 'H'},
		"hide":      {'\033', '[', '?', '2', '5', 'l'},
		"show":      {'\033', '[', '?', '2', '5', 'h'},
	}
	e.lines = []string{}
	e.cx, e.cy = e.document.l, e.document.t
	e.writer = os.Stdout
	e.reader = bufio.NewReader(os.Stdin)
	e.mode = "move"
	e.row = 0
	e.col = 0
	e.offset = 0
	e.cmd = []string{"", ""}
	e.tabWidth = 4
}

func (e *Editor) hideCursor() {
	fmt.Fprintf(e.writer, string(e.ansiCodes["hide"]))
}

func (e *Editor) showCursor() {
	fmt.Fprintf(e.writer, string(e.ansiCodes["show"]))
}

func (e *Editor) moveCursor(col, row int) {
	fmt.Fprintf(e.writer, "\033[%d;%dH", row, col)
	e.cx = col
	e.cy = row
	e.row = e.cy + e.offset - e.document.t
	e.col = e.cx - e.document.l
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
	e.moveCursor(e.document.l, e.cy)
}

func (e *Editor) clearScreen() {
	fmt.Fprintf(e.writer, string(e.ansiCodes["clear"]))
	e.moveCursor(1, 1)
}

func (e *Editor) refreshScreen() {
	x, y := e.cx, e.cy
	e.clearScreen()
	e.drawLineNums()
	e.drawDocument()
	e.drawBottomInfo()
	e.setCursorStyle()
	e.moveCursor(x, y)
}

func (e *Editor) drawLineNums() {
	for i := 1; i < e.lineNums.h; i++ {
		e.moveCursor(1, i)
		if i <= len(e.lines)-e.offset {
			fmt.Print(i + e.offset)
		} else {
			fmt.Print("~")
		}
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
	if len(e.lines) > e.document.h {
		drawHeight = e.document.h
	} else {
		drawHeight = len(e.lines)
	}
	if len(e.lines)-e.offset < drawHeight {
		drawHeight = len(e.lines) - e.offset
	}
	for i := 1; i <= drawHeight; i++ {
		e.moveCursor(e.document.l, i)
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

func (e *Editor) drawBottomInfo() {
	e.moveCursor(1, e.infoBar.t)
	bg := ""
	e.infoBar.mode = " + "
	e.infoBar.cmd = strings.Join(e.cmd, "")
	if e.mode == "command" {
		bg = "\x1b[41m"
	} else if e.mode == "input" {
		bg = "\x1b[42m"
	} else {
		bg = "\x1b[46m"
		e.infoBar.cmd = e.fileName
	}
	e.infoBar.pos = strings.Join([]string{strconv.Itoa(e.col), ":", strconv.Itoa(e.row)}, "")
	// Clear the line
	fmt.Fprintf(e.writer, "\x1b[2K")

	fmt.Fprintf(e.writer, bg)
	fmt.Fprintf(e.writer, "\x1b[30m")
	fmt.Print(e.infoBar.mode)

	fmt.Fprintf(e.writer, "\x1b[m")
	// Reverse bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[7m")

	for i := len(e.infoBar.mode); i < e.infoBar.w-len(e.infoBar.pos); i++ {
		fmt.Print(" ")
	}
	fmt.Print(e.infoBar.pos)

	// Reset the bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[m")
	e.moveCursor(1, e.infoBar.t+1)
	fmt.Print(e.infoBar.cmd)
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

func (e *Editor) open(fn string) {
	tab := ""
	for i := 0; i < e.tabWidth; i++ {
		tab += " "
	}
	file, err := os.Open(fn)
	defer file.Close()
	check(err)
	e.fileName = fn
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()
		if len(s) > 0 {
			s = strings.ReplaceAll(scanner.Text(), string('\t'), tab)
		}
		e.lines = append(e.lines, s)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
