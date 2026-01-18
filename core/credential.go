// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package core

import (
	"fmt"
	"os"
	"strings"
)

// CredentialType represents the type of credential being used
type CredentialType int

const (
	// CredentialTypeUndefined represents an undefined credential type
	CredentialTypeUndefined CredentialType = iota
	// CredentialTypePassword represents username/password authentication
	CredentialTypePassword
	// CredentialTypePrivateKey represents private key authentication
	CredentialTypePrivateKey
	// CredentialTypeSSHAgent represents SSH agent authentication
	CredentialTypeSSHAgent
	// CredentialTypeBearer represents bearer token authentication
	CredentialTypeBearer
	// CredentialTypeCredentialsQuery represents credential query authentication
	CredentialTypeCredentialsQuery
	// CredentialTypeJSON represents JSON-based authentication
	CredentialTypeJSON
	// CredentialTypeAWSEC2InstanceConnect represents AWS EC2 Instance Connect
	CredentialTypeAWSEC2InstanceConnect
	// CredentialTypeAWSEC2SSMSession represents AWS EC2 SSM Session
	CredentialTypeAWSEC2SSMSession
	// CredentialTypePKCS12 represents PKCS12 certificate authentication
	CredentialTypePKCS12
	// CredentialTypeEnv represents environment variable based authentication
	CredentialTypeEnv
)

// String returns the string representation of the credential type
func (ct CredentialType) String() string {
	switch ct { //nolint:exhaustive // CredentialTypeUndefined handled in default
	case CredentialTypePassword:
		return "password"
	case CredentialTypePrivateKey:
		return "private_key"
	case CredentialTypeSSHAgent:
		return "ssh_agent"
	case CredentialTypeBearer:
		return "bearer"
	case CredentialTypeCredentialsQuery:
		return "credentials_query"
	case CredentialTypeJSON:
		return "json"
	case CredentialTypeAWSEC2InstanceConnect:
		return "aws_ec2_instance_connect"
	case CredentialTypeAWSEC2SSMSession:
		return "aws_ec2_ssm_session"
	case CredentialTypePKCS12:
		return "pkcs12"
	case CredentialTypeEnv:
		return "env"
	default:
		return "undefined"
	}
}

// ParseCredentialType parses a string into a CredentialType
func ParseCredentialType(s string) CredentialType {
	switch strings.ToLower(s) {
	case "password":
		return CredentialTypePassword
	case "private_key":
		return CredentialTypePrivateKey
	case "ssh_agent":
		return CredentialTypeSSHAgent
	case "bearer":
		return CredentialTypeBearer
	case "credentials_query":
		return CredentialTypeCredentialsQuery
	case "json":
		return CredentialTypeJSON
	case "aws_ec2_instance_connect":
		return CredentialTypeAWSEC2InstanceConnect
	case "aws_ec2_ssm_session":
		return CredentialTypeAWSEC2SSMSession
	case "pkcs12":
		return CredentialTypePKCS12
	case "env":
		return CredentialTypeEnv
	default:
		return CredentialTypeUndefined
	}
}

// Credential represents authentication credentials for a provider
type Credential struct {
	// Type of credential
	Type CredentialType

	// Username for password-based authentication
	User string

	// Password/Secret for password-based or bearer authentication
	Secret string

	// PrivateKey contains the private key data (for private_key type)
	PrivateKey []byte

	// PrivateKeyPath contains the path to a private key file
	PrivateKeyPath string

	// PrivateKeyPassword is the password for encrypted private keys
	PrivateKeyPassword string

	// EnvVarName is the name of the environment variable to read credentials from
	EnvVarName string

	// JSONData contains JSON-formatted credential data
	JSONData string

	// PKCS12Data contains PKCS12 certificate data
	PKCS12Data []byte

	// PKCS12Password is the password for PKCS12 certificates
	PKCS12Password string
}

