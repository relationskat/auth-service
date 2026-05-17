package provider

import (
	"crypto/rsa"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/relationskat/auth-service/internal/config"
)

type Provider struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	cfg        *config.Config
}

func New(cfg *config.Config) (*Provider, error) {
	publicKeyPEM, err := os.ReadFile(cfg.Provider.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	privateKeyPEM, err := os.ReadFile(cfg.Provider.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return &Provider{
		publicKey:  publicKey,
		privateKey: privateKey,
		cfg:        cfg,
	}, nil
}
