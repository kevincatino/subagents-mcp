package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirValidation(t *testing.T) {
	t.Run("rejects empty", func(t *testing.T) {
		if _, err := Dir(""); err == nil {
			t.Fatal("expected error for empty path")
		}
	})

	t.Run("rejects relative", func(t *testing.T) {
		if _, err := Dir("relative/path"); err == nil {
			t.Fatal("expected error for relative path")
		}
	})

	t.Run("rejects root", func(t *testing.T) {
		if _, err := Dir("/"); err == nil {
			t.Fatal("expected error for root")
		}
	})

	t.Run("rejects non-existent", func(t *testing.T) {
		if _, err := Dir("/tmp/definitely-does-not-exist-12345"); err == nil {
			t.Fatal("expected error for non-existent path")
		}
	})

	t.Run("accepts existing dir", func(t *testing.T) {
		dir := t.TempDir()
		abs, err := filepath.Abs(dir)
		if err != nil {
			t.Fatalf("abs: %v", err)
		}
		expectedResolved, err := filepath.EvalSymlinks(abs)
		if err != nil {
			t.Fatalf("eval symlinks: %v", err)
		}
		resolved, err := Dir(abs)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if resolved != expectedResolved {
			t.Fatalf("expected %s, got %s", expectedResolved, resolved)
		}
	})

	t.Run("rejects file path", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "file.txt")
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		if _, err := Dir(file); err == nil {
			t.Fatal("expected error for file path")
		}
	})
}
