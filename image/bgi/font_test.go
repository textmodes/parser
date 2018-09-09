package bgi

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestFont(t *testing.T) {
	r, err := os.Open("SANS.CHR")
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()

	sans, err := NewFont(r)
	if err != nil {
		t.Fatal(err)
	}

	r, err = os.Open("SCRI.CHR")
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()

	scri, err := NewFont(r)
	if err != nil {
		t.Fatal(err)
	}

	r, err = os.Open("GOTH.CHR")
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()

	goth, err := NewFont(r)
	if err != nil {
		t.Fatal(err)
	}

	r, err = os.Open("TRIP.CHR")
	if err != nil {
		t.Skip(err)
	}
	defer r.Close()

	trip, err := NewFont(r)
	if err != nil {
		t.Fatal(err)
	}

	o := image.NewRGBA(image.Rect(0, 0, 640, 640))
	s := 4
	sans.Draw(o, 0, 0, s, "the quick brown", color.Black)
	sans.Draw(o, 0, 80, s, "fox jumps over", color.Black)
	sans.Draw(o, 0, 160, s, "the lazy dog!", color.Black)
	goth.Draw(o, 320, 0, s, "the quick brown", color.Black)
	goth.Draw(o, 320, 80, s, "fox jumps over", color.Black)
	goth.Draw(o, 320, 160, s, "the lazy dog!", color.Black)
	scri.Draw(o, 0, 400, s, "the quick brown", color.Black)
	scri.Draw(o, 0, 480, s, "fox jumps over", color.Black)
	scri.Draw(o, 0, 560, s, "the lazy dog!", color.Black)
	trip.Draw(o, 320, 400, s, "the quick brown", color.Black)
	trip.Draw(o, 320, 480, s, "fox jumps over", color.Black)
	trip.Draw(o, 320, 560, s, "the lazy dog!", color.Black)
	sans.Draw(o, 0, 240, 1, "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz", color.Black)
	goth.Draw(o, 0, 280, 1, "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz", color.Black)
	scri.Draw(o, 0, 320, 1, "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz", color.Black)
	trip.Draw(o, 0, 360, 1, "AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz", color.Black)

	if n := os.Getenv("TEST_IMAGE"); n != "" {
		w, err := os.OpenFile(n, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Close()
		if err = png.Encode(w, o); err != nil {
			t.Fatal(err)
		}
	}

	//t.Logf("%#+v", f)
}
