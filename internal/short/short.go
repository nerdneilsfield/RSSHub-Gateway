package short

import "strings"

type Runtime struct {
	Enabled bool
	Path    string
	Targets map[string]string
}

func NewRuntime(enabled bool, path string, entries map[string]string) *Runtime {
	return &Runtime{
		Enabled: enabled,
		Path:    path,
		Targets: entries,
	}
}

func Resolve(rt *Runtime, path string, rawQuery string) (location string, matched bool, ok bool) {
	if rt == nil || !rt.Enabled {
		return "", false, false
	}
	if !matchesPath(path, rt.Path) {
		return "", false, false
	}
	name := extractName(path, rt.Path)
	if name == "" {
		return "", true, false
	}
	target, exists := rt.Targets[name]
	if !exists {
		return "", true, false
	}
	return appendQuery(target, rawQuery), true, true
}

func matchesPath(path string, base string) bool {
	if base == "" {
		return false
	}
	if path == base {
		return true
	}
	return strings.HasPrefix(path, base+"/")
}

func extractName(path string, base string) string {
	rest := strings.TrimPrefix(path, base)
	if rest == "" {
		return ""
	}
	if !strings.HasPrefix(rest, "/") {
		return ""
	}
	name := strings.TrimPrefix(rest, "/")
	name = strings.TrimSuffix(name, "/")
	if name == "" {
		return ""
	}
	if strings.Contains(name, "/") {
		return ""
	}
	return name
}

func appendQuery(target string, rawQuery string) string {
	if rawQuery == "" {
		return target
	}
	if strings.Contains(target, "?") {
		return target + "&" + rawQuery
	}
	return target + "?" + rawQuery
}
