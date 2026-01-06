package core

import (
	"sync"
	"testing"
)

func TestEvaluator_Evaluate(t *testing.T) {
	tests := []struct {
		name     string
		policy   string
		resource Resource
		want     bool
		wantErr  bool
	}{
		{
			name:     "Simple boolean true",
			policy:   "true",
			resource: Resource{},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Simple boolean false",
			policy:   "false",
			resource: Resource{},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Check resource field equality",
			policy:   "resource.name == 'test-repo'",
			resource: Resource{"name": "test-repo"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Check resource field inequality",
			policy:   "resource.private == false",
			resource: Resource{"private": true},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Macro has() check",
			policy:   "has(resource.tags)",
			resource: Resource{"tags": []string{"production"}},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Macro has() check missing",
			policy:   "has(resource.tags)",
			resource: Resource{"name": "no-tags"},
			want:     false,
			wantErr:  false,
		},
		{
			name:   "Collection all() check",
			policy: "resource.collaborators.all(c, c.role != 'admin')",
			resource: Resource{
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
			resource: Resource{
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
			pass, err := evaluator.Evaluate(tt.policy, tt.resource, nil, nil, Asset{})
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
		resource Resource
		config   map[string]string
		asset    Asset
		want     bool
		wantErr  bool
	}{
		{
			name:   "Exists check on list",
			policy: "resource.rules.exists(r, r.action == 'allow')",
			resource: Resource{
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
			resource: Resource{
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
			resource: Resource{"count": 15},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Numeric comparison - equal edge case",
			policy:   "resource.count >= 10",
			resource: Resource{"count": 10},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "String contains",
			policy:   "resource.description.contains('security')",
			resource: Resource{"description": "This is a security policy"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "String startsWith",
			policy:   "resource.name.startsWith('prod-')",
			resource: Resource{"name": "prod-server-1"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "String endsWith",
			policy:   "resource.domain.endsWith('.com')",
			resource: Resource{"domain": "example.com"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Logical AND",
			policy:   "resource.enabled == true && resource.verified == true",
			resource: Resource{"enabled": true, "verified": true},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Logical AND - partial false",
			policy:   "resource.enabled == true && resource.verified == true",
			resource: Resource{"enabled": true, "verified": false},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Logical OR",
			policy:   "resource.admin == true || resource.superuser == true",
			resource: Resource{"admin": false, "superuser": true},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Negation",
			policy:   "!resource.public",
			resource: Resource{"public": false},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Ternary/conditional",
			policy:   "resource.type == 'admin' ? resource.mfa_enabled : true",
			resource: Resource{"type": "admin", "mfa_enabled": true},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Ternary - non-admin",
			policy:   "resource.type == 'admin' ? resource.mfa_enabled : true",
			resource: Resource{"type": "user", "mfa_enabled": false},
			want:     true,
			wantErr:  false,
		},
		{
			name:   "Size check on list",
			policy: "size(resource.members) > 0",
			resource: Resource{
				"members": []string{"alice", "bob"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "Size check on empty list",
			policy: "size(resource.members) == 0",
			resource: Resource{
				"members": []string{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name:     "In operator with list",
			policy:   "resource.status in ['active', 'pending']",
			resource: Resource{"status": "active"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "In operator - not in list",
			policy:   "resource.status in ['active', 'pending']",
			resource: Resource{"status": "disabled"},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Config variable access",
			policy:   "config.threshold == '100'",
			resource: Resource{},
			config:   map[string]string{"threshold": "100"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Asset type check",
			policy:   "asset.type == 'github-org'",
			resource: Resource{},
			asset:    Asset{Type: "github-org"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Combined resource and asset check",
			policy:   "resource.private == false || asset.type == 'internal'",
			resource: Resource{"private": true},
			asset:    Asset{Type: "internal"},
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
			_, err := evaluator.Evaluate(tt.expression, Resource{}, nil, nil, Asset{})
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
		resource Resource
		want     bool
		wantErr  bool
	}{
		{
			name:     "Reverse valid IPv4",
			policy:   "reverseIP('192.168.1.1') == '1.1.168.192.in-addr.arpa'",
			resource: Resource{},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Reverse another IPv4",
			policy:   "reverseIP('10.0.0.1') == '1.0.0.10.in-addr.arpa'",
			resource: Resource{},
			want:     true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, err := evaluator.Evaluate(tt.policy, tt.resource, nil, nil, Asset{})
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
	_, _ = evaluator.Evaluate("true", Resource{}, nil, nil, Asset{})
	if stats := evaluator.CacheStats(); stats != 1 {
		t.Errorf("Cache size after 1 eval = %d, want 1", stats)
	}

	// Same expression should hit cache
	_, _ = evaluator.Evaluate("true", Resource{}, nil, nil, Asset{})
	if stats := evaluator.CacheStats(); stats != 1 {
		t.Errorf("Cache size after cache hit = %d, want 1", stats)
	}

	// Different expression should add to cache
	_, _ = evaluator.Evaluate("false", Resource{}, nil, nil, Asset{})
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
			_, _ = evaluator.Evaluate(expr, Resource{
				"a": true, "b": true, "c": true, "d": true,
			}, nil, nil, Asset{})
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
	eval, err = NewEvaluator(map[string]ResourceSpec{})
	if err != nil {
		t.Fatalf("NewEvaluator(empty) error = %v", err)
	}
	if eval == nil {
		t.Error("NewEvaluator(empty) returned nil evaluator")
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
		resource Resource
		want     bool
	}{
		{
			name:   "Nested map access",
			policy: "resource.settings.security.enabled == true",
			resource: Resource{
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
			resource: Resource{
				"config": map[string]interface{}{
					"api_key": "secret",
				},
			},
			want: true,
		},
		{
			name:   "Has on nested map - missing",
			policy: "has(resource.config.api_key)",
			resource: Resource{
				"config": map[string]interface{}{
					"api_url": "https://example.com",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, err := evaluator.Evaluate(tt.policy, tt.resource, nil, nil, Asset{})
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
