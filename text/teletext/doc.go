// Package teletext implements parsers for TeleText Level 1 formats.
//
//go:generate esc -o data.go -pkg teletext -prefix data/ data
package teletext

import (
	"fmt"
	"os"
	"strings"
)

var (
	// Debug flag
	Debug bool

	// Trace flag
	Trace bool
)

func debug(v ...interface{}) {
	if !Debug {
		return
	}
	fmt.Fprintln(os.Stderr, append([]interface{}{"teletext: "}, v...)...)
}

func debugf(format string, v ...interface{}) {
	if !Debug {
		return
	}
	fmt.Fprintf(os.Stderr, "teletext: "+strings.TrimRight(format, "\r\n")+"\n", v...)
}

func trace(v ...interface{}) {
	if !(Trace || Debug) {
		return
	}
	fmt.Fprintln(os.Stderr, append([]interface{}{"teletext: "}, v...)...)
}

func tracef(format string, v ...interface{}) {
	if !(Trace || Debug) {
		return
	}
	fmt.Fprintf(os.Stderr, "teletext: "+strings.TrimRight(format, "\r\n")+"\n", v...)
}
