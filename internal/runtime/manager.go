package runtime

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/cache"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/config"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/metrics"
	"go.uber.org/zap"
)

type Manager struct {
	cfgPath string
	metrics *metrics.Metrics
	logger  *zap.Logger

	current atomic.Value
	store   atomic.Value
	mu      sync.Mutex
	hashMu  sync.Mutex
	hash    string

	autoStop chan struct{}
}

func NewManager(cfgPath string, m *metrics.Metrics, logger *zap.Logger) (*Manager, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	store := cache.NewStore(cache.Options{
		Enabled:       cfg.Cache.Enabled,
		Provider:      cfg.Cache.Provider,
		TTL:           timeDurationMS(cfg.Cache.TTLMS),
		MaxItemBytes:  cfg.Cache.MaxItemBytes,
		MaxTotalBytes: cfg.Cache.MaxTotalBytes,
		Redis: cache.RedisOptions{
			Addr:         cfg.Cache.Redis.Addr,
			Password:     cfg.Cache.Redis.Password,
			DB:           cfg.Cache.Redis.DB,
			DialTimeout:  timeDurationMS(cfg.Cache.Redis.DialTimeoutMS),
			ReadTimeout:  timeDurationMS(cfg.Cache.Redis.ReadTimeoutMS),
			WriteTimeout: timeDurationMS(cfg.Cache.Redis.WriteTimeoutMS),
			KeyPrefix:    cfg.Cache.Redis.KeyPrefix,
		},
	}, logger)
	rt, err := Build(cfg, m, logger, store)
	if err != nil {
		if store != nil {
			_ = store.Close()
		}
		return nil, err
	}
	mgr := &Manager{cfgPath: cfgPath, metrics: m, logger: logger}
	mgr.current.Store(rt)
	mgr.store.Store(store)
	mgr.autoStop = make(chan struct{})
	if hash, err := config.FileHash(cfgPath); err == nil {
		mgr.setHash(hash)
	} else if logger != nil {
		logger.Warn("config hash init failed", zap.Error(err))
	}
	go mgr.autoReloadLoop()
	return mgr, nil
}

func (m *Manager) Get() *Runtime {
	value := m.current.Load()
	if value == nil {
		return nil
	}
	return value.(*Runtime)
}

func (m *Manager) Reload() error {
	hash, err := config.FileHash(m.cfgPath)
	if err != nil {
		m.recordReload("fail")
		return fmt.Errorf("reload hash: %w", err)
	}
	return m.reloadWithHash(hash)
}

func (m *Manager) reloadWithHash(hash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cfg, err := config.Load(m.cfgPath)
	if err != nil {
		m.recordReload("fail")
		return fmt.Errorf("reload config: %w", err)
	}
	store := cache.NewStore(cache.Options{
		Enabled:       cfg.Cache.Enabled,
		Provider:      cfg.Cache.Provider,
		TTL:           timeDurationMS(cfg.Cache.TTLMS),
		MaxItemBytes:  cfg.Cache.MaxItemBytes,
		MaxTotalBytes: cfg.Cache.MaxTotalBytes,
		Redis: cache.RedisOptions{
			Addr:         cfg.Cache.Redis.Addr,
			Password:     cfg.Cache.Redis.Password,
			DB:           cfg.Cache.Redis.DB,
			DialTimeout:  timeDurationMS(cfg.Cache.Redis.DialTimeoutMS),
			ReadTimeout:  timeDurationMS(cfg.Cache.Redis.ReadTimeoutMS),
			WriteTimeout: timeDurationMS(cfg.Cache.Redis.WriteTimeoutMS),
			KeyPrefix:    cfg.Cache.Redis.KeyPrefix,
		},
	}, m.logger)
	rt, err := Build(cfg, m.metrics, m.logger, store)
	if err != nil {
		if store != nil {
			_ = store.Close()
		}
		m.recordReload("fail")
		return fmt.Errorf("build runtime: %w", err)
	}
	old := m.current.Swap(rt)
	if old != nil {
		old.(*Runtime).Stop()
	}
	m.store.Store(store)
	m.setHash(hash)
	m.recordReload("success")
	return nil
}

func (m *Manager) recordReload(result string) {
	if m.metrics != nil {
		m.metrics.ConfigReload.WithLabelValues(result).Inc()
	}
	if m.logger != nil {
		m.logger.Info("config reload", zap.String("result", result))
	}
}

func (m *Manager) getHash() (string, error) {
	store := m.getStore()
	if store != nil {
		value, ok, err := store.GetString(cache.ConfigHashKey)
		if err != nil {
			if m.logger != nil {
				m.logger.Warn("config hash read failed", zap.Error(err))
			}
		} else if ok {
			return value, nil
		}
	}
	m.hashMu.Lock()
	defer m.hashMu.Unlock()
	return m.hash, nil
}

func (m *Manager) setHash(value string) {
	store := m.getStore()
	if store != nil {
		if err := store.SetString(cache.ConfigHashKey, value, 0); err != nil && m.logger != nil {
			m.logger.Warn("config hash write failed", zap.Error(err))
		}
	}
	m.hashMu.Lock()
	m.hash = value
	m.hashMu.Unlock()
}

func (m *Manager) getStore() cache.Store {
	value := m.store.Load()
	if value == nil {
		return nil
	}
	return value.(cache.Store)
}
