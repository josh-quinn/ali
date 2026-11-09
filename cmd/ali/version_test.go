package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	info := func(v string) *debug.BuildInfo { return &debug.BuildInfo{Main: debug.Module{Version: v}} }
	cases := []struct {
		name, stamped string
		info          *debug.BuildInfo
		want          string
	}{
		{"release build (ldflags)", "0.1.2", info("(devel)"), "0.1.2"},
		{"go install @v0.1.2", "dev", info("v0.1.2"), "0.1.2"},
		{"go install @latest pseudo-version", "dev", info("v0.1.3-0.20261109020000-abcdef123456"), "0.1.3-0.20261109020000-abcdef123456"},
		{"local go build", "dev", info("(devel)"), "dev"},
		{"no build info", "dev", nil, "dev"},
	}
	for _, c := range cases {
		if got := resolveVersion(c.stamped, c.info); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
