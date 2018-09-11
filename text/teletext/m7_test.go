package teletext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeM7(t *testing.T) {
	names, err := filepath.Glob("testdata/*.[mM]7")
	if err != nil {
		t.Skip(err)
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(name)
			if err != nil {
				t.Skip(err)
			}
			defer f.Close()

			page, err := DecodeM7(f)
			if err != nil {
				t.Fatal(err)
			}
			testPage(t, page, name)
		})
	}
}
