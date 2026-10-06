package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"testing"
)

func TestBrowserURL(t *testing.T) {
	cases := map[string]string{
		"127.0.0.1:8080": "http://127.0.0.1:8080/",
		"0.0.0.0:9000":   "http://localhost:9000/",
		"[::]:9000":      "http://localhost:9000/",
	}
	for in, want := range cases {
		addr, err := net.ResolveTCPAddr("tcp", in)
		if err != nil {
			t.Fatal(err)
		}
		if got := browserURL(addr); got != want {
			t.Errorf("browserURL(%s) = %q, want %q", in, got, want)
		}
	}
}

func TestHelpIsNotAnError(t *testing.T) {
	err := run(context.Background(), []string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("run(-h) = %v, want flag.ErrHelp", err)
	}
	if code := exitCode(err); code != 0 {
		t.Fatalf("exit code for -h = %d, want 0", code)
	}
	if code := exitCode(errors.New("boom")); code != 1 {
		t.Fatalf("exit code for an error = %d, want 1", code)
	}
}
