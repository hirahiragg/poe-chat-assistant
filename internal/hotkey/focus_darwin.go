package hotkey

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit

#import <AppKit/AppKit.h>

static const char* foregroundAppName(void) {
	@autoreleasepool {
		NSRunningApplication *app = [[NSWorkspace sharedWorkspace] frontmostApplication];
		NSString *name = [app localizedName];
		if (name == nil) return "";
		return strdup([name UTF8String]);
	}
}

static int isSelfForeground(void) {
	@autoreleasepool {
		NSRunningApplication *front = [[NSWorkspace sharedWorkspace] frontmostApplication];
		NSRunningApplication *self = [NSRunningApplication currentApplication];
		return [front isEqual:self] ? 1 : 0;
	}
}
*/
import "C"
import "unsafe"

func foregroundAppName() string {
	cs := C.foregroundAppName()
	if cs == nil {
		return ""
	}
	s := C.GoString(cs)
	C.free(unsafe.Pointer(cs))
	return s
}

func isSelfFocused() bool {
	return C.isSelfForeground() == 1
}
