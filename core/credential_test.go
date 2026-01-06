package core

import (
	"os"
	"testing"
)

func TestCredentialType_String(t *testing.T) {
	tests := []struct {
		name string
		ct   CredentialType
		want string
	}{
		{"password", CredentialTypePassword, "password"},
		{"private_key", CredentialTypePrivateKey, "private_key"},
		{"bearer", CredentialTypeBearer, "bearer"},
		{"env", CredentialTypeEnv, "env"},
		{"undefined", CredentialTypeUndefined, "undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.String(); got != tt.want {
				t.Errorf("CredentialType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCredentialType(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want CredentialType
	}{
		{"password", "password", CredentialTypePassword},
		{"PASSWORD", "PASSWORD", CredentialTypePassword},
		{"private_key", "private_key", CredentialTypePrivateKey},
		{"bearer", "bearer", CredentialTypeBearer},
		{"env", "env", CredentialTypeEnv},
		{"unknown", "unknown", CredentialTypeUndefined},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCredentialType(tt.s); got != tt.want {
				t.Errorf("ParseCredentialType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredential_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cred    *Credential
		wantErr bool
	}{
		{
			name: "valid password",
			cred: &Credential{
				Type:   CredentialTypePassword,
				User:   "testuser",
				Secret: "testpass",
			},
			wantErr: false,
		},
		{
			name: "invalid password - no user",
			cred: &Credential{
				Type:   CredentialTypePassword,
				Secret: "testpass",
			},
			wantErr: true,
		},
		{
			name: "invalid password - no secret",
			cred: &Credential{
				Type: CredentialTypePassword,
				User: "testuser",
			},
			wantErr: true,
		},
		{
			name: "valid bearer",
			cred: &Credential{
				Type:   CredentialTypeBearer,
				Secret: "token123",
			},
			wantErr: false,
		},
		{
			name: "invalid bearer - no secret",
			cred: &Credential{
				Type: CredentialTypeBearer,
			},
			wantErr: true,
		},
		{
			name: "valid env",
			cred: &Credential{
				Type:       CredentialTypeEnv,
				EnvVarName: "TEST_TOKEN",
			},
			wantErr: false,
		},
		{
			name: "invalid env - no var name",
			cred: &Credential{
				Type: CredentialTypeEnv,
			},
			wantErr: true,
		},
		{
			name: "valid ssh agent",
			cred: &Credential{
				Type: CredentialTypeSSHAgent,
			},
			wantErr: false,
		},
		{
			name: "valid private key with path",
			cred: &Credential{
				Type:           CredentialTypePrivateKey,
				PrivateKeyPath: "/path/to/key",
			},
			wantErr: false,
		},
		{
			name: "valid private key with data",
			cred: &Credential{
				Type:       CredentialTypePrivateKey,
				PrivateKey: []byte("key-data"),
			},
			wantErr: false,
		},
		{
			name: "invalid private key - no data or path",
			cred: &Credential{
				Type: CredentialTypePrivateKey,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cred.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Credential.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCredential_ResolveSecret(t *testing.T) {
	// Set up test environment variable
	testEnvVar := "TEST_SECRET_VAR"
	testEnvValue := "secret-from-env"
	os.Setenv(testEnvVar, testEnvValue)
	defer os.Unsetenv(testEnvVar)

	tests := []struct {
		name    string
		cred    *Credential
		want    string
		wantErr bool
	}{
		{
			name: "resolve from env",
			cred: &Credential{
				Type:       CredentialTypeEnv,
				EnvVarName: testEnvVar,
			},
			want:    testEnvValue,
			wantErr: false,
		},
		{
			name: "resolve bearer token",
			cred: &Credential{
				Type:   CredentialTypeBearer,
				Secret: "direct-token",
			},
			want:    "direct-token",
			wantErr: false,
		},
		{
			name: "resolve password secret",
			cred: &Credential{
				Type:   CredentialTypePassword,
				Secret: "password123",
			},
			want:    "password123",
			wantErr: false,
		},
		{
			name: "error - env var not set",
			cred: &Credential{
				Type:       CredentialTypeEnv,
				EnvVarName: "NONEXISTENT_VAR",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cred.ResolveSecret()
			if (err != nil) != tt.wantErr {
				t.Errorf("Credential.ResolveSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Credential.ResolveSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCredentialFromConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]string
		want    *Credential
		wantErr bool
	}{
		{
			name: "password credentials",
			config: map[string]string{
				"credential_type": "password",
				"user":            "testuser",
				"secret":          "testpass",
			},
			want: &Credential{
				Type:   CredentialTypePassword,
				User:   "testuser",
				Secret: "testpass",
			},
			wantErr: false,
		},
		{
			name: "bearer token",
			config: map[string]string{
				"credential_type": "bearer",
				"secret":          "tokenvalue",
			},
			want: &Credential{
				Type:   CredentialTypeBearer,
				Secret: "tokenvalue",
			},
			wantErr: false,
		},
		{
			name: "bearer token alternative key",
			config: map[string]string{
				"credential_type": "bearer",
				"token":           "tokenvalue",
			},
			want: &Credential{
				Type:   CredentialTypeBearer,
				Secret: "tokenvalue",
			},
			wantErr: false,
		},
		{
			name: "legacy token support",
			config: map[string]string{
				"token": "legacy-token",
			},
			want: &Credential{
				Type:   CredentialTypeBearer,
				Secret: "legacy-token",
			},
			wantErr: false,
		},
		{
			name: "env credentials",
			config: map[string]string{
				"credential_type": "env",
				"env_var":         "MY_TOKEN",
			},
			want: &Credential{
				Type:       CredentialTypeEnv,
				EnvVarName: "MY_TOKEN",
			},
			wantErr: false,
		},
		{
			name: "private key with path",
			config: map[string]string{
				"credential_type":      "private_key",
				"private_key_path":     "/path/to/key",
				"private_key_password": "keypass",
			},
			want: &Credential{
				Type:               CredentialTypePrivateKey,
				PrivateKeyPath:     "/path/to/key",
				PrivateKeyPassword: "keypass",
			},
			wantErr: false,
		},
		{
			name: "no credential type",
			config: map[string]string{
				"user": "testuser",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid credential - missing required fields",
			config: map[string]string{
				"credential_type": "password",
				"user":            "testuser",
				// missing secret
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCredentialFromConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCredentialFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Type != tt.want.Type {
				t.Errorf("ParseCredentialFromConfig() Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.User != tt.want.User {
				t.Errorf("ParseCredentialFromConfig() User = %v, want %v", got.User, tt.want.User)
			}
			if got.Secret != tt.want.Secret {
				t.Errorf("ParseCredentialFromConfig() Secret = %v, want %v", got.Secret, tt.want.Secret)
			}
		})
	}
}
