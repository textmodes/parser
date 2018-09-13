/*
Package ansi implements a text terminal that can render characters using a
chargen font, such as found on IBM PC Code Page 437 but also on Commodore
Amiga's DOS.
*/
package ansi

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
	fmt.Fprintln(os.Stderr, append([]interface{}{"ansi: "}, v...)...)
}

func debugf(format string, v ...interface{}) {
	if !Debug {
		return
	}
	fmt.Fprintf(os.Stderr, "ansi: "+strings.TrimRight(format, "\r\n")+"\n", v...)
}

func trace(v ...interface{}) {
	if !(Trace || Debug) {
		return
	}
	fmt.Fprintln(os.Stderr, append([]interface{}{"ansi: "}, v...)...)
}

func tracef(format string, v ...interface{}) {
	if !(Trace || Debug) {
		return
	}
	fmt.Fprintf(os.Stderr, "ansi: "+strings.TrimRight(format, "\r\n")+"\n", v...)
}
