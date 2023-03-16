package main

type KeyCode struct {
	esc         byte
	cr          byte
	backspace   byte
	tab         byte
	ctrlD       byte
	ctrlU       byte
	parenthesis byte
	bracket     byte
	quote       byte
	dquote      byte
	backtick    byte
	brace       byte
}

func (k *KeyCode) initKeyCode() {
	k.esc = 27
	k.cr = 13
	k.backspace = 127
	k.tab = 9
	k.ctrlD = 4
	k.ctrlU = 21
	k.parenthesis = 40
	k.bracket = 91
	k.quote = 39
	k.dquote = 34
	k.backtick = 96
	k.brace = 123
}
