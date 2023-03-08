package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var target io.Writer = os.Stdout
var height = 1

func main() {
	mode := "move"
	b := bufio.NewReader(os.Stdin)
	s, _ := term.MakeRaw(0)
	defer term.Restore(0, s)
	for {
		if mode != "input" {
			inp, _, _ := b.ReadRune()
			if inp == 'q' {
				fmt.Printf("\033c")
				break
			} else if inp == 'k' {
				Up(1)
			} else if inp == 'j' {
				Down(1)
			} else if inp == 'h' {
				Left(1)
			} else if inp == 'l' {
				Right(1)
			} else if inp == 'i' {
				mode = "input"
				term.Restore(0, s)
			}
		}
		if mode == "input" {
			var inp string
			fmt.Scan(&inp)

			if strings.Contains(inp, "\\") {
				mode = "move"
				s, _ = term.MakeRaw(0)
			}
		}
	}
}

func Up(n int) {
	fmt.Fprintf(target, "\x1b[%dA", n)
	height += n
}
func Down(n int) {
	fmt.Fprintf(target, "\x1b[%dB", n)
	if height-n <= 0 {
		height = 0
	} else {
		height -= n
	}
}

func Left(n int) {
	fmt.Fprintf(target, "\x1b[%dD", n)

}

func Right(n int) {
	fmt.Fprintf(target, "\x1b[%dC", n)

}
