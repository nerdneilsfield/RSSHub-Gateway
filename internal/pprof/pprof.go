package pprof

import (
	"net/http"
	netpprof "net/http/pprof"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

const defaultBasePath = "/debug/pprof"

func normalizeBasePath(basePath string) string {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" {
		return defaultBasePath
	}
	if basePath != "/" && strings.HasSuffix(basePath, "/") {
		return strings.TrimSuffix(basePath, "/")
	}
	return basePath
}

func MatchPath(path string, basePath string) bool {
	base := normalizeBasePath(basePath)
	if path == base {
		return true
	}
	return strings.HasPrefix(path, base+"/")
}

func Handle(c *fiber.Ctx, basePath string, accessKey string) error {
	if accessKey != "" && c.Query("accesskey") != accessKey {
		return c.SendStatus(http.StatusForbidden)
	}

	path := c.Path()
	base := normalizeBasePath(basePath)
	if path == base {
		return c.Redirect(base+"/", http.StatusMovedPermanently)
	}

	suffix := strings.TrimPrefix(path, base)
	if suffix == "" || suffix == "/" {
		return adaptor.HTTPHandlerFunc(netpprof.Index)(c)
	}

	name := strings.TrimPrefix(suffix, "/")
	switch name {
	case "cmdline":
		return adaptor.HTTPHandlerFunc(netpprof.Cmdline)(c)
	case "profile":
		return adaptor.HTTPHandlerFunc(netpprof.Profile)(c)
	case "symbol":
		return adaptor.HTTPHandlerFunc(netpprof.Symbol)(c)
	case "trace":
		return adaptor.HTTPHandlerFunc(netpprof.Trace)(c)
	default:
		return adaptor.HTTPHandler(netpprof.Handler(name))(c)
	}
}
