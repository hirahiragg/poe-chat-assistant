package hotkey

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Carbon

#include <Carbon/Carbon.h>

extern void goHotkeyPressed(void);

static EventHotKeyRef gHotKeyRef = NULL;
static EventHandlerRef gHandlerRef = NULL;

static OSStatus hotKeyHandler(EventHandlerCallRef nextHandler, EventRef event, void *userData) {
	(void)nextHandler;
	(void)event;
	(void)userData;
	goHotkeyPressed();
	return noErr;
}

static int registerHotkey(UInt32 keyCode, UInt32 modifiers) {
	if (gHandlerRef == NULL) {
		EventTypeSpec eventType = {kEventClassKeyboard, kEventHotKeyPressed};
		OSStatus err = InstallApplicationEventHandler(&hotKeyHandler, 1, &eventType, NULL, &gHandlerRef);
		if (err != noErr) return (int)err;
	}
	if (gHotKeyRef != NULL) {
		UnregisterEventHotKey(gHotKeyRef);
		gHotKeyRef = NULL;
	}
	EventHotKeyID hotKeyID = {'POEC', 1};
	OSStatus err = RegisterEventHotKey(keyCode, modifiers, hotKeyID, GetApplicationEventTarget(), 0, &gHotKeyRef);
	return (int)err;
}

static void unregisterHotkey(void) {
	if (gHotKeyRef != NULL) {
		UnregisterEventHotKey(gHotKeyRef);
		gHotKeyRef = NULL;
	}
}
*/
import "C"

import (
	"fmt"
	"sync"
)

var (
	mu       sync.Mutex
	callback func()
)

//export goHotkeyPressed
func goHotkeyPressed() {
	mu.Lock()
	cb := callback
	mu.Unlock()
	if cb != nil {
		cb()
	}
}

func Register(hotkeyStr string, onPress func()) error {
	mods, key, err := Parse(hotkeyStr)
	if err != nil {
		return err
	}

	Unregister()

	mu.Lock()
	callback = onPress
	mu.Unlock()

	result := C.registerHotkey(C.UInt32(key), C.UInt32(mods))
	if result != 0 {
		return fmt.Errorf("register hotkey failed: OSStatus %d", result)
	}
	return nil
}

func Unregister() {
	C.unregisterHotkey()
	mu.Lock()
	callback = nil
	mu.Unlock()
}
