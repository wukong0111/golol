package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAddr            = ":8080"
	defaultShutdownTimeout = 15 * time.Second
)

// listenAddr prefers ADDR, then Railway's PORT, then :8080.
func listenAddr() string {
	if addr := strings.TrimSpace(os.Getenv("ADDR")); addr != "" {
		if !strings.Contains(addr, ":") {
			return ":" + addr
		}
		return addr
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		if strings.HasPrefix(port, ":") {
			return port
		}
		return ":" + port
	}
	return defaultAddr
}

// shutdownTimeout is how long in-flight requests may finish after SIGTERM.
// Leaves one second of margin when Railway's draining window is set, so
// Shutdown returns before SIGKILL.
func shutdownTimeout() time.Duration {
	if raw := strings.TrimSpace(os.Getenv("SHUTDOWN_TIMEOUT")); raw != "" {
		d, err := time.ParseDuration(raw)
		if err == nil && d > 0 {
			return d
		}
	}
	if raw := strings.TrimSpace(os.Getenv("RAILWAY_DEPLOYMENT_DRAINING_SECONDS")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err == nil && n > 0 {
			d := time.Duration(n) * time.Second
			if d > time.Second {
				return d - time.Second
			}
			return d
		}
	}
	return defaultShutdownTimeout
}
