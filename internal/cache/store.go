package cache

import (
	"errors"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

const ConfigHashKey = "config_hash"

var ErrEntryTooLarge = errors.New("cache entry too large")

type Options struct {
	Enabled       bool
	Provider      string
	TTL           time.Duration
	MaxItemBytes  int
	MaxTotalBytes int
	Redis         RedisOptions
}

type RedisOptions struct {
	Addr           string
	Password       string
	DB             int
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	KeyPrefix      string
}

type Header struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type Entry struct {
	Status  int      `json:"status"`
	Headers []Header `json:"headers"`
	Body    []byte   `json:"body"`
}

type Store interface {
	Provider() string
	GetResponse(key string) (Entry, bool)
	SetResponse(key string, entry Entry, ttl time.Duration) error
	GetString(key string) (string, bool, error)
	SetString(key string, value string, ttl time.Duration) error
	Close() error
}

func BuildKey(path string, args *fasthttp.Args) string {
	values := url.Values{}
	args.VisitAll(func(k, v []byte) {
		key := string(k)
		if key == "key" || key == "code" {
			return
		}
		values.Add(key, string(v))
	})
	if len(values) == 0 {
		return path
	}
	query := encodeValues(values)
	return path + "?" + query
}

func encodeValues(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	first := true
	for _, key := range keys {
		vals := values[key]
		if len(vals) == 0 {
			continue
		}
		sort.Strings(vals)
		for _, val := range vals {
			if !first {
				b.WriteByte('&')
			}
			first = false
			b.WriteString(url.QueryEscape(key))
			b.WriteByte('=')
			b.WriteString(url.QueryEscape(val))
		}
	}
	return b.String()
}

func EntrySize(entry Entry) int {
	size := len(entry.Body)
	for _, header := range entry.Headers {
		size += len(header.Key) + len(header.Value)
	}
	return size
}
