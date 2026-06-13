package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreFileAndBufferedFileNames(t *testing.T) {
	c := &Client{}
	c.storeFile("b.txt", []byte("b"))
	c.storeFile("a.txt", []byte("a"))
	c.storeFile("c.txt", []byte("c"))

	got := c.BufferedFileNames()
	want := []string{"a.txt", "b.txt", "c.txt"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, name := range want {
		if got[i] != name {
			t.Errorf("[%d] got %q, want %q", i, got[i], name)
		}
	}
}

func TestStoreFileOverwrites(t *testing.T) {
	c := &Client{}
	c.storeFile("a.txt", []byte("first"))
	c.storeFile("a.txt", []byte("second"))

	names := c.BufferedFileNames()
	if len(names) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(names), names)
	}
}

func TestBufferedFileNamesEmptyClient(t *testing.T) {
	c := &Client{}
	got := c.BufferedFileNames()
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestSaveFile(t *testing.T) {
	c := &Client{}
	data := []byte("file contents")
	c.storeFile("hello.txt", data)

	dir := t.TempDir()
	if err := c.SaveFile("hello.txt", dir); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestSaveFileUnknown(t *testing.T) {
	c := &Client{}
	err := c.SaveFile("nonexistent.txt", t.TempDir())
	if err == nil {
		t.Error("expected error for unknown file, got nil")
	}
}

func TestSaveFileCreatesDirectory(t *testing.T) {
	c := &Client{}
	c.storeFile("x.txt", []byte("x"))

	dir := filepath.Join(t.TempDir(), "nested", "dir")
	if err := c.SaveFile("x.txt", dir); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "x.txt")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestSaveFilePathTraversal(t *testing.T) {
	c := &Client{}
	c.storeFile("../../evil.txt", []byte("evil"))

	dir := t.TempDir()
	if err := c.SaveFile("../../evil.txt", dir); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}
	// filepath.Base strips the traversal; file must land inside dir
	if _, err := os.Stat(filepath.Join(dir, "evil.txt")); err != nil {
		t.Errorf("expected file at safe path: %v", err)
	}
}

func TestSaveFilePreservesContents(t *testing.T) {
	c := &Client{}
	data := []byte{0x00, 0xFF, 0xAB, 0xCD, 0x12, 0x34}
	c.storeFile("binary.bin", data)

	dir := t.TempDir()
	if err := c.SaveFile("binary.bin", dir); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "binary.bin"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("contents mismatch")
	}
}
