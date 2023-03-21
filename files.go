package main

import (
	"bufio"
	"os"
	"strings"
)

type File struct {
	name     string
	filetype string
	changed  bool
}

func (f *File) init(fn string) {
	f.name = fn
	f.changed = false
	temp := strings.Split(fn, ".")
	if len(temp) > 1 && (temp[1] == "go" || temp[1] == "txt") {
		f.filetype = temp[1]
	}
}

func (f *File) save(e *Editor) {
	file, err := os.Create(f.name)
	defer file.Close()
	check(err)
	e.fileInfo[0] = f.name
	writer := bufio.NewWriter(file)
	for _, v := range e.lines {
		_, err := writer.WriteString(v + "\n")
		check(err)
	}
	writer.Flush()
	e.fileInfo[1] = "saved"
}

func (f *File) open(e *Editor) {
	file, err := os.Open(f.name)
	defer file.Close()
	check(err)
	e.fileInfo[0] = f.name
	e.fileInfo[2] = f.filetype
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
