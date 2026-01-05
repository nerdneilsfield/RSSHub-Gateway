package home

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"go.uber.org/zap"
)

type Renderer struct {
	readmePaths map[string]string
	defaultLang string
	logger      *zap.Logger
}

type pageData struct {
	Title      string
	Content    template.HTML
	Lang       string
	EnglishURL string
	ChineseURL string
}

const pageTemplate = `<!doctype html>
<html lang="{{ .Lang }}">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/github-markdown-css@5.5.1/github-markdown.min.css">
  <style>
    body {
      box-sizing: border-box;
      max-width: 980px;
      margin: 0 auto;
      padding: 32px 20px 48px;
      background: #f6f8fa;
      color: #24292f;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", Helvetica, Arial, sans-serif;
    }
    .lang-switch {
      display: flex;
      justify-content: flex-end;
      font-size: 14px;
      margin-bottom: 12px;
      gap: 8px;
    }
    .lang-switch a {
      color: #0969da;
      text-decoration: none;
      font-weight: 600;
    }
    .lang-switch a.active {
      color: #24292f;
    }
    .markdown-body {
      background: #ffffff;
      border-radius: 16px;
      padding: 32px 32px 24px;
      box-shadow: 0 8px 30px rgba(0, 0, 0, 0.08);
    }
    @media (prefers-color-scheme: dark) {
      body {
        background: #0b0f17;
        color: #e6edf3;
      }
      .lang-switch a {
        color: #58a6ff;
      }
      .lang-switch a.active {
        color: #e6edf3;
      }
      .markdown-body {
        background: #0d1117;
        box-shadow: 0 12px 40px rgba(0, 0, 0, 0.45);
      }
    }
  </style>
</head>
<body>
  <div class="lang-switch">
    <a href="{{ .EnglishURL }}" class="{{ if eq .Lang "en" }}active{{ end }}">English</a>
    <span>/</span>
    <a href="{{ .ChineseURL }}" class="{{ if eq .Lang "zh" }}active{{ end }}">中文</a>
  </div>
  <article class="markdown-body">{{ .Content }}</article>
  <script src="https://cdn.jsdelivr.net/npm/mermaid@10.9.1/dist/mermaid.min.js"></script>
  <script>
    mermaid.initialize({ startOnLoad: true });
  </script>
</body>
</html>`

var compiledTemplate = template.Must(template.New("home").Parse(pageTemplate))

func New(readmePath string, readmeZhPath string, logger *zap.Logger) *Renderer {
	if logger == nil {
		logger = zap.NewNop()
	}
	paths := map[string]string{"en": readmePath}
	if readmeZhPath != "" {
		paths["zh"] = readmeZhPath
	}
	return &Renderer{readmePaths: paths, defaultLang: "en", logger: logger}
}

func (r *Renderer) Serve(c *fiber.Ctx) error {
	lang := r.langFromRequest(c)
	output, err := r.Render(lang)
	if err != nil {
		r.logger.Warn("home render failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("home render failed")
	}
	c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
	return c.Send(output)
}

func (r *Renderer) Render(lang string) ([]byte, error) {
	path := resolveReadmePath(r.pathForLang(lang))
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read README: %w", err)
	}
	page := pageData{
		Title:      titleFromMarkdown(content),
		Content:    renderMarkdown(content),
		Lang:       lang,
		EnglishURL: "/?lang=en",
		ChineseURL: "/?lang=zh",
	}
	var buf bytes.Buffer
	if err := compiledTemplate.Execute(&buf, page); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	return buf.Bytes(), nil
}

func renderMarkdown(content []byte) template.HTML {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.FencedCode | parser.Tables | parser.Strikethrough
	p := parser.NewWithExtensions(extensions)
	renderer := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})
	doc := p.Parse(content)
	out := markdown.Render(doc, renderer)
	htmlContent := convertMermaidBlocks(string(out))
	return template.HTML(htmlContent)
}

func titleFromMarkdown(content []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return "RSSHub Gateway"
}

func resolveReadmePath(path string) string {
	if path == "" {
		return path
	}
	if filepath.IsAbs(path) {
		return path
	}
	if exists(path) {
		return path
	}
	dir, err := os.Getwd()
	if err != nil {
		return path
	}
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, path)
		if exists(candidate) {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return path
}

func exists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (r *Renderer) pathForLang(lang string) string {
	if path, ok := r.readmePaths[lang]; ok && path != "" {
		return path
	}
	if path, ok := r.readmePaths[r.defaultLang]; ok && path != "" {
		return path
	}
	for _, path := range r.readmePaths {
		if path != "" {
			return path
		}
	}
	return ""
}

func (r *Renderer) langFromRequest(c *fiber.Ctx) string {
	if value := normalizeLang(c.Query("lang")); value != "" {
		return value
	}
	path := strings.TrimSpace(strings.ToLower(c.Path()))
	switch path {
	case "/zh", "/zh/":
		return "zh"
	case "/en", "/en/":
		return "en"
	default:
		return r.defaultLang
	}
}

func normalizeLang(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "zh", "zh-cn", "zh-hans", "cn":
		return "zh"
	case "en", "en-us", "en-gb":
		return "en"
	default:
		return ""
	}
}

func convertMermaidBlocks(input string) string {
	const open = `<pre><code class="language-mermaid">`
	const close = `</code></pre>`
	var out strings.Builder
	for {
		idx := strings.Index(input, open)
		if idx == -1 {
			out.WriteString(input)
			break
		}
		out.WriteString(input[:idx])
		out.WriteString(`<pre class="mermaid">`)
		input = input[idx+len(open):]
		end := strings.Index(input, close)
		if end == -1 {
			out.WriteString(input)
			break
		}
		out.WriteString(input[:end])
		out.WriteString(`</pre>`)
		input = input[end+len(close):]
	}
	return out.String()
}
