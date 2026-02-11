// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Example: High-availability worker that processes multiple integrations concurrently.
//
// This example demonstrates the full Kopexa integration lifecycle:
//
//  1. Discovery — What assets exist?
//  2. Asset Sync — Track each asset with a stable fingerprint
//  3. Scan — Evaluate policies against discovered assets
//  4. Result Collection — Results keyed by fingerprint for platform mapping
//
// Usage:
//
//	go run ./_examples/worker/
//	KSPEC_WORKERS=10 go run ./_examples/worker/
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/discovery"
	"github.com/kopexa-grc/kspec/policy"
	"github.com/kopexa-grc/kspec/policy/scoring"
	"github.com/kopexa-grc/kspec/provider/scanner"

	// IMPORTANT: Register all providers via blank import.
	_ "github.com/kopexa-grc/kspec/provider/all"
)

// ---------------------------------------------------------------------------
// Disabled queries (per asset)
// ---------------------------------------------------------------------------

// disabledQueries maps asset fingerprints to query UIDs that should be skipped.
// In production, this would come from a database or API.
var disabledQueries = map[string][]string{
	AssetFingerprint("network", "host", "example.com"):    {"tls-cert-not-expiring"},
	AssetFingerprint("network", "host", "cloudflare.com"): {"tls-modern-version", "tls-cert-not-expiring"},
}

// filterDisabledQueries returns a deep copy of the policies with checks matching
// the disabled UIDs removed from both Groups[].Checks and top-level Queries.
// The original slice is not modified so it can be reused across integrations.
func filterDisabledQueries(policies []policy.Policy, disabled []string) []policy.Policy {
	if len(disabled) == 0 {
		return policies
	}

	set := make(map[string]struct{}, len(disabled))
	for _, uid := range disabled {
		set[uid] = struct{}{}
	}

	out := make([]policy.Policy, len(policies))
	copy(out, policies)

	for i, p := range out {
		// Filter groups and their checks.
		var groups []policy.Group
		for _, g := range p.Groups {
			var checks []policy.Check
			for _, c := range g.Checks {
				if _, skip := set[c.UID]; !skip {
					checks = append(checks, c)
				}
			}
			if len(checks) > 0 {
				g.Checks = checks
				groups = append(groups, g)
			}
		}
		out[i].Groups = groups

		// Filter top-level queries.
		var queries []policy.Check
		for _, q := range p.Queries {
			if _, skip := set[q.UID]; !skip {
				queries = append(queries, q)
			}
		}
		out[i].Queries = queries
	}

	return out
}

// ---------------------------------------------------------------------------
// Asset Store
// ---------------------------------------------------------------------------

// AssetRecord represents a tracked asset in Kopexa's asset inventory.
type AssetRecord struct {
	Fingerprint string    // Stable ID for Kopexa mapping (sha256-based)
	Provider    string    // "network", "github", "ms365"
	AssetType   string    // "host", "github-org", ...
	AssetName   string    // "example.com", "my-org", ...
	FirstSeen   time.Time // When first discovered
	LastSeen    time.Time // When last discovered
	LastScore   *uint32   // Last scan score (nil = never scanned)
	LastGrade   string    // Last scan grade
}

// AssetStore persists asset records. Swap the in-memory implementation
// for a database, Redis, or API client in production.
type AssetStore interface {
	Upsert(record AssetRecord) error
	Get(fingerprint string) (*AssetRecord, bool)
	List() []AssetRecord
}

// memoryStore is the default in-memory implementation.
type memoryStore struct {
	mu      sync.RWMutex
	records map[string]AssetRecord
}

func newMemoryStore() *memoryStore {
	return &memoryStore{records: make(map[string]AssetRecord)}
}

func (m *memoryStore) Upsert(record AssetRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.records[record.Fingerprint]; ok {
		record.FirstSeen = existing.FirstSeen // preserve original first_seen
	}

	m.records[record.Fingerprint] = record

	return nil
}

func (m *memoryStore) Get(fingerprint string) (*AssetRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.records[fingerprint]
	if !ok {
		return nil, false
	}

	return &r, true
}

func (m *memoryStore) List() []AssetRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]AssetRecord, 0, len(m.records))
	for _, r := range m.records {
		out = append(out, r)
	}

	return out
}

// ---------------------------------------------------------------------------
// Job types
// ---------------------------------------------------------------------------

