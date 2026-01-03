package proxy

import "strings"

var hopByHopHeaders = map[string]struct{}{
	"connection":          {},
	"proxy-connection":    {},
	"keep-alive":          {},
	"transfer-encoding":   {},
	"upgrade":             {},
	"proxy-authenticate":  {},
	"proxy-authorization": {},
	"te":                  {},
	"trailer":             {},
}

func isHopByHop(header string) bool {
	_, ok := hopByHopHeaders[strings.ToLower(header)]
	return ok
}
