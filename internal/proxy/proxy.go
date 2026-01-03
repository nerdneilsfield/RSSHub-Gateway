package proxy

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/auth"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/metrics"
	pprofhandler "github.com/nerdneilsfield/RSSHub-Gateway/internal/pprof"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/runtime"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Proxy struct {
	manager *runtime.Manager
	metrics *metrics.Metrics
	logger  *zap.Logger
	client  *fasthttp.Client
}

func New(manager *runtime.Manager, m *metrics.Metrics, logger *zap.Logger) *Proxy {
	return &Proxy{
		manager: manager,
		metrics: m,
		logger:  logger,
		client:  &fasthttp.Client{},
	}
}

func (p *Proxy) Serve(c *fiber.Ctx) error {
	start := time.Now()
	rt := p.manager.Get()
	if rt == nil {
		return c.SendStatus(http.StatusServiceUnavailable)
	}

	path := c.Path()
	method := string(c.Context().Method())
	if rt.Metrics.Enabled && method == fiber.MethodGet && path == rt.Metrics.Path {
		if p.metrics == nil {
			return c.SendStatus(http.StatusNotFound)
		}
		return p.metrics.FiberHandler(rt.Metrics.AccessKey)(c)
	}
	if rt.Pprof.Enabled && pprofhandler.MatchPath(path, rt.Pprof.Path) {
		return pprofhandler.Handle(c, rt.Pprof.Path, rt.Pprof.AccessKey)
	}

	args := c.Context().QueryArgs()
	if !auth.ValidateGateway(rt.Auth, path, args) {
		p.logAccess(start, method, path, "", "", http.StatusForbidden, 0, nil, "auth", errors.New("gateway auth failed"))
		return c.SendStatus(http.StatusForbidden)
	}

	selection := rt.Router.Select(path)
	groupName := selection.Group
	routePrefix := selection.RoutePrefix
	chain := buildChain(groupName, rt)
	fallbackChain := make([]string, 0, len(chain))
	fallbackChain = append(fallbackChain, groupName)
	finalGroup := groupName

	var lastErr error
	var lastErrType string
	var status int
	retries := 0

	for i, name := range chain {
		group := rt.Groups[name]
		if group == nil {
			continue
		}
		finalGroup = name
		if i > 0 && p.metrics != nil {
			p.metrics.FallbackTotal.WithLabelValues(chain[i-1], name).Inc()
		}
		if i > 0 {
			fallbackChain = append(fallbackChain, name)
		}

		attempts := 1
		if rt.Failover.Retry.Enabled && (method == fiber.MethodGet || method == fiber.MethodHead) {
			if rt.Failover.Retry.MaxRetries > 0 {
				attempts = 1 + rt.Failover.Retry.MaxRetries
			}
		}
		avoid := map[*upstream.State]struct{}{}
		for attempt := 0; attempt < attempts; attempt++ {
			up := group.Picker.Pick(path, time.Now(), avoid)
			if up == nil {
				break
			}
			if attempt > 0 {
				retries++
				if p.metrics != nil {
					p.metrics.RetryTotal.WithLabelValues(group.Name).Inc()
				}
			}
			resp, errType, err := p.forward(c, rt, group.Name, up, path)
			status = resp.status
			lastErrType = errType
			lastErr = err

			if err == nil {
				p.recordUpstreamMetrics(group.Name, up.HostLabel, resp.status)
				if resp.status >= 500 {
					p.recordFailure(group, up, errType, resp.status)
				}
				p.recordRequestMetrics(group.Name, routePrefix, method, resp.status, start)
				p.logAccess(start, method, path, group.Name, routePrefix, resp.status, retries, fallbackChain, errType, nil)
				return copyResponse(c, resp)
			}

			p.recordFailure(group, up, errType, resp.status)
			avoid[up] = struct{}{}
			if !shouldRetry(errType) {
				p.recordRequestMetrics(group.Name, routePrefix, method, resp.status, start)
				p.logAccess(start, method, path, group.Name, routePrefix, resp.status, retries, fallbackChain, errType, err)
				return c.SendStatus(resp.status)
			}
		}
	}

	if lastErrType == "timeout" {
		status = http.StatusGatewayTimeout
	} else {
		status = http.StatusBadGateway
	}
	p.recordRequestMetrics(finalGroup, routePrefix, method, status, start)
	p.logAccess(start, method, path, finalGroup, routePrefix, status, retries, fallbackChain, lastErrType, lastErr)
	return c.SendStatus(status)
}

type responseData struct {
	status  int
	headers [][2][]byte
	body    []byte
}

