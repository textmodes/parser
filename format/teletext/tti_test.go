package teletext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeTTI(t *testing.T) {
	names, err := filepath.Glob("testdata/*.[tT][tT][iI]")
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

			pages, err := DecodeTTI(f)
			if err != nil {
				t.Fatal(err)
			}
			for _, page := range pages {
				testPage(t, page, name)
			}
		})
	}
}
