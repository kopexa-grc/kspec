// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package cel

import (
	"sync"
	"testing"

	"github.com/kopexa-grc/kspec/core"
)

func TestEvaluator_Evaluate(t *testing.T) {
	tests := []struct {
		name     string
		policy   string
		resource core.Resource
		want     bool
		wantErr  bool
	}{
		{
			name:     "Simple boolean true",
			policy:   "true",
			resource: core.Resource{},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Simple boolean false",
			policy:   "false",
			resource: core.Resource{},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Check resource field equality",
			policy:   "resource.name == 'test-repo'",
			resource: core.Resource{"name": "test-repo"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Check resource field inequality",
			policy:   "resource.private == false",
			resource: core.Resource{"private": true},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Macro has() check",
			policy:   "has(resource.tags)",
			resource: core.Resource{"tags": []string{"production"}},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Macro has() check missing",
			policy:   "has(resource.tags)",
			resource: core.Resource{"name": "no-tags"},
			want:     false,
			wantErr:  false,
		},
		{
			name:   "Collection all() check",
			policy: "resource.collaborators.all(c, c.role != 'admin')",
			resource: core.Resource{
				"collaborators": []map[string]interface{}{
					{"user": "alice", "role": "write"},
					{"user": "bob", "role": "read"},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "Collection all() check fail",
			policy: "resource.collaborators.all(c, c.role != 'admin')",
			resource: core.Resource{
				"collaborators": []map[string]interface{}{
					{"user": "alice", "role": "admin"},
				},
			},
			want:    false,
			wantErr: false,
		},
	}

	evaluator, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, err := evaluator.Evaluate(tt.policy, tt.resource, nil, nil, core.Asset{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluator.Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if pass != tt.want {
				t.Errorf("Evaluator.Evaluate() = %v, want %v", pass, tt.want)
			}
		})
	}
}

func TestEvaluator_ComplexQueries(t *testing.T) {
	evaluator, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	tests := []struct {
		name     string
		policy   string
		resource core.Resource
		config   map[string]string
		asset    core.Asset
		want     bool
		wantErr  bool
	}{
		{
			name:   "Exists check on list",
			policy: "resource.rules.exists(r, r.action == 'allow')",
			resource: core.Resource{
				"rules": []map[string]interface{}{
					{"action": "deny", "port": 22},
					{"action": "allow", "port": 443},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "Exists check on list - not found",
			policy: "resource.rules.exists(r, r.action == 'block')",
			resource: core.Resource{
				"rules": []map[string]interface{}{
					{"action": "deny", "port": 22},
					{"action": "allow", "port": 443},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name:     "Numeric comparison",
			policy:   "resource.count > 10",
			resource: core.Resource{"count": 15},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Numeric comparison - equal edge case",
			policy:   "resource.count >= 10",
			resource: core.Resource{"count": 10},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "String contains",
			policy:   "resource.description.contains('security')",
			resource: core.Resource{"description": "This is a security policy"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "String startsWith",
			policy:   "resource.name.startsWith('prod-')",
			resource: core.Resource{"name": "prod-server-1"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "String endsWith",
			policy:   "resource.domain.endsWith('.com')",
			resource: core.Resource{"domain": "example.com"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Logical AND",
			policy:   "resource.enabled == true && resource.verified == true",
			resource: core.Resource{"enabled": true, "verified": true},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Logical AND - partial false",
			policy:   "resource.enabled == true && resource.verified == true",
			resource: core.Resource{"enabled": true, "verified": false},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Logical OR",
			policy:   "resource.admin == true || resource.superuser == true",
			resource: core.Resource{"admin": false, "superuser": true},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Negation",
			policy:   "!resource.public",
			resource: core.Resource{"public": false},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Ternary/conditional",
			policy:   "resource.type == 'admin' ? resource.mfa_enabled : true",
			resource: core.Resource{"type": "admin", "mfa_enabled": true},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Ternary - non-admin",
			policy:   "resource.type == 'admin' ? resource.mfa_enabled : true",
			resource: core.Resource{"type": "user", "mfa_enabled": false},
			want:     true,
			wantErr:  false,
		},
		{
			name:   "Size check on list",
			policy: "size(resource.members) > 0",
			resource: core.Resource{
				"members": []string{"alice", "bob"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "Size check on empty list",
			policy: "size(resource.members) == 0",
			resource: core.Resource{
				"members": []string{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name:     "In operator with list",
			policy:   "resource.status in ['active', 'pending']",
			resource: core.Resource{"status": "active"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "In operator - not in list",
			policy:   "resource.status in ['active', 'pending']",
			resource: core.Resource{"status": "disabled"},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Config variable access",
			policy:   "config.threshold == '100'",
			resource: core.Resource{},
			config:   map[string]string{"threshold": "100"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Asset type check",
			policy:   "asset.type == 'github-org'",
			resource: core.Resource{},
			asset:    core.Asset{Type: "github-org"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Combined resource and asset check",
			policy:   "resource.private == false || asset.type == 'internal'",
			resource: core.Resource{"private": true},
			asset:    core.Asset{Type: "internal"},
			want:     true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, err := evaluator.Evaluate(tt.policy, tt.resource, tt.config, nil, tt.asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluator.Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if pass != tt.want {
				t.Errorf("Evaluator.Evaluate() = %v, want %v", pass, tt.want)
			}
		})
	}
}

func TestEvaluator_InvalidExpressions(t *testing.T) {
	evaluator, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	tests := []struct {
		name       string
		expression string
	}{
		{
			name:       "Invalid syntax - unclosed bracket",
			expression: "resource.name[",
		},
		{
			name:       "Invalid syntax - unclosed paren",
			expression: "has(resource.name",
		},
		{
			name:       "Undeclared variable",
			expression: "unknown_var == true",
		},
		{
			name:       "Invalid operator",
			expression: "resource.name === 'test'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.Evaluate(tt.expression, core.Resource{}, nil, nil, core.Asset{})
			if err == nil {
				t.Errorf("Evaluate(%q) expected error, got nil", tt.expression)
			}
		})
	}
}

func TestEvaluator_ReverseIP(t *testing.T) {
	evaluator, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	tests := []struct {
		name     string
		policy   string
		resource core.Resource
		want     bool
		wantErr  bool
	}{
		{
			name:     "Reverse valid IPv4",
			policy:   "reverseIP('192.168.1.1') == '1.1.168.192.in-addr.arpa'",
			resource: core.Resource{},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Reverse another IPv4",
			policy:   "reverseIP('10.0.0.1') == '1.0.0.10.in-addr.arpa'",
			resource: core.Resource{},
			want:     true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, err := evaluator.Evaluate(tt.policy, tt.resource, nil, nil, core.Asset{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if pass != tt.want {
				t.Errorf("Evaluate() = %v, want %v", pass, tt.want)
			}
		})
	}
}

func TestEvaluator_CacheStats(t *testing.T) {
	evaluator, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	// Initially cache should be empty
	if stats := evaluator.CacheStats(); stats != 0 {
		t.Errorf("Initial cache size = %d, want 0", stats)
	}

	// Evaluate an expression
	_, _ = evaluator.Evaluate("true", core.Resource{}, nil, nil, core.Asset{})
	if stats := evaluator.CacheStats(); stats != 1 {
		t.Errorf("Cache size after 1 eval = %d, want 1", stats)
	}

	// Same expression should hit cache
	_, _ = evaluator.Evaluate("true", core.Resource{}, nil, nil, core.Asset{})
	if stats := evaluator.CacheStats(); stats != 1 {
		t.Errorf("Cache size after cache hit = %d, want 1", stats)
	}

	// Different expression should add to cache
	_, _ = evaluator.Evaluate("false", core.Resource{}, nil, nil, core.Asset{})
	if stats := evaluator.CacheStats(); stats != 2 {
		t.Errorf("Cache size after 2 unique evals = %d, want 2", stats)
	}
}

func TestEvaluator_ConcurrentAccess(t *testing.T) {
	evaluator, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	var wg sync.WaitGroup
	expressions := []string{
		"resource.a == true",
		"resource.b == true",
		"resource.c == true",
		"resource.d == true",
	}

	// Run concurrent evaluations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			expr := expressions[idx%len(expressions)]
			_, _ = evaluator.Evaluate(expr, core.Resource{
				"a": true, "b": true, "c": true, "d": true,
			}, nil, nil, core.Asset{})
		}(i)
	}

	wg.Wait()

	// Should have cached all unique expressions
	if stats := evaluator.CacheStats(); stats != len(expressions) {
		t.Errorf("Cache size after concurrent evals = %d, want %d", stats, len(expressions))
	}
}

func TestNewEvaluator(t *testing.T) {
	// Test creating evaluator with nil registry
	eval, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("NewEvaluator(nil) error = %v", err)
	}
	if eval == nil {
		t.Fatal("NewEvaluator(nil) returned nil evaluator")
	}
	if eval.Env == nil {
		t.Error("NewEvaluator(nil) returned evaluator with nil Env")
	}

	// Test creating evaluator with empty registry
	eval, err = NewEvaluator(map[string]core.ResourceSpec{})
	if err != nil {
		t.Fatalf("NewEvaluator(empty) error = %v", err)
	}
	if eval == nil {
		t.Error("NewEvaluator(empty) returned nil evaluator")
	}
}

// FuzzEvaluate tests the evaluator with random CEL expressions.
// This helps find crashes, panics, or unexpected behavior with malformed input.
func FuzzEvaluate(f *testing.F) {
	// Seed corpus with valid and edge-case expressions
	seeds := []string{
		"true",
		"false",
		"resource.name == 'test'",
		"has(resource.field)",
		"resource.count > 10",
		"size(resource.list) == 0",
		"resource.status in ['a', 'b']",
		"!resource.enabled",
		"resource.a && resource.b",
		"resource.a || resource.b",
		"resource.x ? resource.y : resource.z",
		"",
		"   ",
		"(((((",
		")))))",
		"resource.",
		".resource",
		"resource..name",
		"'unclosed string",
		"resource['key']",
		"resource.name.contains('')",
		"resource.name.startsWith('')",
		"resource.name.endsWith('')",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	evaluator, err := NewEvaluator(nil)
	if err != nil {
		f.Fatalf("Failed to create evaluator: %v", err)
	}

	f.Fuzz(func(t *testing.T, expression string) {
		// The evaluator should never panic, regardless of input
		// Errors are expected for invalid expressions
		resource := core.Resource{
			"name":    "test",
			"count":   10,
			"enabled": true,
			"list":    []string{"a", "b"},
			"status":  "active",
			"a":       true,
			"b":       false,
			"x":       true,
			"y":       "yes",
			"z":       "no",
		}

		// This should not panic
		_, _ = evaluator.Evaluate(expression, resource, nil, nil, core.Asset{})
	})
}

// FuzzReverseIP tests the reverseIP function with random input strings.
func FuzzReverseIP(f *testing.F) {
	// Seed corpus
	seeds := []string{
		"192.168.1.1",
		"10.0.0.1",
		"255.255.255.255",
		"0.0.0.0",
		"",
		"...",
		"1.2.3",
		"1.2.3.4.5",
		"abc.def.ghi.jkl",
		"-1.-1.-1.-1",
		"256.256.256.256",
		"1",
		"1.2",
		"1.2.3.4.",
		".1.2.3.4",
		"    ",
		"1 .2 .3 .4",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	evaluator, err := NewEvaluator(nil)
	if err != nil {
		f.Fatalf("Failed to create evaluator: %v", err)
	}

	f.Fuzz(func(t *testing.T, ip string) {
		// Build expression that uses reverseIP
		// The evaluator should handle any input gracefully
		expr := "reverseIP('" + escapeString(ip) + "') == 'test'"

		// This should not panic
		_, _ = evaluator.Evaluate(expr, core.Resource{}, nil, nil, core.Asset{})
	})
}

// escapeString escapes special characters for CEL string literals.
func escapeString(s string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			result = append(result, '\\', '\\')
		case '\'':
			result = append(result, '\\', '\'')
		case '\n':
			result = append(result, '\\', 'n')
		case '\r':
			result = append(result, '\\', 'r')
		case '\t':
			result = append(result, '\\', 't')
		default:
			result = append(result, s[i])
		}
	}
	return string(result)
}

// TestEvaluator_EdgeCases tests edge cases that might cause issues.
func TestEvaluator_EdgeCases(t *testing.T) {
	evaluator, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	tests := []struct {
		name       string
		expression string
		resource   core.Resource
		wantErr    bool
	}{
		{
			name:       "Empty expression",
			expression: "",
			resource:   core.Resource{},
			wantErr:    true,
		},
		{
			name:       "Whitespace only",
			expression: "   ",
			resource:   core.Resource{},
			wantErr:    true,
		},
		{
			name:       "Very long expression",
			expression: "resource.name == '" + string(make([]byte, 10000)) + "'",
			resource:   core.Resource{"name": "test"},
			wantErr:    false,
		},
		{
			name:       "Deeply nested access",
			expression: "resource.a.b.c.d.e.f.g == true",
			resource: core.Resource{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"d": map[string]interface{}{
								"e": map[string]interface{}{
									"f": map[string]interface{}{
										"g": true,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "Null value comparison",
			expression: "resource.value == null",
			resource:   core.Resource{"value": nil},
			wantErr:    false,
		},
		{
			name:       "Empty string comparison",
			expression: "resource.name == ''",
			resource:   core.Resource{"name": ""},
			wantErr:    false,
		},
		{
			name:       "Unicode in expression",
			expression: "resource.name == '日本語'",
			resource:   core.Resource{"name": "日本語"},
			wantErr:    false,
		},
		{
			name:       "Special characters in string",
			expression: "resource.name == 'test\\nwith\\nnewlines'",
			resource:   core.Resource{"name": "test\nwith\nnewlines"},
			wantErr:    false,
		},
		{
			name:       "Number as string field",
			expression: "resource.id == '123'",
			resource:   core.Resource{"id": "123"},
			wantErr:    false,
		},
		{
			name:       "Boolean string comparison",
			expression: "resource.flag == 'true'",
			resource:   core.Resource{"flag": "true"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.Evaluate(tt.expression, tt.resource, nil, nil, core.Asset{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%q) error = %v, wantErr %v", tt.expression, err, tt.wantErr)
			}
		})
	}
}

// TestEvaluator_ReverseIPEdgeCases tests edge cases for reverseIP.
func TestEvaluator_ReverseIPEdgeCases(t *testing.T) {
	evaluator, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	tests := []struct {
		name    string
		policy  string
		wantErr bool
	}{
		{
			name:    "Empty IP",
			policy:  "reverseIP('')",
			wantErr: true,
		},
		{
			name:    "Too few octets",
			policy:  "reverseIP('1.2.3')",
			wantErr: true,
		},
		{
			name:    "Too many octets",
			policy:  "reverseIP('1.2.3.4.5')",
			wantErr: true,
		},
		{
			name:    "Non-numeric octets",
			policy:  "reverseIP('a.b.c.d')",
			wantErr: false, // Current implementation doesn't validate numeric values
		},
		{
			name:    "Just dots",
			policy:  "reverseIP('...')",
			wantErr: true,
		},
		{
			name:    "Leading dot",
			policy:  "reverseIP('.1.2.3.4')",
			wantErr: true,
		},
		{
			name:    "Trailing dot",
			policy:  "reverseIP('1.2.3.4.')",
			wantErr: true,
		},
		{
			name:    "Whitespace around IP (trimmed)",
			policy:  "reverseIP(' 1.2.3.4 ')",
			wantErr: false, // Whitespace is trimmed, so this is valid
		},
		{
			name:    "Whitespace within IP",
			policy:  "reverseIP('1. 2.3.4')",
			wantErr: false, // Current implementation doesn't validate octet content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wrap in a comparison to get boolean result
			expr := tt.policy + " == 'test'"
			_, err := evaluator.Evaluate(expr, core.Resource{}, nil, nil, core.Asset{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate(%q) error = %v, wantErr %v", tt.policy, err, tt.wantErr)
			}
		})
	}
}

func TestEvaluator_NestedMapAccess(t *testing.T) {
	evaluator, err := NewEvaluator(nil)
	if err != nil {
		t.Fatalf("Failed to create evaluator: %v", err)
	}

	tests := []struct {
		name     string
		policy   string
		resource core.Resource
		want     bool
	}{
		{
			name:   "Nested map access",
			policy: "resource.settings.security.enabled == true",
			resource: core.Resource{
				"settings": map[string]interface{}{
					"security": map[string]interface{}{
						"enabled": true,
					},
				},
			},
			want: true,
		},
		{
			name:   "Has on nested map",
			policy: "has(resource.config.api_key)",
			resource: core.Resource{
				"config": map[string]interface{}{
					"api_key": "secret",
				},
			},
			want: true,
		},
		{
			name:   "Has on nested map - missing",
			policy: "has(resource.config.api_key)",
			resource: core.Resource{
				"config": map[string]interface{}{
					"api_url": "https://example.com",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, err := evaluator.Evaluate(tt.policy, tt.resource, nil, nil, core.Asset{})
			if err != nil {
				t.Errorf("Evaluate() error = %v", err)
				return
			}
			if pass != tt.want {
				t.Errorf("Evaluate() = %v, want %v", pass, tt.want)
			}
		})
	}
}