func (p *Proxy) forward(c *fiber.Ctx, rt *runtime.Runtime, groupName string, up *upstream.State, path string) (responseData, string, error) {
	var resp responseData
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	args := auth.InjectUpstreamCode(c.Context().QueryArgs(), path, up.AccessKey)
	uri := up.URL.String() + path
	if args.Len() > 0 {
		uri += "?" + args.String()
	}
	req.SetRequestURI(uri)
	req.Header.SetMethodBytes(c.Context().Method())
	req.SetBodyRaw(c.Body())
	copyRequestHeaders(req, &c.Context().Request)
	req.Header.SetHost(up.URL.Host)

	timeout := time.Duration(rt.Server.TimeoutMS) * time.Millisecond
	err := p.client.DoTimeout(req, res, timeout)
	if err != nil {
		errType := classifyError(err)
		resp.status = statusFromError(errType)
		p.recordUpstreamMetrics(groupName, up.HostLabel, resp.status)
		return resp, errType, err
	}

	resp.status = res.StatusCode()
	resp.body = append(resp.body, res.Body()...)
	res.Header.VisitAll(func(k, v []byte) {
		key := string(k)
		if isHopByHop(key) {
			return
		}
		resp.headers = append(resp.headers, [2][]byte{append([]byte(nil), k...), append([]byte(nil), v...)})
	})

	p.recordUpstreamMetrics(groupName, up.HostLabel, resp.status)
	return resp, "", nil
}

func copyRequestHeaders(dst *fasthttp.Request, src *fasthttp.Request) {
	src.Header.VisitAll(func(k, v []byte) {
		key := string(k)
		if isHopByHop(key) {
			return
		}
		if strings.EqualFold(key, "host") {
			return
		}
		dst.Header.AddBytesKV(k, v)
	})
}

func copyResponse(c *fiber.Ctx, resp responseData) error {
	for _, kv := range resp.headers {
		c.Response().Header.AddBytesKV(kv[0], kv[1])
	}
	c.Status(resp.status)
	return c.Send(resp.body)
}

func (p *Proxy) recordUpstreamMetrics(groupName string, upstreamLabel string, status int) {
	if p.metrics == nil {
		return
	}
	p.metrics.UpstreamRequests.WithLabelValues(groupName, upstreamLabel, statusLabel(status)).Inc()
}

func (p *Proxy) recordRequestMetrics(groupName string, routePrefix string, method string, status int, start time.Time) {
	if p.metrics == nil {
		return
	}
	p.metrics.Requests.WithLabelValues(method, groupName, routePrefix, statusLabel(status)).Inc()
	p.metrics.RequestDuration.WithLabelValues(groupName, routePrefix).Observe(time.Since(start).Seconds())
}

func (p *Proxy) recordFailure(group *runtime.GroupRuntime, up *upstream.State, errType string, status int) {
	if errType == "client" {
		return
	}
	if errType == "" && status < 500 {
		return
	}
	if !group.PassiveEject.Enabled {
		return
	}

	ejected, until := up.RecordFailure(time.Now(), group.PassiveEject.FailThreshold,
		time.Duration(group.PassiveEject.BaseEjectMS)*time.Millisecond,
		time.Duration(group.PassiveEject.MaxEjectMS)*time.Millisecond)
	if ejected {
		if p.metrics != nil {
			p.metrics.UpstreamEject.WithLabelValues(group.Name, up.HostLabel).Inc()
		}
		if p.logger != nil {
			p.logger.Warn("upstream eject", zap.String("group", group.Name), zap.String("upstream", up.HostLabel), zap.Time("eject_until", until))
		}
	}
}

func shouldRetry(errType string) bool {
	return errType == "timeout" || errType == "connect"
}

func statusLabel(status int) string {
	if status == 0 {
		return "0"
	}
	return strconv.Itoa(status)
}

func classifyError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, fasthttp.ErrTimeout) {
		return "timeout"
	}
	if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		return "timeout"
	}
	return "connect"
}

func statusFromError(errType string) int {
	if errType == "timeout" {
		return http.StatusGatewayTimeout
	}
	return http.StatusBadGateway
}

func buildChain(groupName string, rt *runtime.Runtime) []string {
	group := rt.Groups[groupName]
	if group == nil {
		return []string{groupName}
	}
	chain := make([]string, 0, 1+len(group.FallbackGroups))
	chain = append(chain, groupName)
	chain = append(chain, group.FallbackGroups...)
	return chain
}

func (p *Proxy) logAccess(start time.Time, method string, path string, group string, routePrefix string, status int, retries int, fallbackChain []string, errType string, err error) {
	if p.logger == nil {
		return
	}
	fields := []zap.Field{
		zap.Time("ts", time.Now()),
		zap.String("method", method),
		zap.String("path", path),
		zap.String("group", group),
		zap.String("route_prefix", routePrefix),
		zap.Int("status", status),
		zap.Int("duration_ms", int(time.Since(start).Milliseconds())),
		zap.Int("retries", retries),
	}
	if len(fallbackChain) > 0 {
		fields = append(fields, zap.Strings("fallback_chain", fallbackChain))
	}
	if errType != "" {
		fields = append(fields, zap.String("err_type", errType))
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	p.logger.Info("access", fields...)
}
