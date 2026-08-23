package main

import (
	"path/filepath"
	"testing"
)

func TestMainUsage(t *testing.T) {
	if err := run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"init", filepath.Join(t.TempDir(), "main.db")}); err != nil {
		t.Fatal(err)
	}
}