// Integration represents one customer integration (e.g., one OAuth connection).
type Integration struct {
	ID         string            // Kopexa integration ID
	Provider   string            // Provider name
	Config     map[string]string // Credentials / settings
	AssetType  string            // What to discover
	AssetName  string            // Target name
	PolicyPath string            // Policy file
	Timeout    time.Duration     // Per-integration deadline
}

// IntegrationResult is the outcome of processing one integration.
type IntegrationResult struct {
	IntegrationID string
	Fingerprint   string // Stable asset ID
	Provider      string
	Asset         string
	// Discovery
	ResourceTypes  int
	TotalResources int
	// Scan
	Score         uint32
	Grade         string
	RiskLevel     string
	Findings      scoring.Findings
	DisabledCount int // Number of queries disabled for this asset
	Duration      time.Duration
	Phase         string // "discovery", "sync", "scan", "complete"
	Err           error
}

// ---------------------------------------------------------------------------
// Fingerprint
// ---------------------------------------------------------------------------

// AssetFingerprint generates a stable identifier for an asset.
// This allows Kopexa to map scan results back to the correct entity.
func AssetFingerprint(provider, assetType, assetName string) string {
	h := sha256.Sum256([]byte(provider + ":" + assetType + ":" + assetName))
	return hex.EncodeToString(h[:16])
}

// ---------------------------------------------------------------------------
// Worker pool
// ---------------------------------------------------------------------------

func processIntegrations(ctx context.Context, store AssetStore, integrations []Integration, numWorkers int) []IntegrationResult {
	jobs := make(chan Integration, len(integrations))
	results := make(chan IntegrationResult, len(integrations))

	// Start workers.
	var wg sync.WaitGroup

	for i := range numWorkers {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for job := range jobs {
				select {
				case <-ctx.Done():
					results <- IntegrationResult{
						IntegrationID: job.ID,
						Provider:      job.Provider,
						Asset:         job.AssetName,
						Phase:         "cancelled",
						Err:           ctx.Err(),
					}
				default:
					fmt.Fprintf(os.Stderr, "[worker-%d] Processing integration %s (%s/%s)\n",
						workerID, job.ID, job.Provider, job.AssetName)
					results <- processIntegration(ctx, store, job)
				}
			}
		}(i)
	}

	// Feed jobs.
	for _, integration := range integrations {
		jobs <- integration
	}

	close(jobs)

	// Wait for all workers, then close results.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results.
	var out []IntegrationResult
	for r := range results {
		out = append(out, r)
	}

	return out
}

// ---------------------------------------------------------------------------
// Per-integration pipeline
// ---------------------------------------------------------------------------

