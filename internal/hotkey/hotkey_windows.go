package hotkey

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procGetMessage       = user32.NewProc("GetMessageW")

	mu       sync.Mutex
	callback func()
	stopCh   chan struct{}
)

const (
	wmHotkey   = 0x0312
	modWinAlt  = 0x0001
	modWinCtrl = 0x0002
	modWinShft = 0x0004
)

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      [2]int32
}

var winKeyCodes = map[string]uint32{
	"a": 0x41, "b": 0x42, "c": 0x43, "d": 0x44,
	"e": 0x45, "f": 0x46, "g": 0x47, "h": 0x48,
	"i": 0x49, "j": 0x4A, "k": 0x4B, "l": 0x4C,
	"m": 0x4D, "n": 0x4E, "o": 0x4F, "p": 0x50,
	"q": 0x51, "r": 0x52, "s": 0x53, "t": 0x54,
	"u": 0x55, "v": 0x56, "w": 0x57, "x": 0x58,
	"y": 0x59, "z": 0x5A,
	"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33,
	"4": 0x34, "5": 0x35, "6": 0x36, "7": 0x37,
	"8": 0x38, "9": 0x39,
	"space": 0x20,
	"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73,
	"f5": 0x74, "f6": 0x75, "f7": 0x76, "f8": 0x77,
	"f9": 0x78, "f10": 0x79, "f11": 0x7A, "f12": 0x7B,
}

func parseWindows(s string) (mods, vk uint32, err error) {
	macMods, macKey, err := Parse(s)
	if err != nil {
		return 0, 0, err
	}

	// Parse() で得た macOS keyCode をキー名に逆引きして Windows VK に変換
	var keyName string
	for name, code := range keyCodes {
		if code == macKey {
			keyName = name
			break
		}
	}
	if keyName == "" {
		return 0, 0, fmt.Errorf("unsupported key code: %d", macKey)
	}

	wvk, ok := winKeyCodes[keyName]
	if !ok {
		return 0, 0, fmt.Errorf("unsupported key on Windows: %s", keyName)
	}

	var wmods uint32
	if macMods&modControl != 0 {
		wmods |= modWinCtrl
	}
	if macMods&modShift != 0 {
		wmods |= modWinShft
	}
	if macMods&modOption != 0 || macMods&modCommand != 0 {
		wmods |= modWinAlt
	}

	return wmods, wvk, nil
}

func Register(hotkeyStr string, onPress func()) error {
	mods, vk, err := parseWindows(hotkeyStr)
	if err != nil {
		return err
	}

	Unregister()

	mu.Lock()
	callback = onPress
	stopCh = make(chan struct{})
	mu.Unlock()

	ret, _, _ := procRegisterHotKey.Call(0, 1, uintptr(mods), uintptr(vk))
	if ret == 0 {
		return fmt.Errorf("RegisterHotKey failed")
	}

	go messageLoop()
	return nil
}

func messageLoop() {
	mu.Lock()
	ch := stopCh
	mu.Unlock()

	var m msg
	for {
		select {
		case <-ch:
			return
		default:
		}
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 {
			return
		}
		if m.message == wmHotkey {
			mu.Lock()
			cb := callback
			mu.Unlock()
			if cb != nil {
				cb()
			}
		}
	}
}

func Unregister() {
	mu.Lock()
	if stopCh != nil {
		close(stopCh)
		stopCh = nil
	}
	callback = nil
	mu.Unlock()

	procUnregisterHotKey.Call(0, 1)
}
