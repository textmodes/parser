package teletext

import (
	"os"
	"testing"
)

func TestParseEP1(t *testing.T) {
	f, err := os.Open("testdata/illarterate-08_biffo.ep1")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer f.Close()

	pages, err := DecodeEP1(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) == 0 {
		t.Fatal("no pages decoded")
	}
	for _, page := range pages {
		testPage(t, page, "testdata/illarterate-08_biffo.ep1")
	}
}
