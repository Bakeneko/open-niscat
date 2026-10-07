// Command open-niscat serves the NISCAT parts catalogue in a web browser.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	"open-niscat/internal/api"
	"open-niscat/internal/catalog"
	"open-niscat/internal/config"
	"open-niscat/web"
)

// version is set at build time (-ldflags "-X main.version=v1.2.3", see the Makefile).
var version string

// appVersion is the injected version, else the git revision Go records in the binary, else "dev".
func appVersion() string {
	if version != "" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	var rev, dirty string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				dirty = "-dirty"
			}
		}
	}
	if rev == "" {
		return "dev"
	}
	return rev[:min(len(rev), 12)] + dirty
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args[1:])
	stop()
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, "open-niscat:", err)
		pauseIfOwnConsole()
	}
	os.Exit(exitCode(err))
}

// exitCode maps run's result to the process exit code; -h (usage already printed) is a success.
func exitCode(err error) int {
	if err == nil || errors.Is(err, flag.ErrHelp) {
		return 0
	}
	return 1
}

func run(ctx context.Context, args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate binary: %w", err)
	}
	cfg, err := config.Load(args, filepath.Dir(exe))
	if errors.Is(err, config.ErrVersion) {
		fmt.Println("open-niscat", appVersion())
		return nil
	}
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	store, err := catalog.Open(ctx, cfg.Data)
	if err != nil {
		return fmt.Errorf("data directory %s: %w", cfg.Data, err)
	}
	defer store.Close()
	lang, err := catalog.ParseLang(cfg.DefaultLang)
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.Addr)
	if err != nil {
		hint := "is the address valid?"
		if addr, resolveErr := net.ResolveTCPAddr("tcp", cfg.Addr); resolveErr == nil {
			hint = "if open-niscat is already running, open " + browserURL(addr) + " in your browser"
		}
		return fmt.Errorf("cannot listen on %s (%s): %w", cfg.Addr, hint, err)
	}
	srv := &http.Server{Handler: api.New(store, web.Dist(), lang, appVersion()), ReadHeaderTimeout: 10 * time.Second}
	url := browserURL(listener.Addr())
	slog.Info("open-niscat ready", "version", appVersion(), "url", url, "data", cfg.Data, "edition", store.Manifest().Edition)

	served := make(chan error, 1)
	go func() { served <- srv.Serve(listener) }()
	if cfg.OpenBrowser {
		if err := openBrowser(ctx, url); err != nil {
			slog.Warn("could not open the browser", "err", err)
		}
	}

	select {
	case err := <-served:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	}
}
