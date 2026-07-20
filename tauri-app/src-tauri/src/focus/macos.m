#import <AppKit/AppKit.h>
#include <stdlib.h>

const char* foreground_app_name(void) {
    @autoreleasepool {
        NSRunningApplication *app = [[NSWorkspace sharedWorkspace] frontmostApplication];
        NSString *name = [app localizedName];
        if (name == nil) return strdup("");
        return strdup([name UTF8String]);
    }
}

int is_self_foreground(void) {
    @autoreleasepool {
        NSRunningApplication *front = [[NSWorkspace sharedWorkspace] frontmostApplication];
        NSRunningApplication *me = [NSRunningApplication currentApplication];
        return [front isEqual:me] ? 1 : 0;
    }
}
