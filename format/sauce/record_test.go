package sauce

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var tests = []struct {
	Name string
	Record
}{
	{"testdata/test.ans", Record{
		Title:    "Hello",
		Author:   "maze",
		Info:     "IBM VGA",
		DataType: Character,
		FileType: ANSi,
	}},
	{"testdata/test.bin", Record{
		Title:    "Hello",
		Author:   "maze",
		Info:     "IBM VGA",
		DataType: BinaryText,
		FileType: 0x28,
	}},
	{"testdata/test.xb", Record{
		Title:    "Hello",
		Author:   "maze",
		DataType: XBIN,
	}},
}

func TestParse(t *testing.T) {
	for _, test := range tests {
		t.Run(filepath.Base(test.Name), func(t *testing.T) {
			f, err := os.Open(test.Name)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			r, err := Parse(f)
			if err != nil {
				t.Fatal(err)
			}
			if r == nil {
				t.Fatal("Parse returned nil Record")
			}
			assertRecord(t, test.Record, *r)
		})
	}
}

func TestParseReader(t *testing.T) {
	for _, test := range tests {
		t.Run(filepath.Base(test.Name), func(t *testing.T) {
			b, err := ioutil.ReadFile(test.Name)
			if err != nil {
				t.Fatal(err)
			}

			r, err := Parse(bytes.NewBuffer(b))
			if err != nil {
				t.Fatal(err)
			}
			if r == nil {
				t.Fatal("Parse returned nil Record")
			}
			assertRecord(t, test.Record, *r)
		})
	}
}

func assertRecord(t *testing.T, want, got Record) {
	assertEqual(t, "Record.Title", want.Title, got.Title)
	assertEqual(t, "Record.Group", want.Group, got.Group)
	assertEqual(t, "Record.Author", want.Author, got.Author)
	assertEqual(t, "Record.Info", want.Info, got.Info)
	assertEqualByte(t, "Record.DataType", want.DataType, got.DataType)
	assertEqualByte(t, "Record.FileType", want.FileType, got.FileType)
}

func assertEqual(t *testing.T, name, want, got string) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %s to be %q, got %q", name, want, got)
	}
}

func assertEqualByte(t *testing.T, name string, want, got byte) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %s to be %#02x, got %#02x", name, want, got)
	}
}
