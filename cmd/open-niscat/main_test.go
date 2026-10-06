package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"strings"
	"testing"

	"open-niscat/internal/catalog/catalogtest"
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

func TestPortInUseExplainsWhatToDo(t *testing.T) {
	busy, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	addr := busy.Addr().String()
	err = run(context.Background(), []string{"--data", catalogtest.NewDataDir(t), "--addr", addr, "--open-browser=false"})
	if err == nil || !strings.Contains(err.Error(), "already running") || !strings.Contains(err.Error(), "http://"+addr+"/") {
		t.Fatalf("run on a busy port = %v, want a hint that open-niscat may already be running at http://%s/", err, addr)
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
