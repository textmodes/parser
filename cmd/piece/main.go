package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/textmodes/parser"
	"github.com/textmodes/parser/format/sauce"
	"github.com/textmodes/parser/text/ansi"
)

var program = filepath.Base(os.Args[0])

var (
	stdout = bufio.NewWriter(os.Stdout)
	quiet  bool
)

func usage() {
	fmt.Fprintf(os.Stderr, "%s [<options>] <input>\n", program)

	opts := func(s ...string) {
		m := make(map[string]bool)
		for _, v := range s {
			m[v] = true
		}

		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			if !m[f.Name] {
				return
			}

			s := fmt.Sprintf("  -%s", f.Name) // Two spaces before -; see next two comments.
			name, usage := flag.UnquoteUsage(f)
			if len(name) > 0 {
				s += " " + name
			}
			if len(s) <= 4 { // space, space, '-', 'x'.
				s += "\t"
			} else {
				s += "\n    \t"
			}
			s += strings.Replace(usage, "\n", "\n    \t", -1)
			if !isZeroValue(f, f.DefValue) {
				s += fmt.Sprintf(" (default %v)", f.DefValue)
			}
			fmt.Fprint(flag.CommandLine.Output(), s, "\n")
		})
	}

	fmt.Fprintln(os.Stderr, "\nOptions:")
	opts("o", "q")

	fmt.Fprintln(os.Stderr, "\nRender options:")
	opts("animate", "scroll")

	fmt.Fprintln(os.Stderr, "\nANSi specific options:")
	opts("blink", "font", "noblink")

	os.Exit(1)
}

