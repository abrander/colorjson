//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package colorjson

import (
	"io"
	"syscall"
	"unsafe"
)

// isatty returns true if w behaves like a terminal.
func isatty(w io.Writer) bool {
	f, ok := w.(interface {
		Fd() uintptr
	})
	if !ok {
		return false
	}

	var value int
	const TIOCGETD = 0x5424

	_, _, result := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), TIOCGETD, uintptr(unsafe.Pointer(&value)))

	return result == 0
}
