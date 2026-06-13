package cmd

import "testing"

func TestNormalizeVersion(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"1.2.3", "v1.2.3"},
		{"v1.2.3", "v1.2.3"},
		{"0.1.0", "v0.1.0"},
		{"v0.1.0", "v0.1.0"},
		{"0.1.4", "v0.1.4"},
		{"", "v"},
	}
	for _, tc := range cases {
		got := normalizeVersion(tc.in)
		if got != tc.want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
