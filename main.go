package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// TODO
// Fix bug with prompt and height
// Too many variables are changing in too many places, fix that
// Split this up into multiple files
// Put the ANSI escape codes into a struct or consts
// Need a better way to store and edit lines, possibly use a map
// Look into ropes later

var target io.Writer = os.Stdout

var x, y, cx, cy = 0, 0, 4, 1
var prompt, height = 1, 1

func main() {
	fmt.Fprintf(target, "\x1b[2J")
	fmt.Fprintf(target, "\x1b[1;1H")

	fmt.Print(prompt)
	lines := []string{}
	var line []rune
	mode := "move"
	b := bufio.NewReader(os.Stdin)
	s, _ := term.MakeRaw(0)
	defer term.Restore(0, s)
	for {
		fmt.Fprintf(target, "\x1b[%d;%dH", cy, cx)
		if mode != "input" {
			inp, _, _ := b.ReadRune()
			if inp == 'q' {
				fmt.Printf("\033c")
				for i := 0; i < len(lines); i++ {
					fmt.Println(lines[i])
				}
				break
			} else if inp == 'k' {
				Up(1)
			} else if inp == 'j' {
				if cy < height {
					Down(1)
				}
			} else if inp == 'h' {
				Left(1)
			} else if inp == 'l' {
				Right(1)
			} else if inp == 'i' {
				mode = "input"
			}
		}
		if mode == "input" {
			inp, _, _ := b.ReadRune()
			if inp == '\033' {
				mode = "move"
			} else if inp == '\r' {
				if prompt == height {
					height++
				}
				lines = append(lines, string(line))
				line = []rune{}
				x = 0
				cx = 4
				Down(1)
				fmt.Fprintf(target, "\x1b[G")
				fmt.Print(prompt)
			} else if inp == '\u007F' {
				line = backspace(line)
			} else {
				fmt.Print(string(inp))
				line = append(line, inp)
				x++
				cx++
			}
		}
	}
}

func backspace(line []rune) []rune {

	if len(line) >= 1 {
		line = line[:len(line)-1]
	}
	fmt.Fprintf(target, "\x1b[2K")
	fmt.Fprintf(target, "\x1b[G")
	fmt.Print(prompt)
	fmt.Print(string(line))
	Left(1)
	cx = len(string(line)) + 4
	x = len(string(line))

	fmt.Fprintf(target, "\x1b[%d;%dH", cy, cx)

	return line
}

func Up(n int) {
	fmt.Fprintf(target, "\x1b[%dA", n)
	if y-n <= 0 {
		y = 0
		cy = 1
		prompt = 1
	} else {
		y -= n
		cy -= n
		prompt -= n
	}
	fmt.Print(cy, height)
}

func Down(n int) {
	fmt.Fprintf(target, "\x1b[%dB", n)
	y += n
	cy += n
	prompt += n
	fmt.Print(cy, height)
}

func Left(n int) {
	fmt.Fprintf(target, "\x1b[%dD", n)
	if x-n <= 0 {
		x = 0
		cx = 4
	} else {
		x -= n
		cx -= n
	}

}

func Right(n int) {
	fmt.Fprintf(target, "\x1b[%dC", n)
	x += n
	cx += n
}
