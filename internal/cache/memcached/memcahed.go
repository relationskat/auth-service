package memcached

import (
	"errors"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"go.uber.org/zap"
)

var ErrNotFound = errors.New("token not found or expired")

type Memcached struct {
	log   *zap.Logger
	cache *memcache.Client
}

func New(log *zap.Logger, cache *memcache.Client) *Memcached {
	mem := &Memcached{
		log:   log.Named("memcache"),
		cache: cache,
	}

	return mem
}

func (m *Memcached) SetToken(token, email string, ttl time.Duration) error {
	const op = "memcached.SetToken"

	seconds := int32(ttl.Seconds())
	if seconds <= 0 {
		return fmt.Errorf("%s: ttl must be > 0", op)
	}

	err := m.cache.Set(&memcache.Item{
		Key:        token,
		Value:      []byte(email),
		Expiration: seconds,
	})
	if err != nil {
		m.log.Error("failed to set token", zap.String("token", token), zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (m *Memcached) GetToken(token string) (string, error) {
	const op = "memcached.GetToken"

	item, err := m.cache.Get(token)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return "", ErrNotFound
	}
	if err != nil {
		m.log.Error("failed to get token", zap.String("token", token), zap.Error(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return string(item.Value), nil
}

func (m *Memcached) DeleteToken(token string) error {
	const op = "memcached.DeleteToken"

	err := m.cache.Delete(token)
	if err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		m.log.Error("failed to delete token", zap.String("token", token), zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
