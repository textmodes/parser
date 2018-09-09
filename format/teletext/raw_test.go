package teletext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRaw(t *testing.T) {
	names, err := filepath.Glob("testdata/*.[rR][aA][wW]")
	if err != nil {
		t.Skip(err)
	}

	for _, name := range names {
		f, err := os.Open(name)
		if err != nil {
			if os.IsNotExist(err) {
				t.Skip(err)
			}
			t.Fatal(err)
		}
		defer f.Close()

		pages, err := DecodeRaw(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(pages) == 0 {
			t.Fatal("no pages decoded")
		}
		for _, page := range pages {
			testPage(t, page, name)
		}
	}
}
