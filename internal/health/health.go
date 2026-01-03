package health

import (
	"context"
	"net"
	"time"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/config"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/metrics"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func Start(groupName string, upstreams []*upstream.State, cfg config.ActiveHealthConfig, stop <-chan struct{}, m *metrics.Metrics, logger *zap.Logger) {
	interval := time.Duration(cfg.IntervalMS) * time.Millisecond
	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	path := cfg.Path
	if path == "" {
		path = "/healthz"
	}

	client := &fasthttp.Client{
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				for _, up := range upstreams {
					probe(groupName, up, path, cfg.Retries, client, timeout, m, logger)
				}
			}
		}
	}()
}

func probe(groupName string, up *upstream.State, path string, retries int, client *fasthttp.Client, timeout time.Duration, m *metrics.Metrics, logger *zap.Logger) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(up.URL.String() + path)
	req.Header.SetMethod(fasthttp.MethodGet)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := doWithContext(ctx, client, req, resp, timeout)
	if err != nil || resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		becameUnhealthy := up.HealthFail(retries)
		if becameUnhealthy {
			if m != nil {
				m.UpstreamHealth.WithLabelValues(groupName, up.HostLabel).Set(0)
			}
			if logger != nil {
				logger.Warn("health change", zap.String("group", groupName), zap.String("upstream", up.HostLabel), zap.String("status", "unhealthy"))
			}
		}
		return
	}
	if up.HealthSuccess() {
		if m != nil {
			m.UpstreamHealth.WithLabelValues(groupName, up.HostLabel).Set(1)
		}
		if logger != nil {
			logger.Info("health change", zap.String("group", groupName), zap.String("upstream", up.HostLabel), zap.String("status", "healthy"))
		}
	}
}

func doWithContext(ctx context.Context, client *fasthttp.Client, req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error {
	deadline, ok := ctx.Deadline()
	if ok {
		return client.DoDeadline(req, resp, deadline)
	}
	return client.DoTimeout(req, resp, timeout)
}

func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	if nerr, ok := err.(net.Error); ok {
		return nerr.Timeout()
	}
	return false
}
