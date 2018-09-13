package binarytext

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/textmodes/parser/format/sauce"
)

func TestDecode(t *testing.T) {
	names, err := filepath.Glob("testdata/*.bin")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	for _, name := range names {
		t.Run(filepath.Base(name), func(t *testing.T) {
			testDecode(t, name)
		})
	}
}

func testDecode(t *testing.T, name string) {
	t.Helper()

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("%s open error: %v", name, err)
	}
	defer f.Close()

	if _, err = Decode(f); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeError(t *testing.T) {
	t.Run("No SAUCE", func(t *testing.T) {
		if _, err := Decode(bytes.NewReader(make([]byte, 128))); err == nil {
			t.Fatal("expected error")
		} else if err != sauce.ErrNoRecord {
			t.Fatalf("expected error %q, got %q", sauce.ErrNoRecord, err)
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("Wrong SAUCE DataType", func(t *testing.T) {
		b := new(bytes.Buffer)
		r := &sauce.Record{DataType: sauce.XBIN}
		b.WriteByte(0x1a)
		r.WriteTo(b)

		if _, err := Decode(bytes.NewReader(b.Bytes())); err == nil {
			t.Fatal("expected error")
		} else if err != ErrSAUCEDataType {
			t.Fatalf("expected error %q, got %q", ErrSAUCEDataType, err)
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("Wrong SAUCE Font", func(t *testing.T) {
		b := new(bytes.Buffer)
		r := &sauce.Record{DataType: sauce.BinaryText, Info: "Fairy godmother's font"}
		r.WriteTo(b)

		if _, err := Decode(bytes.NewReader(b.Bytes())); err == nil {
			t.Fatal("expected error")
		} else {
			t.Logf("expected error: %v", err)
		}
	})

	t.Run("Decode", func(t *testing.T) {
		b := new(bytes.Buffer)
		r := &sauce.Record{DataType: sauce.BinaryText, Info: "IBM VGA"}
		r.WriteTo(b)

		if _, err := Decode(bytes.NewReader(b.Bytes())); err == nil {
			t.Fatal("expected error")
		} else {
			t.Logf("expected error: %v", err)
		}
	})
}
