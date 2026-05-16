package provider

import (
	"os"

	"github.com/relationskat/auth-service/internal/config"
)

type Provider struct {
	publicKey  []byte
	privateKey []byte
	cfg        *config.Config
}

func New(cfg *config.Config) (*Provider, error) {
	provider := &Provider{
		cfg: cfg,
	}

	publicKeyPath := provider.cfg.Provider.PublicKeyPath

	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	privateKeyPath := provider.cfg.Provider.PrivateKeyPath

	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	provider.privateKey = privateKey
	provider.publicKey = publicKey

	return provider, nil
}
