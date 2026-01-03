package runtime

import (
	"fmt"
	"net/url"
	"time"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/config"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/health"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/lb"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/metrics"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/router"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/short"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
	"go.uber.org/zap"
)

type Runtime struct {
	Router       *router.Router
	Groups       map[string]*GroupRuntime
	DefaultGroup string
	Auth         config.GatewayAuthConfig
	Metrics      config.MetricsConfig
	Pprof        config.PprofConfig
	Short        *short.Runtime
	Failover     config.FailoverConfig
	Server       config.ServerConfig

	stop chan struct{}
}

type GroupRuntime struct {
	Name           string
	Backend        string
	StripPrefix    string
	Priority       int
	Allow          []string
	Deny           []string
	Picker         lb.Picker
	Upstreams      []*upstream.State
	FallbackGroups []string
	Health         config.ActiveHealthConfig
	PassiveEject   config.PassiveEjectConfig
}

func Build(cfg *config.Config, m *metrics.Metrics, logger *zap.Logger) (*Runtime, error) {
	groups := make(map[string]*GroupRuntime, len(cfg.Groups))
	routes := make([]router.Route, 0, len(cfg.Groups))

	for idx, groupCfg := range cfg.Groups {
		upstates := make([]*upstream.State, 0, len(groupCfg.Upstreams))
		for _, up := range groupCfg.Upstreams {
			parsed, err := url.Parse(up.URL)
			if err != nil {
				return nil, fmt.Errorf("parse upstream url %s: %w", up.URL, err)
			}
			state := upstream.NewState(parsed, up.Weight, up.AccessKey)
			upstates = append(upstates, state)
			if m != nil {
				m.UpstreamHealth.WithLabelValues(groupCfg.Name, state.HostLabel).Set(1)
			}
		}

		var picker lb.Picker
		switch groupCfg.LB.Policy {
		case "hash":
			picker = lb.NewHashPicker(upstates)
		default:
			picker = lb.NewWRRPicker(upstates)
		}

		gr := &GroupRuntime{
			Name:           groupCfg.Name,
			Backend:        groupCfg.Backend,
			StripPrefix:    groupCfg.StripPrefix,
			Priority:       groupCfg.Priority,
			Allow:          groupCfg.Allow,
			Deny:           groupCfg.Deny,
			Picker:         picker,
			Upstreams:      upstates,
			FallbackGroups: groupCfg.FallbackGroups,
			Health:         groupCfg.Health.Active,
			PassiveEject:   cfg.Failover.PassiveEject,
		}
		groups[groupCfg.Name] = gr
		routes = append(routes, router.Route{
			Name:     groupCfg.Name,
			Allow:    groupCfg.Allow,
			Deny:     groupCfg.Deny,
			Priority: groupCfg.Priority,
			Order:    idx,
		})
	}

	shortEntries := make(map[string]string, len(cfg.Short.Entries))
	for _, entry := range cfg.Short.Entries {
		shortEntries[entry.Name] = entry.Target
	}

	rt := &Runtime{
		Router:       router.New(routes, cfg.Routing.DefaultGroup),
		Groups:       groups,
		DefaultGroup: cfg.Routing.DefaultGroup,
		Auth:         cfg.GatewayAuth,
		Metrics:      cfg.Metrics,
		Pprof:        cfg.Pprof,
		Short:        short.NewRuntime(cfg.Short.Enabled, cfg.Short.Path, shortEntries),
		Failover:     cfg.Failover,
		Server:       cfg.Server,
		stop:         make(chan struct{}),
	}

	for _, group := range rt.Groups {
		if group.Health.Enabled {
			health.Start(group.Name, group.Upstreams, group.Health, rt.stop, m, logger)
		}
	}

	return rt, nil
}

func (r *Runtime) Stop() {
	select {
	case <-r.stop:
		return
	default:
		close(r.stop)
	}
}

func (r *Runtime) AvailableUpstreams(groupName string) int {
	group := r.Groups[groupName]
	if group == nil {
		return 0
	}
	now := time.Now()
	count := 0
	for _, up := range group.Upstreams {
		if up.Available(now) {
			count++
		}
	}
	return count
}
