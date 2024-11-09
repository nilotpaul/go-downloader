package store

import (
	"database/sql"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/setting"
	"github.com/nilotpaul/go-downloader/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

// Mock Provider
type MockProvider struct{}

func (p *MockProvider) Authenticate(string) error { return nil }

func (p *MockProvider) GetAccessToken() string { return "" }

func (p *MockProvider) GetRefreshToken() string { return "" }

func (p *MockProvider) RefreshToken(*fiber.Ctx, string, bool) (*oauth2.Token, error) {
	return nil, nil
}

func (p *MockProvider) IsTokenValid() bool { return true }

func (p *MockProvider) GetAuthURL(string) string { return "" }

func (p *MockProvider) CreateOrUpdateAccount() (string, error) { return "", nil }

func (p *MockProvider) CreateSession(*fiber.Ctx, string) error { return nil }

func (p *MockProvider) UpdateTokens(*types.GoogleAccount) error { return nil }

func TestNewProviderRegistry(t *testing.T) {
	r := NewProviderRegistry()
	assert.NotNil(t, r)
	assert.Empty(t, r.Providers)
}

func TestProviderRegistry_RegisterAndGetProvider(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := NewProviderRegistry()

	mockProvider := &MockProvider{}

	// Registering an OAuth Provider
	r.Register("mock_provider", mockProvider)
	a.Equal(len(r.Providers), 1)

	// Getting a non-existent Provider
	p, err := r.GetProvider("non-existent")
	a.Error(err)
	a.Equal("provider not found", err.Error())
	a.Nil(p)

	// Getting the Mock Provider
	mp, err := r.GetProvider("mock_provider")
	a.NoError(err)
	a.Equal(mockProvider, mp)

	// Adding another Mock Provider
	r.Register("mock_provider_2", mockProvider)
	newMp, err := r.GetProvider("mock_provider")
	a.NoError(err)
	a.Equal(mockProvider, newMp)
	a.Equal(len(r.Providers), 2)
}

func TestInitStore(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	// Mock the database connection.
	var db *sql.DB

	// Testing with no environment variables which means no providers will be assigned.
	r := InitStore(config.EnvConfig{}, db)
	gp, err := r.GetProvider(setting.GoogleProvider)
	a.Error(err)
	a.Equal("provider not found", err.Error())
	a.Nil(gp)
	a.Equal(len(r.Providers), 0)

	// Testing with Google OAuth Provider.
	env := config.EnvConfig{
		GoogleOAuthEnvConfig: config.GoogleOAuthEnvConfig{
			GoogleClientID:     "mock_client_id",
			GoogleClientSecret: "mock_client_secret",
		},
	}

	r = InitStore(env, db)
	gp, err = r.GetProvider(setting.GoogleProvider)
	a.NoError(err)
	a.NotNil(gp)
	a.Equal(len(r.Providers), 1)
}
