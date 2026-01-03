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
		}, []string{"method", "group", "status"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "rsshub_gateway_request_duration_seconds",
			Help:    "Gateway request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"group"}),
		UpstreamRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rsshub_gateway_upstream_requests_total",
			Help: "Total number of upstream requests.",
		}, []string{"group", "upstream", "status"}),
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
