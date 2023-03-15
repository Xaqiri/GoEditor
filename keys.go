package main

type KeyCode struct {
	esc       byte
	cr        byte
	backspace byte
	ctrlD     byte
	ctrlU     byte
}

func (k *KeyCode) initKeyCode() {
	k.esc = 27
	k.cr = 13
	k.backspace = 8
	k.ctrlD = 4
	k.ctrlU = 21
}
