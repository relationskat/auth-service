package cache

import "time"

type Cache interface {
	SetToken(token, email string, ttl time.Duration) error
	GetToken(token string) (string, error)
	DeleteToken(token string) error
}
