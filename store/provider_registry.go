package store

import (
	"fmt"

	"github.com/nilotpaul/go-downloader/types"
)

type ProviderRegistry struct {
	Providers map[string]types.OAuthProvider
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		Providers: make(map[string]types.OAuthProvider),
	}
}

func (r *ProviderRegistry) Register(providerName string, p types.OAuthProvider) {
	r.Providers[providerName] = p
}

func (r *ProviderRegistry) GetProvider(providerName string) (types.OAuthProvider, error) {
	p, exists := r.Providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider doesn't exists")
	}

	return p, nil
}
