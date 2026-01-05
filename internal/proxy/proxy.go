package proxy

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/auth"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/cache"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/home"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/metrics"
	pprofhandler "github.com/nerdneilsfield/RSSHub-Gateway/internal/pprof"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/runtime"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/short"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Proxy struct {
	manager *runtime.Manager
	metrics *metrics.Metrics
	logger  *zap.Logger
	client  *fasthttp.Client
	home    *home.Renderer

	externalClients sync.Map
}

func New(manager *runtime.Manager, m *metrics.Metrics, logger *zap.Logger) *Proxy {
	return &Proxy{
		manager: manager,
		metrics: m,
		logger:  logger,
		client:  &fasthttp.Client{},
		home:    home.New("README.md", "README_zh.md", logger),
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
	if method == fiber.MethodGet && p.home != nil {
		if path == "/" || path == "/zh" || path == "/zh/" || path == "/en" || path == "/en/" {
			return p.home.Serve(c)
		}
	}
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
	effectivePath := path
	if result, matched, ok := short.Resolve(rt.Short, path); matched {
		if !ok {
			return c.SendStatus(http.StatusNotFound)
		}
		switch result.Method {
		case "301", "302":
			redirectArgs := args
			if !result.Internal {
				redirectArgs = auth.RewriteUpstreamQuery(args, "", "", false)
			}
			location := short.AppendQuery(result.Target, redirectArgs.String())
			c.Location(location)
			if result.Method == "302" {
				return c.SendStatus(http.StatusFound)
			}
			return c.SendStatus(http.StatusMovedPermanently)
		case "proxy":
			if result.Internal {
				targetPath, targetQuery := splitTarget(result.Target)
				if targetPath == "" {
					return c.SendStatus(http.StatusNotFound)
				}
				effectivePath = targetPath
				args = mergeQueryArgs(targetQuery, args)
			} else {
				return p.proxyExternal(c, rt, result.Target, args, start, method, path)
			}
		default:
			return c.SendStatus(http.StatusNotFound)
		}
	}

	if !auth.ValidateGateway(rt.Auth, effectivePath, args) {
		p.logAccess(start, method, path, "", "", http.StatusForbidden, 0, nil, "auth", errors.New("gateway auth failed"))
		return c.SendStatus(http.StatusForbidden)
	}

	selection := rt.Router.Select(effectivePath)
	groupName := selection.Group
	routePrefix := selection.RoutePrefix
	cacheKey := ""
	cacheEnabled := method == fiber.MethodGet && rt.Cache.Enabled && rt.CacheStore != nil
	if cacheEnabled {
		group := rt.Groups[groupName]
		if group != nil {
			upstreamPath := rewritePath(effectivePath, group.StripPrefix)
			cacheKey = cache.BuildKey(upstreamPath, args)
			if entry, ok := rt.CacheStore.GetResponse(cacheKey); ok {
				if p.metrics != nil {
					p.metrics.CacheHit.WithLabelValues(rt.CacheStore.Provider()).Inc()
				}
				status := entry.Status
				p.recordRequestMetrics(groupName, routePrefix, method, status, start)
				p.logAccess(start, method, path, groupName, routePrefix, status, 0, []string{groupName}, "cache", nil)
				return copyResponse(c, responseFromCache(entry))
			}
			if p.metrics != nil {
				p.metrics.CacheMiss.WithLabelValues(rt.CacheStore.Provider()).Inc()
			}
		}
	}
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

		upstreamPath := rewritePath(effectivePath, group.StripPrefix)
		injectCode := group.Backend != "upvote"
		attempts := 1
		if rt.Failover.Retry.Enabled && (method == fiber.MethodGet || method == fiber.MethodHead) {
			if rt.Failover.Retry.MaxRetries > 0 {
				attempts = 1 + rt.Failover.Retry.MaxRetries
			}
		}
		avoid := map[*upstream.State]struct{}{}
		for attempt := 0; attempt < attempts; attempt++ {
			up := group.Picker.Pick(upstreamPath, time.Now(), avoid)
			if up == nil {
				break
			}
			if attempt > 0 {
				retries++
				if p.metrics != nil {
					p.metrics.RetryTotal.WithLabelValues(group.Name).Inc()
				}
			}
			resp, errType, err := p.forward(c, rt, group.Name, up, upstreamPath, injectCode)
			status = resp.status
			lastErrType = errType
			lastErr = err

			p.recordUpstreamMetrics(group.Name, up.HostLabel, resp.status)
			if err == nil {
				if cacheEnabled && cacheKey != "" && shouldCacheStatus(resp.status) {
					entry := cacheEntryFromResponse(resp)
					_ = rt.CacheStore.SetResponse(cacheKey, entry, time.Duration(rt.Cache.TTLMS)*time.Millisecond)
				}
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

func (p *Proxy) forward(c *fiber.Ctx, rt *runtime.Runtime, groupName string, up *upstream.State, upstreamPath string, injectCode bool) (responseData, string, error) {
	var resp responseData
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	args := auth.RewriteUpstreamQuery(c.Context().QueryArgs(), upstreamPath, up.AccessKey, injectCode)
	uri := up.URL.String() + upstreamPath
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
	if isSuccessStatus(status) {
		p.metrics.UpstreamSuccess.WithLabelValues(groupName, upstreamLabel).Inc()
	} else {
		p.metrics.UpstreamFailure.WithLabelValues(groupName, upstreamLabel).Inc()
	}
}

func (p *Proxy) recordRequestMetrics(groupName string, routePrefix string, method string, status int, start time.Time) {
	if p.metrics == nil {
		return
	}
	p.metrics.Requests.WithLabelValues(method, groupName, routePrefix, statusLabel(status)).Inc()
	p.metrics.RequestDuration.WithLabelValues(groupName, routePrefix).Observe(time.Since(start).Seconds())
	if isSuccessStatus(status) {
		p.metrics.RouteSuccess.WithLabelValues(groupName, routePrefix).Inc()
	} else {
		p.metrics.RouteFailure.WithLabelValues(groupName, routePrefix).Inc()
	}
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

func isSuccessStatus(status int) bool {
	return status >= 200 && status < 400
}

func shouldCacheStatus(status int) bool {
	return isSuccessStatus(status)
}

func cacheEntryFromResponse(resp responseData) cache.Entry {
	headers := make([]cache.Header, 0, len(resp.headers))
	for _, kv := range resp.headers {
		headers = append(headers, cache.Header{
			Key:   string(kv[0]),
			Value: string(kv[1]),
		})
	}
	return cache.Entry{
		Status:  resp.status,
		Headers: headers,
		Body:    append([]byte(nil), resp.body...),
	}
}

func responseFromCache(entry cache.Entry) responseData {
	headers := make([][2][]byte, 0, len(entry.Headers))
	for _, header := range entry.Headers {
		headers = append(headers, [2][]byte{[]byte(header.Key), []byte(header.Value)})
	}
	return responseData{
		status:  entry.Status,
		headers: headers,
		body:    append([]byte(nil), entry.Body...),
	}
}

func splitTarget(target string) (string, string) {
	if target == "" {
		return "", ""
	}
	idx := strings.Index(target, "?")
	if idx == -1 {
		return target, ""
	}
	return target[:idx], target[idx+1:]
}

func mergeQueryArgs(targetQuery string, original *fasthttp.Args) *fasthttp.Args {
	var out fasthttp.Args
	if targetQuery != "" {
		values, err := url.ParseQuery(targetQuery)
		if err == nil {
			for key, vals := range values {
				for _, val := range vals {
					out.Add(key, val)
				}
			}
		}
	}
	if original != nil {
		original.VisitAll(func(k, v []byte) {
			out.AddBytesKV(k, v)
		})
	}
	return &out
}

func (p *Proxy) proxyExternal(c *fiber.Ctx, rt *runtime.Runtime, target string, args *fasthttp.Args, start time.Time, method string, path string) error {
	uri, err := buildExternalURL(target, args)
	if err != nil {
		p.logExternalError(start, method, path, target, http.StatusBadGateway, "external", err)
		p.logAccess(start, method, path, "short-external", "short", http.StatusBadGateway, 0, nil, "external", err)
		return c.SendStatus(http.StatusBadGateway)
	}
	timeout := time.Duration(rt.Server.TimeoutMS) * time.Millisecond
	resp, errType, err := p.forwardExternal(c, uri, timeout)
	status := resp.status
	if err != nil {
		p.logExternalError(start, method, path, uri, status, errType, err)
		p.recordRequestMetrics("short-external", "short", method, status, start)
		p.logAccess(start, method, path, "short-external", "short", status, 0, nil, errType, err)
		return c.SendStatus(status)
	}
	if status >= http.StatusBadRequest {
		p.logExternalError(start, method, path, uri, status, "upstream", nil)
	}
	p.recordRequestMetrics("short-external", "short", method, status, start)
	p.logAccess(start, method, path, "short-external", "short", status, 0, nil, "", nil)
	return copyResponse(c, resp)
}

func (p *Proxy) forwardExternal(c *fiber.Ctx, uri string, timeout time.Duration) (responseData, string, error) {
	var resp responseData
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(uri)
	req.Header.SetMethodBytes(c.Context().Method())
	req.SetBodyRaw(c.Body())
	copyRequestHeaders(req, &c.Context().Request)

	parsed, err := url.Parse(uri)
	if err == nil && parsed.Host != "" {
		req.Header.SetHost(parsed.Host)
	}
	if len(req.Header.UserAgent()) == 0 {
		req.Header.SetUserAgent("RSSHub-Gateway/short")
	}

	client := p.externalClient(parsed)
	err = client.DoTimeout(req, res, timeout)
	if err != nil {
		errType := classifyError(err)
		resp.status = statusFromError(errType)
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

	return resp, "", nil
}

func (p *Proxy) externalClient(target *url.URL) *fasthttp.Client {
	if target == nil {
		return p.client
	}
	key := target.Scheme + "://" + target.Host
	if key == "://" {
		return p.client
	}
	if client, ok := p.externalClients.Load(key); ok {
		return client.(*fasthttp.Client)
	}
	created := &fasthttp.Client{}
	if target.Scheme == "https" {
		serverName := target.Hostname()
		if serverName != "" {
			created.TLSConfig = &tls.Config{ServerName: serverName}
		}
	}
	client, _ := p.externalClients.LoadOrStore(key, created)
	return client.(*fasthttp.Client)
}

func buildExternalURL(target string, args *fasthttp.Args) (string, error) {
	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	sanitized := auth.RewriteUpstreamQuery(args, "", "", false)
	extra := sanitized.String()
	parsed.RawQuery = mergeQuery(parsed.RawQuery, extra)
	return parsed.String(), nil
}

func mergeQuery(base string, extra string) string {
	if base == "" {
		return extra
	}
	if extra == "" {
		return base
	}
	return base + "&" + extra
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

func rewritePath(path string, stripPrefix string) string {
	if stripPrefix == "" {
		return path
	}
	if !strings.HasPrefix(path, stripPrefix) {
		return path
	}
	trimmed := strings.TrimPrefix(path, stripPrefix)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
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

func (p *Proxy) logExternalError(start time.Time, method string, path string, target string, status int, errType string, err error) {
	if p.logger == nil {
		return
	}
	fields := []zap.Field{
		zap.Time("ts", time.Now()),
		zap.String("method", method),
		zap.String("path", path),
		zap.String("target", target),
		zap.Int("status", status),
		zap.Int("duration_ms", int(time.Since(start).Milliseconds())),
	}
	if errType != "" {
		fields = append(fields, zap.String("err_type", errType))
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	p.logger.Warn("short external error", fields...)
}
