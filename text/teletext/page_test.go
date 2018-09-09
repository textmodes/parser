package teletext

import (
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
	"testing"
)

func testPage(t *testing.T, page *Page, name string) {
	t.Helper()
	if page == nil {
		t.Fatal("page is nil")
	}

	t.Run(fmt.Sprintf("page %05x", page.Number), func(t *testing.T) {
		im, err := page.Image()
		if err != nil {
			t.Fatal(err)
		}

		t.Run("png", func(t *testing.T) {
			if err := png.Encode(ioutil.Discard, im); err != nil {
				t.Fatal(err)
			}
			if os.Getenv("TEST_PAGE_WRITE_IMAGE") == "" {
				t.Logf("skipping write: TEST_PAGE_WRITE_IMAGE not set")
				return
			}

			o := fmt.Sprintf("%s-%x.png", name, page.Number)
			f, err := os.Create(o)
			if err != nil {
				t.Skip(err)
			}
			defer f.Close()
			if err = png.Encode(f, im); err != nil {
				t.Fatal(err)
			}
		})
	})
}
