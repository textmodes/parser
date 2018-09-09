package teletext

import (
	"testing"
	"time"
)

func TestLineHeader(t *testing.T) {
	var tests = []struct {
		Test Line
		Want string
		Page *Page
	}{
		{BlankHeader, "P100    GOLANG 100 Mon 02 Jan \x03 15:04.05", nil},
		{Line("XXXXXXXXTEEFAX %%# DAY %d MTH   %H:%M.%S"), "P100    TEEFAX 100 Mon 02 Jan   15:04.05", nil},
		{Line("XXXXXXXXShort Line %%#"), "P100    Short Line 100                  ", nil},
		{Line("XXXXXXXXNo page number %%#"), "P100    No page number 100              ", new(Page)},
	}

	now, _ := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2006")
	for _, test := range tests {
		t.Run(string(test.Test), func(t *testing.T) {
			if got := test.Test.HeaderTime(test.Page, now); got != test.Want {
				t.Fatalf("expected %q to return %q, got %q", test.Test, test.Want, got)
			}
		})
	}
}
