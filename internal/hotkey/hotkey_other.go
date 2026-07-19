//go:build !darwin && !windows

package hotkey

func Register(hotkeyStr string, onPress func()) error {
	return nil
}

func Unregister() {}
