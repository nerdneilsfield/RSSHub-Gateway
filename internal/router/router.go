package router

import "strings"

type Route struct {
	Name     string
	Allow    []string
	Deny     []string
	Priority int
	Order    int
}

type Selection struct {
	Group       string
	RoutePrefix string
}

const DefaultRoutePrefix = "default"

type Router struct {
	Routes       []Route
	DefaultGroup string
}

func New(routes []Route, defaultGroup string) *Router {
	return &Router{Routes: routes, DefaultGroup: defaultGroup}
}

func (r *Router) Select(path string) Selection {
	bestLen := -1
	bestPriority := 0
	bestOrder := 0
	bestName := ""
	bestPrefix := ""
	for _, route := range r.Routes {
		if matchesDeny(path, route.Deny) {
			continue
		}
		prefix, prefixLen := longestAllow(path, route.Allow)
		if prefixLen == -1 {
			continue
		}
		if prefixLen > bestLen || (prefixLen == bestLen && route.Priority > bestPriority) ||
			(prefixLen == bestLen && route.Priority == bestPriority && route.Order < bestOrder) {
			bestLen = prefixLen
			bestPriority = route.Priority
			bestOrder = route.Order
			bestName = route.Name
			bestPrefix = prefix
		}
	}
	if bestName == "" {
		return Selection{Group: r.DefaultGroup, RoutePrefix: DefaultRoutePrefix}
	}
	return Selection{Group: bestName, RoutePrefix: bestPrefix}
}

func matchesDeny(path string, deny []string) bool {
	for _, prefix := range deny {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func longestAllow(path string, allow []string) (string, int) {
	best := -1
	bestPrefix := ""
	for _, prefix := range allow {
		if strings.HasPrefix(path, prefix) {
			if len(prefix) > best {
				best = len(prefix)
				bestPrefix = prefix
			}
		}
	}
	return bestPrefix, best
}
