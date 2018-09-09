package ansi

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
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

			want, err := ioutil.ReadFile(test + ".out")
			if err != nil {
				t.Fatal(err)
			}

			f, err := os.Open(test)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			d := NewDecoder()
			if err = d.Decode(f); err != nil {
				t.Fatal(err)
			}

			if got := d.String(); got != string(want) {
				ioutil.WriteFile(test+".failed", []byte(got), 0644)
				t.Fatalf("expected:\n%s\n, got:\n%s", hex.Dump(want), hex.Dump([]byte(got)))
				//t.Fatalf("expected:\n%s\n, got:\n%s", want, got)
			}
		})
	}
}
