package router

import "strings"

type Route struct {
	Name     string
	Allow    []string
	Deny     []string
	Priority int
	Order    int
}

type Router struct {
	Routes       []Route
	DefaultGroup string
}

func New(routes []Route, defaultGroup string) *Router {
	return &Router{Routes: routes, DefaultGroup: defaultGroup}
}

func (r *Router) Select(path string) string {
	bestLen := -1
	bestPriority := 0
	bestOrder := 0
	bestName := ""
	for _, route := range r.Routes {
		if matchesDeny(path, route.Deny) {
			continue
		}
		prefixLen := longestAllow(path, route.Allow)
		if prefixLen == -1 {
			continue
		}
		if prefixLen > bestLen || (prefixLen == bestLen && route.Priority > bestPriority) ||
			(prefixLen == bestLen && route.Priority == bestPriority && route.Order < bestOrder) {
			bestLen = prefixLen
			bestPriority = route.Priority
			bestOrder = route.Order
			bestName = route.Name
		}
	}
	if bestName == "" {
		return r.DefaultGroup
	}
	return bestName
}

func matchesDeny(path string, deny []string) bool {
	for _, prefix := range deny {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func longestAllow(path string, allow []string) int {
	best := -1
	for _, prefix := range allow {
		if strings.HasPrefix(path, prefix) {
			if len(prefix) > best {
				best = len(prefix)
			}
		}
	}
	return best
}
