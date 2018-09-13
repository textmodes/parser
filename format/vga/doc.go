// Package vga implements virtual Video Graphics Array as found on Personal
// Computers widespread in 1980s–1990s.
package vga

import (
	"fmt"
	"os"
	"strings"
)

var (
	// Trace flag
	Trace bool
)

func trace(v ...interface{}) {
	if !Trace {
		return
	}
	fmt.Fprintln(os.Stderr, append([]interface{}{"vga: "}, v...)...)
}

func tracef(format string, v ...interface{}) {
	if !Trace {
		return
	}
	fmt.Fprintf(os.Stderr, "vga: "+strings.TrimRight(format, "\r\n")+"\n", v...)
}