// isZeroValue determines whether the string represents the zero
// value for a flag.
func isZeroValue(f *flag.Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(f.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Ptr {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	return value == z.Interface().(flag.Value).String()
}

func main() {
	kind := flag.String("type", "auto", `parser type ("list" for a list)`)
	output := flag.String("o", "", "output file name (default append extension to input file name)")
	flag.BoolVar(&quiet, "q", false, "be quiet")

	animate := flag.Duration("animate", 0, "create a animated GIF (default false)")
	scroll := flag.Duration("scroll", 0, "create a scrolling GIF (default false)")

	blink := flag.Bool("blink", true, "blink toggle")
	font := flag.String("font", "", `font name (default use SAUCE) ("list" for a list)`)
	ignoreTab := flag.Bool("notab", false, "replace tabs by spaces (default: off)")

	flag.Usage = usage
	flag.Parse()

	if *font == "list" {
		listfonts()
		os.Exit(0)
	}
	if *kind == "list" {
		listtypes()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	}

	var (
		name = args[0]
		f    io.ReadCloser
		err  error
	)
	if name == "-" || filepath.Clean(name) == "/dev/stdin" {
		f = os.Stdin
	} else if f, err = os.Open(name); err != nil {
		fatalf("error opening %s: %v", name, err)
		if *font == "" {
			fmt.Fprintf(os.Stderr, "%s: reading font from SAUCE\n", program)
			var r *sauce.Record
			if r, err = sauce.Parse(f); err != nil && err != sauce.ErrNoRecord {
				fatalf("error scanning for SAUCE record in %s: %v", name, err)
			}
			if _, err = f.(*os.File).Seek(0, io.SeekStart); err != nil {
				fatalf("error rewinding %s: %v", name, err)
			}
			if r.DataType == sauce.Character {
				if *font = r.Info; *font != "" {
					fmt.Fprintf(os.Stderr, "%s: using font %q\n", program, *font)
				}
			}
		}
	}
	defer f.Close()

	// Resolve font aliases
	if alias, ok := fontAlias[strings.ToLower(*font)]; ok {
		*font = alias
	}

	decoder, err := parseFor(*kind, name, *font)
	if err != nil {
		fatalf("unable to find parser for %s: %v", name, err)
	}

	if *ignoreTab {
		var b []byte
		if b, err = ioutil.ReadAll(f); err != nil {
			fatalf("error reading %s: %v", name, err)
		}
		b = bytes.Replace(b, []byte{ansi.TAB}, []byte{' '}, -1)
		f = nullCloser{bytes.NewReader(b)}
	}

	var parsed parser.Parser
	timer("decoding", func() {
		if parsed, err = decoder(f); err != nil {
			fatalf("error decoding %s: %v", name, err)
		}
	})

	if !quiet {
		if p, ok := parsed.(progreser); ok {
			p.Progress(progress)
		}
	}

	var (
		im image.Image
		g  *gif.GIF
	)
	switch {
	case *scroll != 0:
		if s, ok := parsed.(parser.ScrollerDelay); ok {
			timer("rendering", func() {
				if g, err = s.ScrollerDelay(*scroll); err != nil {
					fatalf("error: %v", err)
				}
			})
			writeGIF(g, name, *output)
		}
		if s, ok := parsed.(parser.Scroller); ok {
			timer("rendering", func() {
				if g, err = s.Scroller(); err != nil {
					fatalf("error: %v", err)
				}
			})
			writeGIF(g, name, *output)
		}
		fatalf("%T does not support rendering scrollers", parsed)

	case *animate != 0:
		if a, ok := parsed.(parser.AnimationDelay); ok {
			timer("rendering", func() {
				if g, err = a.AnimateDelay(*animate); err != nil {
					fatalf("error: %v", err)
				}
			})
			writeGIF(g, name, *output)
		}
		if a, ok := parsed.(parser.Animation); ok {
			timer("rendering", func() {
				if g, err = a.Animate(); err != nil {
					fatalf("error: %v", err)
				}
			})
			writeGIF(g, name, *output)
		}
		fatalf("%T does not support rendering animations", parsed)

	default:
		if i, ok := parsed.(imagerWithBlink); ok {
			timer("rendering", func() {
				if im, err = i.Image(*blink); err != nil {
					fatalf("error: %v", err)
				}
			})
			writePNG(im, name, *output)
		}
		if i, ok := parsed.(parser.Image); ok {
			timer("rendering", func() {
				if im, err = i.Image(); err != nil {
					fatalf("error: %v", err)
				}
			})
			writePNG(im, name, *output)
		}
		fatalf("%T does not support rendering images", parsed)
	}
}

type writeCounter struct {
	io.Writer
	Count int64
}

func (w *writeCounter) Write(b []byte) (n int, err error) {
	n, err = w.Writer.Write(b)
	atomic.AddInt64(&w.Count, int64(n))
	return
}

func writeGIF(im *gif.GIF, name, output string) {
	if output == "" {
		output = name + ".gif"
		fmt.Fprintf(os.Stderr, "%s: no output given, using %s\n", program, output)
	}

	f, err := os.Create(output)
	if err != nil {
		fatalf("error creating %s: %v", output, err)
	}
	defer f.Close()

	c := &writeCounter{Writer: f}
	timer("encoding", func() {
		if err = gif.EncodeAll(c, im); err != nil {
			fatalf("error generating %s: %v", output, err)
		}
	})

	if err = f.Close(); err != nil {
		fatalf("error closing %s: %v", output, err)
	}

	infof("%s: wrote %d bytes\n", output, c.Count)
	os.Exit(0)
}

func writePNG(im image.Image, name, output string) {
	if output == "" {
		output = name + ".png"
		fmt.Fprintf(os.Stderr, "%s: no output given, using %s\n", program, output)
	}

	f, err := os.Create(output)
	if err != nil {
		fatalf("error creating %s: %v", output, err)
	}
	defer f.Close()

	c := &writeCounter{Writer: f}
	timer("encoding", func() {
		if err = png.Encode(c, im); err != nil {
			fatalf("error generating %s: %v", output, err)
		}
	})

	infof("%s: wrote %d bytes\n", output, c.Count)
	os.Exit(0)
}

type nullCloser struct {
	io.Reader
}

func (nullCloser) Close() error { return nil }

func timer(what string, fn func()) {
	start := time.Now()
	fn()
	ended := time.Since(start)
	infof("%s took %s", what, ended)
}

type progreser interface {
	Progress(func(float64))
}

func progress(perc float64) {
	var (
		p = int(perc * 40)
		o = make([]byte, 40)
		c = 31
	)
	if perc >= 0.75 {
		c = 32
	} else if perc >= 0.33 {
		c = 33
	}
	for i := 0; i < p; i++ {
		o[i] = '#'
	}
	for i := p; i < 40; i++ {
		o[i] = '-'
	}
	fmt.Fprintf(stdout, "\r%s: rendering [\x1b[1;%dm%s\x1b[0m] %-6.02f%%", program, c, string(o), perc*100)
	if perc == 1 {
		stdout.WriteByte('\n')
	}
	stdout.Flush()
}

func infof(format string, v ...interface{}) {
	if quiet {
		return
	}
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(os.Stderr, program+": "+format, v...)
}

func fatalf(format string, v ...interface{}) {
	format = strings.TrimRight(format, "\r\n") + "\n"
	fmt.Fprintf(os.Stderr, program+": "+format, v...)
	os.Exit(1)
}
