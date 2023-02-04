//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package colorjson

import (
	"io"
)

func isatty(_ io.Writer) bool {
	return false
}
