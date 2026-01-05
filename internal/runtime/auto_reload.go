package runtime

import (
	"time"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/config"
	"go.uber.org/zap"
)

func (m *Manager) autoReloadLoop() {
	for {
		rt := m.Get()
		enabled := false
		interval := 5 * time.Second
		if rt != nil {
			enabled = rt.Reload.Auto.Enabled
			if enabled {
				interval = timeDurationMS(rt.Reload.Auto.IntervalMS)
				if interval <= 0 {
					interval = 30 * time.Second
				}
			}
		}

		timer := time.NewTimer(interval)
		select {
		case <-timer.C:
		case <-m.autoStop:
			timer.Stop()
			return
		}

		if !enabled {
			continue
		}

		fileHash, err := config.FileHash(m.cfgPath)
		if err != nil {
			if m.logger != nil {
				m.logger.Warn("auto reload hash failed", zap.Error(err))
			}
			continue
		}
		storedHash, err := m.getHash()
		if err != nil {
			if m.logger != nil {
				m.logger.Warn("auto reload hash read failed", zap.Error(err))
			}
			continue
		}
		if fileHash == "" || fileHash == storedHash {
			continue
		}
		if err := m.reloadWithHash(fileHash); err != nil && m.logger != nil {
			m.logger.Warn("auto reload failed", zap.Error(err))
		}
	}
}

func timeDurationMS(ms int) time.Duration {
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}
