//go:build !darwin && !windows

package hotkey

func foregroundAppName() string { return "" }

func isSelfFocused() bool { return false }
