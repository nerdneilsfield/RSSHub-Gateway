package auth

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/config"
	"github.com/valyala/fasthttp"
)

func ValidateGateway(cfg config.GatewayAuthConfig, path string, args *fasthttp.Args) bool {
	if !cfg.Enabled {
		return true
	}
	key := string(args.Peek("key"))
	code := string(args.Peek("code"))
	if cfg.AcceptKey && key != "" && key == cfg.AccessKey {
		return true
	}
	if cfg.AcceptCode && code != "" {
		expected := md5Hex(path + cfg.AccessKey)
		return code == expected
	}
	return false
}

func RewriteUpstreamQuery(args *fasthttp.Args, path string, upstreamKey string, injectCode bool) *fasthttp.Args {
	var out fasthttp.Args
	args.VisitAll(func(k, v []byte) {
		key := string(k)
		if key == "key" || key == "code" {
			return
		}
		out.AddBytesKV(k, v)
	})
	if injectCode && upstreamKey != "" {
		out.Set("code", md5Hex(path+upstreamKey))
	}
	return &out
}

func md5Hex(input string) string {
	sum := md5.Sum([]byte(input))
	return hex.EncodeToString(sum[:])
}