// Validate checks if the credential is valid based on its type
func (c *Credential) Validate() error {
	switch c.Type { //nolint:exhaustive // Unhandled types fall through to default error
	case CredentialTypePassword:
		if c.User == "" {
			return fmt.Errorf("username is required for password authentication")
		}
		if c.Secret == "" {
			return fmt.Errorf("password is required for password authentication")
		}

	case CredentialTypePrivateKey:
		if len(c.PrivateKey) == 0 && c.PrivateKeyPath == "" {
			return fmt.Errorf("private key or private key path is required for private key authentication")
		}

	case CredentialTypeBearer:
		if c.Secret == "" {
			return fmt.Errorf("token/secret is required for bearer authentication")
		}

	case CredentialTypeEnv:
		if c.EnvVarName == "" {
			return fmt.Errorf("environment variable name is required for env authentication")
		}

	case CredentialTypeJSON:
		if c.JSONData == "" {
			return fmt.Errorf("JSON data is required for JSON authentication")
		}

	case CredentialTypePKCS12:
		if len(c.PKCS12Data) == 0 {
			return fmt.Errorf("PKCS12 data is required for PKCS12 authentication")
		}

	case CredentialTypeSSHAgent:
		// SSH agent doesn't require additional validation
		return nil

	case CredentialTypeUndefined:
		return fmt.Errorf("credential type is undefined")

	default:
		return fmt.Errorf("unsupported credential type: %s", c.Type)
	}

	return nil
}

// ResolveSecret returns the actual secret value, resolving environment variables if needed
func (c *Credential) ResolveSecret() (string, error) {
	switch c.Type { //nolint:exhaustive // Only certain credential types support ResolveSecret
	case CredentialTypeEnv:
		if c.EnvVarName == "" {
			return "", fmt.Errorf("environment variable name not set")
		}
		secret := os.Getenv(c.EnvVarName)
		if secret == "" {
			return "", fmt.Errorf("environment variable %s is empty or not set", c.EnvVarName)
		}
		return secret, nil

	case CredentialTypeBearer, CredentialTypePassword:
		return c.Secret, nil

	default:
		return "", fmt.Errorf("ResolveSecret is not applicable for credential type: %s", c.Type)
	}
}

// ParseCredentialFromConfig creates a Credential from a configuration map
// Expected keys: credential_type, user, secret, private_key_path, env_var, etc.
func ParseCredentialFromConfig(config map[string]string) (*Credential, error) {
	credTypeStr := config["credential_type"]
	if credTypeStr == "" {
		// Default: check for common keys
		if token := config["token"]; token != "" {
			// Legacy support: if "token" is present, use bearer auth
			return &Credential{
				Type:   CredentialTypeBearer,
				Secret: token,
			}, nil
		}
		return nil, fmt.Errorf("credential_type not specified in config")
	}

	credType := ParseCredentialType(credTypeStr)
	cred := &Credential{
		Type: credType,
	}

	switch credType { //nolint:exhaustive // Unhandled types fall through to default error
	case CredentialTypePassword:
		cred.User = config["user"]
		cred.Secret = config["secret"]

	case CredentialTypePrivateKey:
		cred.PrivateKeyPath = config["private_key_path"]
		cred.PrivateKeyPassword = config["private_key_password"]
		// PrivateKey data would be loaded separately if needed

	case CredentialTypeBearer:
		cred.Secret = config["secret"]
		if cred.Secret == "" {
			cred.Secret = config["token"] // Alternative key
		}

	case CredentialTypeEnv:
		cred.EnvVarName = config["env_var"]

	case CredentialTypeJSON:
		cred.JSONData = config["json_data"]

	case CredentialTypePKCS12:
		cred.PKCS12Password = config["pkcs12_password"]
		// PKCS12 data would be loaded separately

	case CredentialTypeSSHAgent:
		// No additional config needed

	default:
		return nil, fmt.Errorf("unsupported credential type: %s", credTypeStr)
	}

	if err := cred.Validate(); err != nil {
		return nil, fmt.Errorf("credential validation failed: %w", err)
	}

	return cred, nil
}
