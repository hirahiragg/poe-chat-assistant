package hotkey

import (
	"fmt"
	"strings"
)

func Parse(s string) (modifiers uint32, keyCode uint32, err error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, 0, fmt.Errorf("empty hotkey")
	}

	parts := strings.Split(s, "+")
	key := parts[len(parts)-1]
	mods := parts[:len(parts)-1]

	var m uint32
	for _, mod := range mods {
		switch strings.TrimSpace(mod) {
		case "ctrl", "control":
			m |= modControl
		case "shift":
			m |= modShift
		case "alt", "option":
			m |= modOption
		case "cmd", "command":
			m |= modCommand
		default:
			return 0, 0, fmt.Errorf("unknown modifier: %s", mod)
		}
	}

	kc, ok := keyCodes[strings.TrimSpace(key)]
	if !ok {
		return 0, 0, fmt.Errorf("unknown key: %s", key)
	}

	return m, kc, nil
}

const (
	modCommand uint32 = 0x0100
	modShift   uint32 = 0x0200
	modOption  uint32 = 0x0800
	modControl uint32 = 0x1000
)

var keyCodes = map[string]uint32{
	"a": 0x00, "s": 0x01, "d": 0x02, "f": 0x03,
	"h": 0x04, "g": 0x05, "z": 0x06, "x": 0x07,
	"c": 0x08, "v": 0x09, "b": 0x0B, "q": 0x0C,
	"w": 0x0D, "e": 0x0E, "r": 0x0F, "y": 0x10,
	"t": 0x11, "1": 0x12, "2": 0x13, "3": 0x14,
	"4": 0x15, "6": 0x16, "5": 0x17, "9": 0x19,
	"7": 0x1A, "8": 0x1C, "0": 0x1D, "o": 0x1F,
	"u": 0x20, "i": 0x22, "p": 0x23, "l": 0x25,
	"j": 0x26, "k": 0x28, "n": 0x2D, "m": 0x2E,
	"space": 0x31,
	"f1": 0x7A, "f2": 0x78, "f3": 0x63, "f4": 0x76,
	"f5": 0x60, "f6": 0x61, "f7": 0x62, "f8": 0x64,
	"f9": 0x65, "f10": 0x6D, "f11": 0x67, "f12": 0x6F,
}
