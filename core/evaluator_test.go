package core

import (
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
