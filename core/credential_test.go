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

// FuzzParseCredentialType tests credential type parsing with random input.
func FuzzParseCredentialType(f *testing.F) {
	// Seed corpus with known types and edge cases
	seeds := []string{
		"password",
		"PASSWORD",
		"Password",
		"private_key",
		"PRIVATE_KEY",
		"ssh_agent",
		"bearer",
		"credentials_query",
		"json",
		"aws_ec2_instance_connect",
		"aws_ec2_ssm_session",
		"pkcs12",
		"env",
		"",
		"   ",
		"unknown",
		"password123",
		"bearer_token",
		"pass word",
		"PASSWORD\n",
		"password\x00",
		string(make([]byte, 1000)),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		result := ParseCredentialType(input)

		// Result should always be a valid CredentialType
		if result < CredentialTypeUndefined || result > CredentialTypeEnv {
			t.Errorf("ParseCredentialType(%q) returned invalid type: %d", input, result)
		}
	})
}

// FuzzCredentialValidate tests credential validation with random input.
func FuzzCredentialValidate(f *testing.F) {
	// Seed corpus with various field combinations
	seeds := []struct {
		credType int
		user     string
		secret   string
		keyPath  string
		envVar   string
		jsonData string
	}{
		{1, "user", "pass", "", "", ""},
		{1, "", "pass", "", "", ""},
		{1, "user", "", "", "", ""},
		{2, "", "", "/path/to/key", "", ""},
		{2, "", "", "", "", ""},
		{4, "", "token", "", "", ""},
		{4, "", "", "", "", ""},
		{10, "", "", "", "ENV_VAR", ""},
		{10, "", "", "", "", ""},
		{6, "", "", "", "", `{"key":"value"}`},
		{0, "", "", "", "", ""},
	}

	for _, seed := range seeds {
		f.Add(seed.credType, seed.user, seed.secret, seed.keyPath, seed.envVar, seed.jsonData)
	}

	f.Fuzz(func(t *testing.T, credType int, user, secret, keyPath, envVar, jsonData string) {
		cred := &Credential{
			Type:           CredentialType(credType),
			User:           user,
			Secret:         secret,
			PrivateKeyPath: keyPath,
			EnvVarName:     envVar,
			JSONData:       jsonData,
		}

		// Should never panic
		_ = cred.Validate()
	})
}

// FuzzParseCredentialFromConfig tests config parsing with random input.
func FuzzParseCredentialFromConfig(f *testing.F) {
	// Seed corpus
	seeds := []struct {
		credType string
		user     string
		secret   string
		token    string
		envVar   string
	}{
		{"password", "user", "pass", "", ""},
		{"bearer", "", "token", "", ""},
		{"bearer", "", "", "token", ""},
		{"env", "", "", "", "VAR"},
		{"", "", "", "legacy", ""},
		{"unknown", "", "", "", ""},
		{"", "", "", "", ""},
		{"private_key", "", "", "", ""},
	}

	for _, seed := range seeds {
		f.Add(seed.credType, seed.user, seed.secret, seed.token, seed.envVar)
	}

	f.Fuzz(func(t *testing.T, credType, user, secret, token, envVar string) {
		config := make(map[string]string)
		if credType != "" {
			config["credential_type"] = credType
		}
		if user != "" {
			config["user"] = user
		}
		if secret != "" {
			config["secret"] = secret
		}
		if token != "" {
			config["token"] = token
		}
		if envVar != "" {
			config["env_var"] = envVar
		}

		// Should never panic
		_, _ = ParseCredentialFromConfig(config)
	})
}

// TestCredentialType_EdgeCases tests edge cases in credential type handling.
func TestCredentialType_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  CredentialType
	}{
		{"mixed case", "PaSsWoRd", CredentialTypePassword},
		{"leading whitespace", " password", CredentialTypeUndefined},
		{"trailing whitespace", "password ", CredentialTypeUndefined},
		{"empty string", "", CredentialTypeUndefined},
		{"null byte", "password\x00", CredentialTypeUndefined},
		{"newline", "password\n", CredentialTypeUndefined},
		{"unicode", "密码", CredentialTypeUndefined},
		{"very long", string(make([]byte, 1000)), CredentialTypeUndefined},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCredentialType(tt.input)
			if got != tt.want {
				t.Errorf("ParseCredentialType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestCredential_ValidateEdgeCases tests edge cases in credential validation.
func TestCredential_ValidateEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		cred    *Credential
		wantErr bool
	}{
		{
			name: "undefined type",
			cred: &Credential{
				Type: CredentialTypeUndefined,
			},
			wantErr: true,
		},
		{
			name: "unknown type value",
			cred: &Credential{
				Type: CredentialType(999),
			},
			wantErr: true,
		},
		{
			name: "password with whitespace only user",
			cred: &Credential{
				Type:   CredentialTypePassword,
				User:   "   ",
				Secret: "pass",
			},
			wantErr: false, // Current implementation allows whitespace-only strings
		},
		{
			name: "bearer with very long token",
			cred: &Credential{
				Type:   CredentialTypeBearer,
				Secret: string(make([]byte, 10000)),
			},
			wantErr: false, // Should still be valid
		},
		{
			name: "json with invalid json data",
			cred: &Credential{
				Type:     CredentialTypeJSON,
				JSONData: "not-json",
			},
			wantErr: false, // Validation doesn't parse JSON, just checks non-empty
		},
		{
			name: "pkcs12 with empty data",
			cred: &Credential{
				Type:       CredentialTypePKCS12,
				PKCS12Data: []byte{},
			},
			wantErr: true,
		},
		{
			name: "pkcs12 with data",
			cred: &Credential{
				Type:       CredentialTypePKCS12,
				PKCS12Data: []byte{0x01, 0x02, 0x03},
			},
			wantErr: false,
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
