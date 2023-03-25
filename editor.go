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

type mode int

const (
	input = iota
	move
	command
	search
	visual
)

type Editor struct {
	lines            []string
	writer           io.Writer
	reader           *bufio.Reader
	w, h             int // Width and height of terminal
	col, row         int // Width of current row and max number of rows
	cx, cy           int // Cursor position
	offset, tabWidth int // Number of lines offscreen, width of a tab in spaces
	mode             int
	lineNums         LineNumSection
	document         TextDocSection
	infoBar          InfoBarSection
	cmd              []string
	ansiCodes        map[string][]byte
	fileInfo         []string
	tab              string
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
	w, h   int
	l, t   int
	cx, cy int
}

func (e *Editor) initEditor() {
	e.w, e.h, _ = term.GetSize(0)
	e.infoBar = InfoBarSection{w: e.w, h: 2, l: 1, t: e.h - 1, pos: strconv.Itoa(e.row)}
	e.lineNums = LineNumSection{w: 6, h: e.h - 1, l: 1, t: 1}
	e.document = TextDocSection{w: e.w - e.lineNums.w, h: e.h - e.infoBar.h, l: e.lineNums.w + 1, t: 1, cx: e.lineNums.w + 1, cy: 1}
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
	e.mode = move
	e.offset = 0
	e.row = e.cy + e.offset - e.document.t
	e.col = 0
	e.cmd = []string{":", ""}
	e.tabWidth = 4
	e.tab = ""
	for i := 0; i < e.tabWidth; i++ {
		e.tab += " "
	}
	e.fileInfo = []string{
		"", // File name
		"", // File status
		"", // File type
	}
}

func (e *Editor) scroll(num int) {
	fmt.Fprintf(e.writer, "\033[S")
	e.offset += num
	if e.offset < 0 {
		e.offset = 0
	}
	if e.offset > len(e.lines)-e.document.h {
		e.offset = len(e.lines) - e.document.h
	}
}

func (e *Editor) hideCursor() {
	fmt.Fprintf(e.writer, string(e.ansiCodes["hide"]))
}

func (e *Editor) showCursor() {
	fmt.Fprintf(e.writer, string(e.ansiCodes["show"]))
}

func (e *Editor) setCursorStyle() {
	if e.mode == input {
		// Line
		fmt.Fprintf(e.writer, "\x1b[6 q")
	} else {
		// Block
		fmt.Fprintf(e.writer, "\x1b[2 q")
	}
}

func (e *Editor) clearScreen() {
	fmt.Fprintf(e.writer, string(e.ansiCodes["clear"]))
	e.moveCursor(1, 1)
}

func (e *Editor) refreshScreen() {
	e.hideCursor()
	e.infoBar.pos = strconv.Itoa(e.col+1) + ":" + strconv.Itoa(e.row+1) + ":" + strconv.Itoa(len(e.lines))
	e.drawLineNums()
	e.drawDocument()
	e.drawBottomInfo()
	e.setCursorStyle()
	if e.mode == command || e.mode == search {
		e.moveCursor(e.infoBar.l+len(e.cmd[0])+len(e.cmd[1]), e.h)
	} else {
		e.moveCursor(e.document.cx, e.document.cy)
	}
	e.showCursor()
}

func (e *Editor) moveCursor(col, row int) {
	fmt.Fprintf(e.writer, "\033[%d;%dH", row, col)
	e.cx = col
	e.cy = row
}

func (e *Editor) moveDocCursor(col, row int) {
	dif := 0
	e.row = row + e.offset - e.document.t
	e.col = col - e.document.l
	if e.row < len(e.lines) && e.col > len(e.lines[e.row]) {
		dif = e.col - len(e.lines[e.row])
		e.col -= dif
		col = e.col + e.document.l
	}
	e.document.cx, e.document.cy = col, row
	e.moveCursor(col, row)
}

func (e *Editor) drawLineNums() {
	num := ""
	for i := 1; i < e.lineNums.h; i++ {
		e.moveCursor(1, i)
		numlen := len(strconv.Itoa(i + e.offset))
		if i <= len(e.lines)-e.offset {
			if numlen < e.lineNums.w {
				for j := 0; j < e.lineNums.w-numlen-1; j++ {
					num += " "
					// fmt.Fprintf(e.writer, "\u256C")
				}
			}
			num += strconv.Itoa(i + e.offset)
			fmt.Print(num)
			num = ""
		} else {
			fmt.Print("    ~")
		}
		// fmt.Printf("\x1b(0\x78")
		// fmt.Printf("\x1b(B")
		fmt.Fprintf(e.writer, "\u2502")
		fmt.Fprintf(e.writer, "\x1b[K")
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

func (e *Editor) insertLine(lineNumber int, line string) {
	temp := make([]string, len(e.lines)+1)
	copy(temp, e.lines[:lineNumber])
	temp[lineNumber] = line
	copy(temp[lineNumber+1:], e.lines[lineNumber:])
	e.lines = temp
	e.moveDocCursor(e.document.l, e.cy)
}

func (e *Editor) drawDocument() {
	x, y := e.document.cx, e.document.cy
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
		e.moveDocCursor(e.document.l, i)
		comment := strings.Split(e.lines[e.row], "//")
		if len(comment) > 1 {
			comment = comment[1:]
		}
		line := strings.Split(e.lines[e.row], " ")
		for _, s := range line {
			if s == "//" {
				fmt.Fprintf(e.writer, "\x1b[32m%s%s\x1b[m ", s, string(comment[0]))
				break
			} else {
				fmt.Print(s, " ")
			}
		}
		fmt.Fprintf(e.writer, "\033[K")
	}
	e.moveDocCursor(x, y)
}

func (e *Editor) drawBottomInfo() {
	e.moveCursor(1, e.infoBar.t)
	bg := ""
	modeStr := []string{" ", "", " "}
	if e.mode == input {
		modeStr[1] = "i"
	} else if e.mode == move {
		modeStr[1] = "m"
	} else if e.mode == command {
		modeStr[1] = "c"
	} else if e.mode == search {
		modeStr[1] = "s"
	}
	e.infoBar.mode = strings.Join(modeStr, "")
	if e.mode == command || e.mode == search {
		bg = "\x1b[41m"
	} else if e.mode == input {
		bg = "\x1b[42m"
	} else {
		bg = "\x1b[46m"
	}

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

	e.infoBar.pos = ""
	// Reset the bg and fg colors
	fmt.Fprintf(e.writer, "\x1b[m")
	e.moveCursor(1, e.infoBar.t+1)
	if e.mode == command || e.mode == search {
		fmt.Print(strings.Join(e.cmd, ""))
		fmt.Fprintf(e.writer, "\033[K")
	} else {
		fmt.Print(strings.Join(e.fileInfo, " "))
		fmt.Fprintf(e.writer, "\033[K")
	}
}
