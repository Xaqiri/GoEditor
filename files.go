package main

import (
	"bufio"
	"os"
	"strings"
)

type File struct {
	name    string
	changed bool
}

func (f *File) init(fn string) {
	f.name = fn
	f.changed = false
}

func (f *File) save(e *Editor) {
	file, err := os.Create(f.name)
	defer file.Close()
	check(err)
	writer := bufio.NewWriter(file)
	for _, v := range e.lines {
		_, err := writer.WriteString(v + "\n")
		check(err)
	}
	writer.Flush()
	e.cmd[0] = f.name + " saved"
}

func (f *File) open(e *Editor) {
	file, err := os.Open(f.name)
	defer file.Close()
	check(err)
	e.fileName = f.name
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := scanner.Text()
		if len(s) > 0 {
			s = strings.ReplaceAll(scanner.Text(), string('\t'), e.tab)
		}
		e.lines = append(e.lines, s)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
