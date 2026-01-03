package runtime

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/config"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/metrics"
	"go.uber.org/zap"
)

type Manager struct {
	cfgPath string
	metrics *metrics.Metrics
	logger  *zap.Logger

	current atomic.Value
	mu      sync.Mutex
}

func NewManager(cfgPath string, m *metrics.Metrics, logger *zap.Logger) (*Manager, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	rt, err := Build(cfg, m, logger)
	if err != nil {
		return nil, err
	}
	mgr := &Manager{cfgPath: cfgPath, metrics: m, logger: logger}
	mgr.current.Store(rt)
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
	m.mu.Lock()
	defer m.mu.Unlock()
	cfg, err := config.Load(m.cfgPath)
	if err != nil {
		m.recordReload("fail")
		return fmt.Errorf("reload config: %w", err)
	}
	rt, err := Build(cfg, m.metrics, m.logger)
	if err != nil {
		m.recordReload("fail")
		return fmt.Errorf("build runtime: %w", err)
	}
	old := m.current.Swap(rt)
	if old != nil {
		old.(*Runtime).Stop()
	}
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
