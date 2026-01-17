// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package concurrency

import (
	"runtime"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Auto {
		t.Error("expected Auto to be true by default")
	}
	if cfg.MaxWorkers != DefaultMaxWorkers {
		t.Errorf("expected MaxWorkers %d, got %d", DefaultMaxWorkers, cfg.MaxWorkers)
	}
}

func TestOptimalWorkers(t *testing.T) {
	cpus := runtime.GOMAXPROCS(0)

	tests := []struct {
		name          string
		cfg           Config
		resourceCount int
		want          int
	}{
		{
			name:          "auto with large workload",
			cfg:           Config{Auto: true, MaxWorkers: 20},
			resourceCount: 100,
			want:          min(cpus*DefaultMultiplier, 20),
		},
		{
			name:          "auto with small workload",
			cfg:           Config{Auto: true, MaxWorkers: 20},
			resourceCount: 3,
			want:          3,
		},
		{
			name:          "auto with zero resources",
			cfg:           Config{Auto: true, MaxWorkers: 20},
			resourceCount: 0,
			want:          min(cpus*DefaultMultiplier, 20),
		},
		{
			name:          "auto disabled",
			cfg:           Config{Auto: false, MaxWorkers: 10},
			resourceCount: 100,
			want:          10,
		},
		{
			name:          "max workers caps result",
			cfg:           Config{Auto: true, MaxWorkers: 2},
			resourceCount: 100,
			want:          2,
		},
		{
			name:          "zero max workers uses default",
			cfg:           Config{Auto: true, MaxWorkers: 0},
			resourceCount: 100,
			want:          min(cpus*DefaultMultiplier, DefaultMaxWorkers),
		},
		{
			name:          "single resource",
			cfg:           Config{Auto: true, MaxWorkers: 20},
			resourceCount: 1,
			want:          1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OptimalWorkers(tt.cfg, tt.resourceCount)
			if got != tt.want {
				t.Errorf("OptimalWorkers() = %d, want %d (cpus=%d)", got, tt.want, cpus)
			}
		})
	}
}

func TestDiscoveryWorkers(t *testing.T) {
	cfg := Config{Auto: true, MaxWorkers: 20}

	// Should delegate to OptimalWorkers
	got := DiscoveryWorkers(cfg, 5)
	want := OptimalWorkers(cfg, 5)

	if got != want {
		t.Errorf("DiscoveryWorkers() = %d, want %d", got, want)
	}
}

func TestFetchWorkers(t *testing.T) {
	cfg := Config{Auto: true, MaxWorkers: 20}

	// Should delegate to OptimalWorkers
	got := FetchWorkers(cfg, 10)
	want := OptimalWorkers(cfg, 10)

	if got != want {
		t.Errorf("FetchWorkers() = %d, want %d", got, want)
	}
}

func TestEvaluationWorkers(t *testing.T) {
	cpus := runtime.GOMAXPROCS(0)

	tests := []struct {
		name string
		cfg  Config
		want int
	}{
		{
			name: "auto mode uses GOMAXPROCS",
			cfg:  Config{Auto: true, MaxWorkers: 100},
			want: cpus,
		},
		{
			name: "auto mode respects max workers",
			cfg:  Config{Auto: true, MaxWorkers: 2},
			want: min(cpus, 2),
		},
		{
			name: "auto disabled uses max workers",
			cfg:  Config{Auto: false, MaxWorkers: 10},
			want: 10,
		},
		{
			name: "zero max workers uses default",
			cfg:  Config{Auto: true, MaxWorkers: 0},
			want: min(cpus, DefaultMaxWorkers),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluationWorkers(tt.cfg)
			if got != tt.want {
				t.Errorf("EvaluationWorkers() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestOptimalWorkersAlwaysPositive(t *testing.T) {
	// Edge cases that should never return 0 or negative
	cases := []struct {
		cfg           Config
		resourceCount int
	}{
		{Config{Auto: true, MaxWorkers: 0}, 0},
		{Config{Auto: false, MaxWorkers: 0}, 0},
		{Config{Auto: true, MaxWorkers: -1}, 0},
		{Config{Auto: true, MaxWorkers: 1}, -5},
	}

	for _, tc := range cases {
		got := OptimalWorkers(tc.cfg, tc.resourceCount)
		if got < 1 {
			t.Errorf("OptimalWorkers(%+v, %d) = %d, want >= 1", tc.cfg, tc.resourceCount, got)
		}
	}
}
