// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cfg := Config{RequestsPerSecond: 10, Burst: 20}
	limiter := New("test", cfg)

	if limiter.Name() != "test" {
		t.Errorf("expected name 'test', got %q", limiter.Name())
	}

	gotCfg := limiter.Config()
	if gotCfg.RequestsPerSecond != cfg.RequestsPerSecond {
		t.Errorf("expected RPS %v, got %v", cfg.RequestsPerSecond, gotCfg.RequestsPerSecond)
	}
	if gotCfg.Burst != cfg.Burst {
		t.Errorf("expected Burst %v, got %v", cfg.Burst, gotCfg.Burst)
	}
}

func TestNewFromDefaults(t *testing.T) {
	tests := []struct {
		provider    string
		expectedRPS float64
		expectedB   int
	}{
		{"aws", 10, 20},
		{"github", 5, 10},
		{"azure", 10, 20},
		{"cloudflare", 4, 8},
		{"ms365", 10, 20},
		{"hetzner", 10, 20},
		{"factorial", 5, 10},
		{"unknown", 5, 10}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			limiter := NewFromDefaults(tt.provider)

			cfg := limiter.Config()
			if cfg.RequestsPerSecond != tt.expectedRPS {
				t.Errorf("expected RPS %v, got %v", tt.expectedRPS, cfg.RequestsPerSecond)
			}
			if cfg.Burst != tt.expectedB {
				t.Errorf("expected Burst %v, got %v", tt.expectedB, cfg.Burst)
			}
		})
	}
}

func TestLimiterWait(t *testing.T) {
	// Create a limiter with 100 req/s to make test fast
	limiter := New("test", Config{RequestsPerSecond: 100, Burst: 1})

	ctx := context.Background()

	// First request should succeed immediately (burst)
	start := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	firstDuration := time.Since(start)

	// Should be very fast (burst available)
	if firstDuration > 50*time.Millisecond {
		t.Errorf("first request took too long: %v", firstDuration)
	}
}

func TestLimiterWaitCanceled(t *testing.T) {
	// Create a slow limiter
	limiter := New("test", Config{RequestsPerSecond: 1, Burst: 1})

	// Use up the burst
	ctx := context.Background()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Wait should return error due to canceled context
	err := limiter.Wait(ctx)
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestLimiterAllow(t *testing.T) {
	limiter := New("test", Config{RequestsPerSecond: 1, Burst: 2})

	// First two should be allowed (burst)
	if !limiter.Allow() {
		t.Error("expected first Allow to return true")
	}
	if !limiter.Allow() {
		t.Error("expected second Allow to return true")
	}

	// Third should be denied (burst exhausted)
	if limiter.Allow() {
		t.Error("expected third Allow to return false")
	}
}

func TestLimiterAllowN(t *testing.T) {
	limiter := New("test", Config{RequestsPerSecond: 1, Burst: 5})

	// Should allow 3 at once
	if !limiter.AllowN(3) {
		t.Error("expected AllowN(3) to return true")
	}

	// Should allow 2 more (remaining burst)
	if !limiter.AllowN(2) {
		t.Error("expected AllowN(2) to return true")
	}

	// Should deny any more
	if limiter.AllowN(1) {
		t.Error("expected AllowN(1) to return false")
	}
}

func TestLimiterSetLimit(t *testing.T) {
	limiter := New("test", Config{RequestsPerSecond: 10, Burst: 20})

	limiter.SetLimit(50)

	cfg := limiter.Config()
	if cfg.RequestsPerSecond != 50 {
		t.Errorf("expected RPS 50, got %v", cfg.RequestsPerSecond)
	}
}

func TestLimiterSetBurst(t *testing.T) {
	limiter := New("test", Config{RequestsPerSecond: 10, Burst: 20})

	limiter.SetBurst(100)

	cfg := limiter.Config()
	if cfg.Burst != 100 {
		t.Errorf("expected Burst 100, got %v", cfg.Burst)
	}
}

func TestNopLimiter(t *testing.T) {
	limiter := NopLimiter()

	// Should always allow
	for i := 0; i < 1000; i++ {
		if !limiter.Allow() {
			t.Error("expected NopLimiter to always allow")
		}
	}

	// Wait should return immediately
	ctx := context.Background()
	start := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) > 10*time.Millisecond {
		t.Error("NopLimiter Wait took too long")
	}
}
