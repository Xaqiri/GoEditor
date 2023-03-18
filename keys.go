package main

const (
	ctrlA = iota + 1
	ctrlB
	ctrlC
	ctrlD
	ctrlE
	ctrlF
	ctrlG
	ctrlH
	ctrlI
	ctrlJ
	ctrlK
	ctrlL
	ctrlM
	ctrlN
	ctrlO
	ctrlP
	ctrlQ
	ctrlR
	ctrlS
	ctrlT
	ctrlU
	ctrlV
	ctrlW
	ctrlX
	ctrlY
	ctrlZ
)

type KeyCode struct {
	esc          byte
	cr           byte
	backspace    byte
	delete       byte
	tab          byte
	parenthesis  byte
	bracket      byte
	quote        byte
	dquote       byte
	backtick     byte
	brace        byte
	up           byte
	down         byte
	left         byte
	right        byte
	lineAbove    byte
	lineBelow    byte
	ctrlD, ctrlU byte
	brackets     map[byte]byte
}

func (k *KeyCode) initKeyCode() {
	k.esc = 27
	k.cr = 13
	k.backspace = 127
	k.delete = 120
	k.tab = 9
	k.ctrlD = ctrlD
	k.ctrlU = ctrlU
	k.parenthesis = 40
	k.bracket = 91
	k.quote = 39
	k.dquote = 34
	k.backtick = 96
	k.brace = 123
	k.up = 107
	k.down = 106
	k.left = 104
	k.right = 108
	k.lineAbove = 79
	k.lineBelow = 111
	k.brackets = map[byte]byte{
		'(': ')',
		'[': ']',
		'{': '}',
	}
}

func (k *KeyCode) matchingBrackets(left, right byte) bool {
	for _, v := range k.brackets {
		if v == right {
			return true
		}
	}
	return false
}