func processIntegration(ctx context.Context, store AssetStore, integration Integration) IntegrationResult {
	start := time.Now()

	result := IntegrationResult{
		IntegrationID: integration.ID,
		Provider:      integration.Provider,
		Asset:         integration.AssetName,
	}

	// Per-integration timeout.
	ctx, cancel := context.WithTimeout(ctx, integration.Timeout)
	defer cancel()

	// ---- Phase 1: DISCOVERY ----
	result.Phase = "discovery"

	fmt.Fprintf(os.Stderr, "[%s] Phase 1: Discovery (%s/%s)\n",
		integration.ID, integration.Provider, integration.AssetName)

	disc, err := discovery.Discover(ctx,
		integration.Provider,
		integration.Config,
		integration.AssetType,
		integration.AssetName,
	)
	if err != nil {
		result.Err = fmt.Errorf("discovery failed: %w", err)
		result.Duration = time.Since(start)

		return result
	}

	result.ResourceTypes = len(disc.Resources)
	result.TotalResources = disc.TotalCount

	fmt.Fprintf(os.Stderr, "[%s]   Found %d resource types, %d total resources\n",
		integration.ID, result.ResourceTypes, result.TotalResources)

	// ---- Phase 2: ASSET SYNC ----
	result.Phase = "sync"

	fingerprint := AssetFingerprint(integration.Provider, integration.AssetType, integration.AssetName)
	result.Fingerprint = fingerprint

	now := time.Now()

	existing, known := store.Get(fingerprint)
	if known {
		fmt.Fprintf(os.Stderr, "[%s] Phase 2: Asset sync (known asset, last seen %s ago)\n",
			integration.ID, now.Sub(existing.LastSeen).Truncate(time.Second))
	} else {
		fmt.Fprintf(os.Stderr, "[%s] Phase 2: Asset sync (new asset)\n", integration.ID)
	}

	record := AssetRecord{
		Fingerprint: fingerprint,
		Provider:    integration.Provider,
		AssetType:   integration.AssetType,
		AssetName:   integration.AssetName,
		FirstSeen:   now,
		LastSeen:    now,
	}

	if err := store.Upsert(record); err != nil {
		result.Err = fmt.Errorf("asset sync failed: %w", err)
		result.Duration = time.Since(start)

		return result
	}

	// ---- Phase 3: SCAN ----
	result.Phase = "scan"

	fmt.Fprintf(os.Stderr, "[%s] Phase 3: Scan\n", integration.ID)

	policies, err := policy.Load(integration.PolicyPath, "")
	if err != nil {
		result.Err = fmt.Errorf("policy load failed: %w", err)
		result.Duration = time.Since(start)

		return result
	}

	policies = policy.FilterByProvider(policies, integration.Provider, integration.Provider)
	if len(policies) == 0 {
		result.Err = fmt.Errorf("no policies found for provider %s", integration.Provider)
		result.Duration = time.Since(start)

		return result
	}

	// Apply per-asset disabled queries.
	if disabled, ok := disabledQueries[fingerprint]; ok {
		policies = filterDisabledQueries(policies, disabled)
		result.DisabledCount = len(disabled)

		fmt.Fprintf(os.Stderr, "[%s]   Disabled %d queries for asset %s: %v\n",
			integration.ID, len(disabled), integration.AssetName, disabled)
	}

	s := scanner.NewScanner(scanner.ScanConfig{
		ProviderName:   integration.Provider,
		ProviderConfig: integration.Config,
		Asset: core.Asset{
			Type:   integration.AssetType,
			Name:   integration.AssetName,
			Config: integration.Config,
		},
		Policies: policies,
	})

	s.OnEvent(func(event scanner.ScanEvent) {
		switch event.Type {
		case scanner.EventDiscoveryStarted:
			fmt.Fprintf(os.Stderr, "[%s]   Scanner: discovery started\n", integration.ID)
		case scanner.EventDiscoveryComplete:
			fmt.Fprintf(os.Stderr, "[%s]   Scanner: discovery complete\n", integration.ID)
		case scanner.EventResourceScanning:
			fmt.Fprintf(os.Stderr, "[%s]   Scanner: scanning %s\n", integration.ID, event.ResourceType)
		case scanner.EventResourceComplete:
			fmt.Fprintf(os.Stderr, "[%s]   Scanner: complete %s\n", integration.ID, event.ResourceType)
		case scanner.EventError:
			fmt.Fprintf(os.Stderr, "[%s]   Scanner: error: %v\n", integration.ID, event.Error)
		}
	})

	if err := s.Initialize(ctx); err != nil {
		result.Err = fmt.Errorf("scanner init failed: %w", err)
		result.Duration = time.Since(start)

		return result
	}

	scanResult := s.Run(ctx)

	for _, scanErr := range scanResult.Errors {
		fmt.Fprintf(os.Stderr, "[%s]   Scan error: %v\n", integration.ID, scanErr)
	}

	if scanResult.Tree == nil || scanResult.Tree.Root == nil {
		result.Err = fmt.Errorf("scan produced no results")
		result.Duration = time.Since(start)

		return result
	}

	// Score the results.
	scorer := scoring.NewGraphScorer(scoring.SystemBanded)
	score := scorer.ScoreTree(scanResult.Tree)
	findings := scorer.GetRootFindings(scanResult.Tree)

	result.Score = score.Value
	result.Grade = score.Grade
	result.RiskLevel = score.RiskLevel
	result.Findings = findings

	fmt.Fprintf(os.Stderr, "[%s]   Score: %d/100 (Grade: %s, Risk: %s)\n",
		integration.ID, score.Value, score.Grade, score.RiskLevel)

	// ---- Phase 4: UPDATE STORE ----
	scoreVal := score.Value

	record.LastScore = &scoreVal
	record.LastGrade = score.Grade
	record.LastSeen = time.Now()

	if err := store.Upsert(record); err != nil {
		result.Err = fmt.Errorf("store update failed: %w", err)
		result.Duration = time.Since(start)

		return result
	}

	result.Phase = "complete"
	result.Duration = time.Since(start)

	return result
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	// Determine worker count.
	numWorkers := runtime.NumCPU()
	if env := os.Getenv("KSPEC_WORKERS"); env != "" {
		if n, err := strconv.Atoi(env); err == nil && n > 0 {
			numWorkers = n
		}
	}

	// Policy path (bundled with this example).
	policyPath := policyFile()

	// Demo integrations: 5 public domains scanned with the network provider.
	// No credentials needed — the network provider uses public DNS/TLS/HTTP.
	integrations := []Integration{
		{ID: "int-001", Provider: "network", Config: map[string]string{"target": "example.com"}, AssetType: "host", AssetName: "example.com", PolicyPath: policyPath, Timeout: 2 * time.Minute},
		{ID: "int-002", Provider: "network", Config: map[string]string{"target": "google.com"}, AssetType: "host", AssetName: "google.com", PolicyPath: policyPath, Timeout: 2 * time.Minute},
		{ID: "int-003", Provider: "network", Config: map[string]string{"target": "github.com"}, AssetType: "host", AssetName: "github.com", PolicyPath: policyPath, Timeout: 2 * time.Minute},
		{ID: "int-004", Provider: "network", Config: map[string]string{"target": "cloudflare.com"}, AssetType: "host", AssetName: "cloudflare.com", PolicyPath: policyPath, Timeout: 2 * time.Minute},
		{ID: "int-005", Provider: "network", Config: map[string]string{"target": "mozilla.org"}, AssetType: "host", AssetName: "mozilla.org", PolicyPath: policyPath, Timeout: 2 * time.Minute},
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Fprintf(os.Stderr, "=== kspec Worker Example ===\n")
	fmt.Fprintf(os.Stderr, "Workers:      %d\n", numWorkers)
	fmt.Fprintf(os.Stderr, "Integrations: %d\n", len(integrations))
	fmt.Fprintf(os.Stderr, "Policy:       %s\n\n", policyPath)

	store := newMemoryStore()

	// Process all integrations.
	results := processIntegrations(ctx, store, integrations, numWorkers)

	// --- Summary ---
	fmt.Println()
	fmt.Println("=== Results ===")
	fmt.Println()

	header := fmt.Sprintf("%-34s %-18s %6s %6s %-10s %4s %4s %4s %4s %4s %10s %-10s",
		"Fingerprint", "Asset", "ResTyp", "ResAll", "Grade", "Skip", "Crit", "High", "Med", "Low", "Duration", "Phase")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	var succeeded, failed int

	for _, r := range results {
		if r.Err != nil {
			failed++

			fmt.Printf("%-34s %-18s %6s %6s %-10s %4s %4s %4s %4s %4s %10s %-10s  ERR: %v\n",
				truncate(r.Fingerprint, 34),
				truncate(r.Asset, 18),
				"-", "-", "-", "-", "-", "-", "-", "-",
				r.Duration.Truncate(time.Millisecond),
				r.Phase,
				r.Err,
			)

			continue
		}

		succeeded++

		fmt.Printf("%-34s %-18s %6d %6d %-10s %4d %4d %4d %4d %4d %10s %-10s\n",
			truncate(r.Fingerprint, 34),
			truncate(r.Asset, 18),
			r.ResourceTypes,
			r.TotalResources,
			fmt.Sprintf("%s (%d)", r.Grade, r.Score),
			r.DisabledCount,
			r.Findings.Critical,
			r.Findings.High,
			r.Findings.Medium,
			r.Findings.Low,
			r.Duration.Truncate(time.Millisecond),
			r.Phase,
		)
	}

	fmt.Println()
	fmt.Printf("Total: %d succeeded, %d failed out of %d integrations\n", succeeded, failed, len(results))

	// Show asset store contents.
	assets := store.List()
	if len(assets) > 0 {
		fmt.Println()
		fmt.Println("=== Asset Store ===")
		fmt.Println()
		fmt.Printf("%-34s %-10s %-8s %-18s %-6s %-6s\n",
			"Fingerprint", "Provider", "Type", "Name", "Score", "Grade")
		fmt.Println(strings.Repeat("-", 104))

		for _, a := range assets {
			scoreStr := "-"
			if a.LastScore != nil {
				scoreStr = strconv.FormatUint(uint64(*a.LastScore), 10)
			}

			gradeStr := "-"
			if a.LastGrade != "" {
				gradeStr = a.LastGrade
			}

			fmt.Printf("%-34s %-10s %-8s %-18s %-6s %-6s\n",
				truncate(a.Fingerprint, 34),
				a.Provider,
				a.AssetType,
				truncate(a.AssetName, 18),
				scoreStr,
				gradeStr,
			)
		}
	}

	if failed > 0 {
		os.Exit(1)
	}
}

// policyFile returns the path to the bundled policy.yaml next to this example.
func policyFile() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "policy.yaml")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	return s[:max-1] + "…"
}
