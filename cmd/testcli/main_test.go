package main

import "testing"

func TestInsertRuneDoesNotDuplicateTail(t *testing.T) {
	if got, want := insertRune("abc", 0, '你'), "你abc"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInsertRuneClampsIndex(t *testing.T) {
	if got, want := insertRune("abc", -1, '你'), "你abc"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := insertRune("abc", 9, '你'), "abc你"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
