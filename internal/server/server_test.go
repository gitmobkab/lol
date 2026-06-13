package server

import (
	"log/slog"
	"testing"
)

func TestGetConnLimiterSameIP(t *testing.T) {
	s := New(slog.Default())
	l1 := s.getConnLimiter("192.168.1.1")
	l2 := s.getConnLimiter("192.168.1.1")
	if l1 != l2 {
		t.Error("expected same limiter for same IP, got different")
	}
}

func TestGetConnLimiterDifferentIPs(t *testing.T) {
	s := New(slog.Default())
	l1 := s.getConnLimiter("192.168.1.1")
	l2 := s.getConnLimiter("192.168.1.2")
	if l1 == l2 {
		t.Error("expected different limiters for different IPs, got same")
	}
}

func TestGetConnLimiterConcurrent(t *testing.T) {
	s := New(slog.Default())
	done := make(chan *struct{}, 10)
	for range 10 {
		go func() {
			s.getConnLimiter("10.0.0.1")
			done <- nil
		}()
	}
	for range 10 {
		<-done
	}
	// All goroutines stored/loaded the same limiter — just verifying no race.
}

func TestRateLimitConstants(t *testing.T) {
	if msgRate <= 0 {
		t.Error("msgRate must be positive")
	}
	if msgBurst <= 0 {
		t.Error("msgBurst must be positive")
	}
	if connRate <= 0 {
		t.Error("connRate must be positive")
	}
	if connBurst <= 0 {
		t.Error("connBurst must be positive")
	}
	if maxFileSize <= 0 {
		t.Error("maxFileSize must be positive")
	}
	if maxFilePayloadSize <= maxFileSize {
		t.Error("maxFilePayloadSize must exceed maxFileSize (base64 overhead)")
	}
}
