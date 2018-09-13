package ansi

import (
	"image/gif"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/textmodes/parser/format/sauce"
)

func TestDecode(t *testing.T) {
	tests, err := filepath.Glob("testdata/*.ans")
	if err != nil {
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		t.Skip(err)
	}

	for _, test := range tests {
		t.Run(filepath.Base(test), func(t *testing.T) {
			/*
				want, err := ioutil.ReadFile(test + ".out")
				if err != nil {
					t.Fatal(err)
				}
			*/

			s, err := os.Open(test)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()

			r, err := sauce.Parse(s)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.Seek(0, io.SeekStart); err != nil {
				t.Fatal(err)
			}

			d := NewDecoder()
			d.AutoExpand = true
			if r.TypeInfo[0] > 0 {
				t.Logf("resize to (%d, %d)", r.TypeInfo[0], r.TypeInfo[1])
				d.Resize(uint(r.TypeInfo[0]), uint(r.TypeInfo[1]))
			}
			if err = d.Decode(s); err != nil {
				t.Fatal(err)
			}
			if err = s.Close(); err != nil {
				t.Fatal(err)
			}

			/*
				if got := d.String(); got != string(want) {
					ioutil.WriteFile(test+".failed", []byte(got), 0644)
					t.Fatalf("expected:\n%s\n, got:\n%s", hex.Dump(want), hex.Dump([]byte(got)))
					//t.Fatalf("expected:\n%s\n, got:\n%s", want, got)
				}
			*/

			/*
				b, err := data.Bytes("font/chargen/ibm_vga.bin")
				if err != nil {
					t.Skip(err)
				}
				d.Font = chargen.New(chargen.NewBytesMask(b, chargen.MaskOptions{
					Size: image.Pt(8, 16),
				}))
			*/
			if d.Font, err = r.Font(); err != nil {
				t.Fatal(err)
			}

			i, err := d.Image()
			if err != nil {
				t.Fatal(err)
			}

			if os.Getenv("TEST_WRITE_IMAGE") == "" {
				t.Logf("Not writing image, TEST_WRITE_IMAGE not set")
			} else {
				{
					o, err := os.Create(test + ".png")
					if err != nil {
						t.Fatal(err)
					}
					defer o.Close()
					if err = png.Encode(o, i); err != nil {
						t.Fatal(err)
					}
				}
				{
					g, err := d.AnimateDelay(time.Millisecond * 400) // ~
					if err != nil {
						t.Fatal(err)
					}
					o, err := os.Create(test + ".gif")
					if err != nil {
						t.Fatal(err)
					}
					defer o.Close()
					if err = gif.EncodeAll(o, g); err != nil {
						t.Fatal(err)
					}
				}
			}

			if os.Getenv("TEST_WRITE_SCROLLER") == "" {
				t.Logf("Not writing scroller, TEST_WRITE_SCROLLER not set")
			} else {
				g, err := d.Scroller()
				if err != nil {
					t.Fatal(err)
				}
				o, err := os.Create(test + "-scroller.gif")
				if err != nil {
					t.Fatal(err)
				}
				defer o.Close()
				if err = gif.EncodeAll(o, g); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestDelay(t *testing.T) {
	delay := time.Millisecond * 400
	if d := int(delay / (time.Second / 100)); d != 40 {
		t.Fatalf("expected %s in 100ths of a second to be 40, got %d", delay, d)
	}
}
