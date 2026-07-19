package hotkey

import (
	"syscall"
	"unsafe"
)

var (
	procGetForegroundWindow  = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGetWindowThreadPID   = user32.NewProc("GetWindowThreadProcessId")
	procGetCurrentProcessId  = kernel32.NewProc("GetCurrentProcessId")
)

func foregroundAppName() string {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ""
	}
	buf := make([]uint16, 256)
	n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:n])
}

func isSelfFocused() bool {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}
	var pid uint32
	procGetWindowThreadPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	ourPid, _, _ := procGetCurrentProcessId.Call()
	return pid == uint32(ourPid)
}
