package main

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
)

// browserURL returns the URL to open for a listener; unspecified hosts become localhost.
func browserURL(addr net.Addr) string {
	host, port := "localhost", ""
	if tcp, ok := addr.(*net.TCPAddr); ok {
		port = strconv.Itoa(tcp.Port)
		if !tcp.IP.IsUnspecified() {
			host = tcp.IP.String()
		}
	}
	return "http://" + net.JoinHostPort(host, port) + "/"
}

// openBrowser asks the OS to open url in the default browser.
func openBrowser(ctx context.Context, url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", url) //nolint:gosec // url is built by the server from its own listen address
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url) //nolint:gosec // url is built by the server from its own listen address
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", url) //nolint:gosec // url is built by the server from its own listen address
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
