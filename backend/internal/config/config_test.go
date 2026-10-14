package config

import "testing"

func TestLoadDefaultsAndOverrides(t *testing.T) {
	t.Setenv("PORT", "")
	if got := Load().Port; got != "8080" {
		t.Fatalf("default port %q", got)
	}
	t.Setenv("PORT", "9090")
	if got := Load().Port; got != "9090" {
		t.Fatalf("override port %q", got)
	}
}
