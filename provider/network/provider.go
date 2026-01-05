package network

import (
	"context"

	"github.com/kopexa-grc/kspec/core"
)

type NetworkProvider struct{}

func NewNetworkProvider() *NetworkProvider {
	return &NetworkProvider{}
}

func (p *NetworkProvider) Name() string {
	return "network"
}

func (p *NetworkProvider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Network provider currently stateless, potentially could take proxy config etc
	return &NetworkConnection{}, nil
}

type NetworkConnection struct{}

func (c *NetworkConnection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&CertificatesResource{},
		&DNSResource{},
		&HTTPResource{},
		&TLSResource{},
	}
}
