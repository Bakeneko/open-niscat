//go:build !windows

package main

// pauseIfOwnConsole is only needed on Windows, where a double-clicked program's window closes on exit.
func pauseIfOwnConsole() {}
