package metrics

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Registry *prometheus.Registry

	Requests         *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	UpstreamRequests *prometheus.CounterVec
	UpstreamSuccess  *prometheus.CounterVec
	UpstreamFailure  *prometheus.CounterVec
	RouteSuccess     *prometheus.CounterVec
	RouteFailure     *prometheus.CounterVec
	CacheHit         *prometheus.CounterVec
	CacheMiss        *prometheus.CounterVec
	UpstreamHealth   *prometheus.GaugeVec
	UpstreamEject    *prometheus.CounterVec
	RetryTotal       *prometheus.CounterVec
	FallbackTotal    *prometheus.CounterVec
	ConfigReload     *prometheus.CounterVec
}

func New() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		Registry: reg,
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_requests_total",
			Help: "Total number of gateway requests.",
		}, []string{"method", "group", "route_prefix", "status"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "rsshub_gateway_request_duration_seconds",
			Help:    "Gateway request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"group", "route_prefix"}),
		UpstreamRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_upstream_requests_total",
			Help: "Total number of upstream requests.",
		}, []string{"group", "upstream", "status"}),
		UpstreamSuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_upstream_success_total",
			Help: "Total number of successful upstream responses.",
		}, []string{"group", "upstream"}),
		UpstreamFailure: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_upstream_failure_total",
			Help: "Total number of failed upstream responses.",
		}, []string{"group", "upstream"}),
		RouteSuccess: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_route_success_total",
			Help: "Total number of successful route responses.",
		}, []string{"group", "route_prefix"}),
		RouteFailure: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_route_failure_total",
			Help: "Total number of failed route responses.",
		}, []string{"group", "route_prefix"}),
		CacheHit: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_cache_hit_total",
			Help: "Total number of cache hits.",
		}, []string{"provider"}),
		CacheMiss: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_cache_miss_total",
			Help: "Total number of cache misses.",
		}, []string{"provider"}),
		UpstreamHealth: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "rsshub_gateway_upstream_health",
			Help: "Upstream health status (1 healthy, 0 unhealthy).",
		}, []string{"group", "upstream"}),
		UpstreamEject: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_upstream_eject_total",
			Help: "Total number of upstream ejections.",
		}, []string{"group", "upstream"}),
		RetryTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_retry_total",
			Help: "Total number of retries.",
		}, []string{"group"}),
		FallbackTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_fallback_total",
			Help: "Total number of fallback attempts.",
		}, []string{"from", "to"}),
		ConfigReload: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_config_reload_total",
			Help: "Total number of config reloads.",
		}, []string{"result"}),
	}

	reg.MustRegister(
		m.Requests,
		m.RequestDuration,
		m.UpstreamRequests,
		m.UpstreamSuccess,
		m.UpstreamFailure,
		m.RouteSuccess,
		m.RouteFailure,
		m.CacheHit,
		m.CacheMiss,
		m.UpstreamHealth,
		m.UpstreamEject,
		m.RetryTotal,
		m.FallbackTotal,
		m.ConfigReload,
	)
	return m
}

func (m *Metrics) FiberHandler(accessKey string) fiber.Handler {
	handler := promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
	adapted := adaptor.HTTPHandler(handler)
	return func(c *fiber.Ctx) error {
		if accessKey != "" && c.Query("accesskey") != accessKey {
			return c.SendStatus(http.StatusForbidden)
		}
		return adapted(c)
	}
}
