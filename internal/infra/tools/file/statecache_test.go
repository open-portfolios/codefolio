package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateCacheRecordAndCheck(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(fpath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(fpath)
	if err != nil {
		t.Fatal(err)
	}

	cache := NewStateCache()

	ok, msg := cache.Check(fpath)
	if ok {
		t.Errorf("expected Check to fail before Record, got ok=%v msg=%s", ok, msg)
	}

	cache.Record(fpath, content, info.ModTime().UnixMilli())

	ok, msg = cache.Check(fpath)
	if !ok {
		t.Errorf("expected Check to pass after Record, got ok=%v msg=%s", ok, msg)
	}
}

func TestStateCacheModifiedSinceRead(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(fpath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cache := NewStateCache()
	cache.Record(fpath, content, 0)

	ok, msg := cache.Check(fpath)
	if ok {
		t.Errorf("expected Check to fail because mtime is newer, got ok=%v msg=%s", ok, msg)
	}
}

func TestStateCacheUpdate(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(fpath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	info, _ := os.Stat(fpath)

	cache := NewStateCache()
	cache.Record(fpath, content, info.ModTime().UnixMilli())

	newContent := "updated"
	cache.Update(fpath, newContent)

	ok, msg := cache.Check(fpath)
	if !ok {
		t.Errorf("expected Check to pass after Update, got ok=%v msg=%s", ok, msg)
	}
}
