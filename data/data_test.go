package data

import (
	"path/filepath"
	"testing"
)

func Test(t *testing.T) {
	seen := make(map[string]bool)
	for name := range files {
		if dir := filepath.Dir(name); !seen[dir] {
			seen[dir] = true
			t.Run(dir, func(t *testing.T) {
				if info, err := Stat(dir); err != nil {
					t.Fatalf("Stat(%q) error: %v", dir, err)
				} else if !info.IsDir() {
					t.Fatalf("Stat(%q) error: not a dir", dir)
				}
				if _, err := Open(dir); err == nil {
					t.Fatalf("Open(%q) should fail", dir)
				} else {
					t.Logf("Open(%q) error as expected: %v", dir, err)
				}
			})
		}
		t.Run(name, func(t *testing.T) {
			if info, err := Stat(name); err != nil {
				t.Fatalf("Stat(%q) error: %v", name, err)
			} else if info.IsDir() {
				t.Fatalf("Stat(%q) error: is a dir", name)
			}
			if _, err := Open(name); err != nil {
				t.Fatalf("Open(%q) error: %v", name, err)
			}
		})
	}
}
