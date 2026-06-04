package ci

import (
	"testing"
)

func TestMockFileReader(t *testing.T) {
	r := MockFileReader{Files: map[string]string{
		"test.txt": "hello",
	}}
	data, err := r.ReadFile("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected hello, got %s", data)
	}
	_, err = r.ReadFile("missing.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestOSFileReader(t *testing.T) {
	r := OSFileReader{}
	data, err := r.ReadFile("file_reader_test.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty data from reading own file")
	}
}

