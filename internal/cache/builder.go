package cache

import (
	"strings"

	"go.uber.org/zap"
)

func NewStore(opts Options, logger *zap.Logger) Store {
	provider := strings.ToLower(strings.TrimSpace(opts.Provider))
	switch provider {
	case "redis":
		store, err := NewRedisStore(opts.Redis, opts.MaxItemBytes)
		if err == nil {
			return store
		}
		if logger != nil {
			logger.Error("redis cache unavailable, falling back to memory", zap.Error(err))
		}
		return NewMemoryStore(opts)
	case "memory", "":
		return NewMemoryStore(opts)
	default:
		if logger != nil {
			logger.Error("invalid cache provider, falling back to memory", zap.String("provider", opts.Provider))
		}
		return NewMemoryStore(opts)
	}
}
