package cli

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseSchedule(t *testing.T) {
	t.Run("unix-seconds", func(t *testing.T) {
		got, err := parseSchedule("1767595800")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 1767595800 {
			t.Fatalf("expected 1767595800, got %d", got)
		}
	})

	t.Run("unix-millis", func(t *testing.T) {
		got, err := parseSchedule("1767595800123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 1767595800 {
			t.Fatalf("expected 1767595800, got %d", got)
		}
	})

	t.Run("rfc3339", func(t *testing.T) {
		ts := "2026-01-05T09:30:00Z"
		got, err := parseSchedule(ts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := int(time.Date(2026, 1, 5, 9, 30, 0, 0, time.UTC).Unix())
		if got != want {
			t.Fatalf("expected %d, got %d", want, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, err := parseSchedule("tomorrow"); err == nil {
			t.Fatalf("expected error for invalid schedule")
		}
	})

	t.Run("empty", func(t *testing.T) {
		if _, err := parseSchedule(" "); err == nil {
			t.Fatalf("expected error for empty schedule")
		}
	})
}

func TestDetectMedia(t *testing.T) {
	t.Run("jpeg-photo", func(t *testing.T) {
		path := writeTempFile(t, "tmgc-test-*.jpg", []byte{0xFF, 0xD8, 0xFF, 0xDB, 0x00, 0x00})
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		mimeType, isPhoto, err := detectMedia(f, path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(mimeType, "image/") {
			t.Fatalf("expected image mime type, got %q", mimeType)
		}
		if !isPhoto {
			t.Fatalf("expected photo detection for jpeg")
		}
	})

	t.Run("gif-not-photo", func(t *testing.T) {
		path := writeTempFile(t, "tmgc-test-*.gif", []byte("GIF89a"))
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		mimeType, isPhoto, err := detectMedia(f, path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mimeType != "image/gif" {
			t.Fatalf("expected image/gif, got %q", mimeType)
		}
		if isPhoto {
			t.Fatalf("expected gif to be treated as document")
		}
	})

	t.Run("text", func(t *testing.T) {
		path := writeTempFile(t, "tmgc-test-*.txt", []byte("hello"))
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		mimeType, isPhoto, err := detectMedia(f, path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mimeType != "text/plain" {
			t.Fatalf("expected text/plain, got %q", mimeType)
		}
		if isPhoto {
			t.Fatalf("expected text to be treated as document")
		}
	})
}

func writeTempFile(t *testing.T, pattern string, data []byte) string {
	t.Helper()

	f, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}
