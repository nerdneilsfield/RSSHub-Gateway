package short

import "strings"

type Runtime struct {
	Enabled bool
	Path    string
	Targets map[string]Entry
}

type Entry struct {
	Target   string
	Method   string
	Internal bool
}

type Result struct {
	Target   string
	Method   string
	Internal bool
}

func NewRuntime(enabled bool, path string, entries map[string]Entry) *Runtime {
	return &Runtime{
		Enabled: enabled,
		Path:    path,
		Targets: entries,
	}
}

func Resolve(rt *Runtime, path string) (Result, bool, bool) {
	if rt == nil || !rt.Enabled {
		return Result{}, false, false
	}
	if !matchesPath(path, rt.Path) {
		return Result{}, false, false
	}
	name := extractName(path, rt.Path)
	if name == "" {
		return Result{}, true, false
	}
	entry, exists := rt.Targets[name]
	if !exists {
		return Result{}, true, false
	}
	return Result{Target: entry.Target, Method: entry.Method, Internal: entry.Internal}, true, true
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

func AppendQuery(target string, rawQuery string) string {
	if rawQuery == "" {
		return target
	}
	if strings.Contains(target, "?") {
		return target + "&" + rawQuery
	}
	return target + "?" + rawQuery
}

func IsInternalTarget(target string) bool {
	return target == "/rsshub" || strings.HasPrefix(target, "/rsshub/") ||
		target == "/upvote" || strings.HasPrefix(target, "/upvote/")
}
