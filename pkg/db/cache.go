package db

import (
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/relationskat/auth-service/internal/config"
)

func NewCache(cfg *config.Config) (*memcache.Client, error) {

	address := fmt.Sprintf("%s:%d", cfg.Cache.Host, cfg.Cache.Port)

	cache := memcache.New(address)

	return cache, nil
}
