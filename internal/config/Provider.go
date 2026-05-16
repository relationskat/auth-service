package config

type Provider struct {
	PublicKeyPath   string `yaml:"public_key_path"`
	PrivateKeyPath  string `yaml:"private_key_path"`
	RefreshTokenTTL int    `yaml:"refresh_token_ttl"`
	AccessTokenTTL  int    `yaml:"access_token_ttl"`
}
