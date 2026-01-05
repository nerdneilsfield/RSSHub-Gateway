package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server      ServerConfig      `yaml:"server"`
	GatewayAuth GatewayAuthConfig `yaml:"gateway_auth"`
	Metrics     MetricsConfig     `yaml:"metrics"`
	Pprof       PprofConfig       `yaml:"pprof"`
	Cache       CacheConfig       `yaml:"cache"`
	Reload      ReloadConfig      `yaml:"reload"`
	Short       ShortConfig       `yaml:"short"`
	Routing     RoutingConfig     `yaml:"routing"`
	Failover    FailoverConfig    `yaml:"failover"`
	Groups      []GroupConfig     `yaml:"groups"`
}

type ServerConfig struct {
	Listen    string `yaml:"listen"`
	TimeoutMS int    `yaml:"timeout_ms"`
}

type GatewayAuthConfig struct {
	Enabled     bool     `yaml:"enabled"`
	AccessKey   string   `yaml:"access_key"`
	AcceptKey   bool     `yaml:"accept_key"`
	AcceptCode  bool     `yaml:"accept_code"`
	BypassPaths []string `yaml:"bypass_paths"`
}

type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Path      string `yaml:"path"`
	AccessKey string `yaml:"accesskey"`
}

type PprofConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Path      string `yaml:"path"`
	AccessKey string `yaml:"accesskey"`
}

type CacheConfig struct {
	Enabled       bool        `yaml:"enabled"`
	Provider      string      `yaml:"provider"`
	TTLMS         int         `yaml:"ttl_ms"`
	MaxItemBytes  int         `yaml:"max_item_bytes"`
	MaxTotalBytes int         `yaml:"max_total_bytes"`
	Redis         RedisConfig `yaml:"redis"`
}

type RedisConfig struct {
	Addr           string `yaml:"addr"`
	Password       string `yaml:"password"`
	DB             int    `yaml:"db"`
	DialTimeoutMS  int    `yaml:"dial_timeout_ms"`
	ReadTimeoutMS  int    `yaml:"read_timeout_ms"`
	WriteTimeoutMS int    `yaml:"write_timeout_ms"`
	KeyPrefix      string `yaml:"key_prefix"`
}

type ReloadConfig struct {
	Auto AutoReloadConfig `yaml:"auto"`
}

type AutoReloadConfig struct {
	Enabled    bool `yaml:"enabled"`
	IntervalMS int  `yaml:"interval_ms"`
}

type ShortConfig struct {
	Enabled bool         `yaml:"enabled"`
	Path    string       `yaml:"path"`
	Entries []ShortEntry `yaml:"entries"`
}

type ShortEntry struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
}

type RoutingConfig struct {
	DefaultGroup string `yaml:"default_group"`
}

type FailoverConfig struct {
	Retry        RetryConfig        `yaml:"retry"`
	PassiveEject PassiveEjectConfig `yaml:"passive_eject"`
}

type RetryConfig struct {
	Enabled    bool `yaml:"enabled"`
	MaxRetries int  `yaml:"max_retries"`
}

type PassiveEjectConfig struct {
	Enabled       bool `yaml:"enabled"`
	FailThreshold int  `yaml:"fail_threshold"`
	BaseEjectMS   int  `yaml:"base_eject_ms"`
	MaxEjectMS    int  `yaml:"max_eject_ms"`
}

type GroupConfig struct {
	Name           string           `yaml:"name"`
	Backend        string           `yaml:"backend"`
	StripPrefix    string           `yaml:"strip_prefix"`
	Priority       int              `yaml:"priority"`
	Allow          []string         `yaml:"allow"`
	Deny           []string         `yaml:"deny"`
	LB             LBConfig         `yaml:"lb"`
	FallbackGroups []string         `yaml:"fallback_groups"`
	Health         HealthConfig     `yaml:"health"`
	Upstreams      []UpstreamConfig `yaml:"upstreams"`
}

type LBConfig struct {
	Policy string `yaml:"policy"`
}

type HealthConfig struct {
	Active ActiveHealthConfig `yaml:"active"`
}

type ActiveHealthConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Path       string `yaml:"path"`
	IntervalMS int    `yaml:"interval_ms"`
	TimeoutMS  int    `yaml:"timeout_ms"`
	Retries    int    `yaml:"retries"`
}

type UpstreamConfig struct {
	URL       string `yaml:"url"`
	Weight    int    `yaml:"weight"`
	AccessKey string `yaml:"access_key"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	cfg.normalize()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Listen == "" {
		c.Server.Listen = ":8080"
	}
	if c.Server.TimeoutMS == 0 {
		c.Server.TimeoutMS = 8000
	}
	if c.Metrics.Path == "" {
		c.Metrics.Path = "/metrics"
	}
	if c.Pprof.Path == "" {
		c.Pprof.Path = "/debug/pprof"
	}
	if c.Short.Path == "" {
		c.Short.Path = "/short"
	}
	if c.Cache.TTLMS == 0 {
		c.Cache.TTLMS = 3600000
	}
	if c.Cache.MaxItemBytes == 0 {
		c.Cache.MaxItemBytes = 2 * 1024 * 1024
	}
	if c.Cache.MaxTotalBytes == 0 {
		c.Cache.MaxTotalBytes = 50 * 1024 * 1024
	}
	if c.Cache.Provider == "" && c.Cache.Enabled {
		c.Cache.Provider = "memory"
	}
	if c.Cache.Redis.DialTimeoutMS == 0 {
		c.Cache.Redis.DialTimeoutMS = 1000
	}
	if c.Cache.Redis.ReadTimeoutMS == 0 {
		c.Cache.Redis.ReadTimeoutMS = 1000
	}
	if c.Cache.Redis.WriteTimeoutMS == 0 {
		c.Cache.Redis.WriteTimeoutMS = 1000
	}
	if c.Cache.Redis.KeyPrefix == "" {
		c.Cache.Redis.KeyPrefix = "rsshub_gateway"
	}
	if c.Reload.Auto.Enabled && c.Reload.Auto.IntervalMS == 0 {
		c.Reload.Auto.IntervalMS = 30000
	}
	if c.GatewayAuth.Enabled && !c.GatewayAuth.AcceptKey && !c.GatewayAuth.AcceptCode {
		c.GatewayAuth.AcceptKey = true
		c.GatewayAuth.AcceptCode = true
	}
	if c.Failover.Retry.MaxRetries == 0 {
		c.Failover.Retry.MaxRetries = 1
	}
	if c.Failover.PassiveEject.FailThreshold == 0 {
		c.Failover.PassiveEject.FailThreshold = 3
	}
	if c.Failover.PassiveEject.BaseEjectMS == 0 {
		c.Failover.PassiveEject.BaseEjectMS = 10000
	}
	if c.Failover.PassiveEject.MaxEjectMS == 0 {
		c.Failover.PassiveEject.MaxEjectMS = 60000
	}
	for gi := range c.Groups {
		if c.Groups[gi].Backend == "" {
			c.Groups[gi].Backend = "rsshub"
		}
		if c.Groups[gi].LB.Policy == "" {
			c.Groups[gi].LB.Policy = "wrr"
		}
		if c.Groups[gi].Health.Active.Path == "" {
			c.Groups[gi].Health.Active.Path = "/healthz"
		}
		if c.Groups[gi].Health.Active.IntervalMS == 0 {
			c.Groups[gi].Health.Active.IntervalMS = 30000
		}
		if c.Groups[gi].Health.Active.TimeoutMS == 0 {
			c.Groups[gi].Health.Active.TimeoutMS = 10000
		}
		if c.Groups[gi].Health.Active.Retries == 0 {
			c.Groups[gi].Health.Active.Retries = 3
		}
		for ui := range c.Groups[gi].Upstreams {
			if c.Groups[gi].Upstreams[ui].Weight == 0 {
				c.Groups[gi].Upstreams[ui].Weight = 1
			}
		}
	}
}

func (c *Config) normalize() {
	c.Metrics.Path = strings.TrimSpace(c.Metrics.Path)
	c.Pprof.Path = strings.TrimSpace(c.Pprof.Path)
	c.Short.Path = normalizePathPrefix(c.Short.Path)
	c.GatewayAuth.BypassPaths = normalizePaths(c.GatewayAuth.BypassPaths)
	c.Cache.Provider = strings.ToLower(strings.TrimSpace(c.Cache.Provider))
	c.Cache.Redis.Addr = strings.TrimSpace(c.Cache.Redis.Addr)
	c.Cache.Redis.Password = strings.TrimSpace(c.Cache.Redis.Password)
	c.Cache.Redis.KeyPrefix = strings.TrimSpace(c.Cache.Redis.KeyPrefix)
	for i := range c.Short.Entries {
		c.Short.Entries[i].Name = strings.TrimSpace(c.Short.Entries[i].Name)
		c.Short.Entries[i].Target = strings.TrimSpace(c.Short.Entries[i].Target)
	}
	for gi := range c.Groups {
		c.Groups[gi].Backend = strings.ToLower(strings.TrimSpace(c.Groups[gi].Backend))
		if c.Groups[gi].Backend == "" {
			c.Groups[gi].Backend = "rsshub"
		}
		c.Groups[gi].StripPrefix = strings.TrimSpace(c.Groups[gi].StripPrefix)
		if c.Groups[gi].StripPrefix != "" && !strings.HasPrefix(c.Groups[gi].StripPrefix, "/") {
			c.Groups[gi].StripPrefix = "/" + c.Groups[gi].StripPrefix
		}
		c.Groups[gi].Allow = normalizePrefixes(c.Groups[gi].Allow)
		c.Groups[gi].Deny = normalizePrefixes(c.Groups[gi].Deny)
	}
}

func normalizePathPrefix(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	if value != "/" {
		value = strings.TrimRight(value, "/")
	}
	return value
}

func normalizePrefixes(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !strings.HasPrefix(v, "/") {
			v = "/" + v
		}
		out = append(out, v)
	}
	return out
}

func normalizePaths(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		v = normalizePathPrefix(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func (c *Config) validate() error {
	if c.GatewayAuth.Enabled {
		if c.GatewayAuth.AccessKey == "" {
			return fmt.Errorf("gateway_auth.access_key is required")
		}
		if !c.GatewayAuth.AcceptKey && !c.GatewayAuth.AcceptCode {
			return fmt.Errorf("gateway_auth requires accept_key or accept_code")
		}
	}
	if c.Metrics.Enabled && c.Metrics.AccessKey == "" {
		return fmt.Errorf("metrics.accesskey is required")
	}
	if c.Metrics.Enabled {
		if c.Metrics.Path == "" || !strings.HasPrefix(c.Metrics.Path, "/") {
			return fmt.Errorf("metrics.path must start with /")
		}
	}
	if c.Pprof.Enabled {
		if c.Pprof.AccessKey == "" {
			return fmt.Errorf("pprof.accesskey is required")
		}
		if c.Pprof.Path == "" || !strings.HasPrefix(c.Pprof.Path, "/") {
			return fmt.Errorf("pprof.path must start with /")
		}
	}
	if c.Cache.Enabled {
		switch c.Cache.Provider {
		case "memory", "redis":
		default:
			return fmt.Errorf("cache.provider must be memory or redis")
		}
		if c.Cache.TTLMS <= 0 {
			return fmt.Errorf("cache.ttl_ms must be > 0")
		}
		if c.Cache.MaxItemBytes <= 0 {
			return fmt.Errorf("cache.max_item_bytes must be > 0")
		}
		if c.Cache.MaxTotalBytes <= 0 {
			return fmt.Errorf("cache.max_total_bytes must be > 0")
		}
		if c.Cache.Provider == "redis" {
			if c.Cache.Redis.Addr == "" {
				return fmt.Errorf("cache.redis.addr is required")
			}
			if c.Cache.Redis.DialTimeoutMS <= 0 || c.Cache.Redis.ReadTimeoutMS <= 0 || c.Cache.Redis.WriteTimeoutMS <= 0 {
				return fmt.Errorf("cache.redis timeouts must be > 0")
			}
		}
	}
	if c.Reload.Auto.Enabled && c.Reload.Auto.IntervalMS <= 0 {
		return fmt.Errorf("reload.auto.interval_ms must be > 0")
	}
	if c.Short.Enabled {
		if c.Short.Path == "" || !strings.HasPrefix(c.Short.Path, "/") {
			return fmt.Errorf("short.path must start with /")
		}
		names := make(map[string]struct{}, len(c.Short.Entries))
		for _, entry := range c.Short.Entries {
			if entry.Name == "" {
				return fmt.Errorf("short entry name is required")
			}
			if _, exists := names[entry.Name]; exists {
				return fmt.Errorf("short entry name must be unique: %s", entry.Name)
			}
			names[entry.Name] = struct{}{}
			if !isShortTarget(entry.Target) {
				return fmt.Errorf("short entry %s has invalid target: %s", entry.Name, entry.Target)
			}
		}
	}
	if c.Routing.DefaultGroup == "" {
		return fmt.Errorf("routing.default_group is required")
	}
	if c.Server.TimeoutMS <= 0 {
		return fmt.Errorf("server.timeout_ms must be > 0")
	}
	if c.Failover.PassiveEject.BaseEjectMS > c.Failover.PassiveEject.MaxEjectMS {
		return fmt.Errorf("failover.passive_eject.base_eject_ms must be <= max_eject_ms")
	}

	groups := make(map[string]struct{}, len(c.Groups))
	for _, g := range c.Groups {
		if g.Name == "" {
			return fmt.Errorf("group name is required")
		}
		switch g.Backend {
		case "rsshub", "upvote":
		default:
			return fmt.Errorf("group %s has invalid backend: %s", g.Name, g.Backend)
		}
		if g.StripPrefix != "" && !strings.HasPrefix(g.StripPrefix, "/") {
			return fmt.Errorf("group %s strip_prefix must start with /", g.Name)
		}
		if _, exists := groups[g.Name]; exists {
			return fmt.Errorf("duplicate group name: %s", g.Name)
		}
		groups[g.Name] = struct{}{}
		switch g.LB.Policy {
		case "wrr", "hash":
		default:
			return fmt.Errorf("group %s has invalid lb policy: %s", g.Name, g.LB.Policy)
		}
		if len(g.Upstreams) == 0 {
			return fmt.Errorf("group %s must have at least one upstream", g.Name)
		}
		if g.Health.Active.Enabled {
			if g.Health.Active.IntervalMS <= 0 || g.Health.Active.TimeoutMS <= 0 || g.Health.Active.Retries < 1 {
				return fmt.Errorf("group %s has invalid health.active settings", g.Name)
			}
		}
		for _, up := range g.Upstreams {
			if up.URL == "" {
				return fmt.Errorf("group %s upstream url is required", g.Name)
			}
			parsed, err := url.Parse(up.URL)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return fmt.Errorf("group %s has invalid upstream url: %s", g.Name, up.URL)
			}
			if parsed.Scheme != "http" && parsed.Scheme != "https" {
				return fmt.Errorf("group %s upstream url scheme must be http/https: %s", g.Name, up.URL)
			}
			if up.Weight <= 0 {
				return fmt.Errorf("group %s upstream weight must be > 0", g.Name)
			}
			if g.Backend == "rsshub" && up.AccessKey == "" {
				return fmt.Errorf("group %s upstream access_key is required", g.Name)
			}
		}
	}
	for _, g := range c.Groups {
		for _, fb := range g.FallbackGroups {
			if fb == g.Name {
				return fmt.Errorf("group %s fallback_groups cannot include itself", g.Name)
			}
			if _, ok := groups[fb]; !ok {
				return fmt.Errorf("group %s fallback_groups references unknown group %s", g.Name, fb)
			}
		}
	}
	if _, ok := groups[c.Routing.DefaultGroup]; !ok {
		return fmt.Errorf("routing.default_group %s not found in groups", c.Routing.DefaultGroup)
	}
	return nil
}

func isShortTarget(target string) bool {
	if target == "" {
		return false
	}
	if strings.HasPrefix(target, "https://") {
		return true
	}
	return target == "/rsshub" || strings.HasPrefix(target, "/rsshub/") ||
		target == "/upvote" || strings.HasPrefix(target, "/upvote/")
}
