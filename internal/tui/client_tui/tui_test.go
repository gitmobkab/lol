package client_tui

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{10 * 1024 * 1024, "10.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tc := range cases {
		got := humanBytes(tc.in)
		if got != tc.want {
			t.Errorf("humanBytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestShortID(t *testing.T) {
	id := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	got := shortID(id)
	if len(got) != 6 {
		t.Errorf("shortID length = %d, want 6", len(got))
	}
	if got != "123456" {
		t.Errorf("shortID = %q, want %q", got, "123456")
	}
}

func TestShortIDNoDashes(t *testing.T) {
	got := shortID(uuid.New())
	if strings.ContainsRune(got, '-') {
		t.Errorf("shortID %q contains dash", got)
	}
}

func TestShortIDLength(t *testing.T) {
	for range 20 {
		got := shortID(uuid.New())
		if len(got) != 6 {
			t.Errorf("shortID(%v) length = %d, want 6", got, len(got))
		}
	}
}

func TestMemberTag(t *testing.T) {
	id := uuid.MustParse("12345678-1234-1234-1234-123456789abc")
	got := memberTag("alice", id)
	want := "alice#123456"
	if got != want {
		t.Errorf("memberTag = %q, want %q", got, want)
	}
}

func TestMemberTagFormat(t *testing.T) {
	id := uuid.New()
	name := "bob"
	got := memberTag(name, id)
	if !strings.HasPrefix(got, name+"#") {
		t.Errorf("memberTag %q does not start with %q", got, name+"#")
	}
	suffix := strings.TrimPrefix(got, name+"#")
	if len(suffix) != 6 {
		t.Errorf("shortid suffix %q has length %d, want 6", suffix, len(suffix))
	}
}
