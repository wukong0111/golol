package main

import "testing"

func TestListenAddr(t *testing.T) {
	t.Setenv("ADDR", "")
	t.Setenv("PORT", "")
	if got := listenAddr(); got != ":8080" {
		t.Fatalf("default: %s", got)
	}

	t.Setenv("PORT", "3000")
	if got := listenAddr(); got != ":3000" {
		t.Fatalf("PORT: %s", got)
	}

	t.Setenv("ADDR", "8081")
	if got := listenAddr(); got != ":8081" {
		t.Fatalf("ADDR without colon: %s", got)
	}

	t.Setenv("ADDR", ":9090")
	t.Setenv("PORT", "3000")
	if got := listenAddr(); got != ":9090" {
		t.Fatalf("ADDR wins over PORT: %s", got)
	}
}

func TestShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	t.Setenv("RAILWAY_DEPLOYMENT_DRAINING_SECONDS", "")
	if got := shutdownTimeout(); got != defaultShutdownTimeout {
		t.Fatalf("default: %s", got)
	}

	t.Setenv("RAILWAY_DEPLOYMENT_DRAINING_SECONDS", "15")
	if got := shutdownTimeout(); got.Seconds() != 14 {
		t.Fatalf("draining margin: %s", got)
	}

	t.Setenv("SHUTDOWN_TIMEOUT", "8s")
	if got := shutdownTimeout(); got.Seconds() != 8 {
		t.Fatalf("explicit timeout wins: %s", got)
	}
}
