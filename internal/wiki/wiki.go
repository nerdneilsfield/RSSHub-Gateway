package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	repowiki "github.com/nerdneilsfield/go-embed-qorder-wiki"
	fiberadapter "github.com/nerdneilsfield/go-embed-qorder-wiki/adapters/fiber"
	"go.uber.org/zap"
)

const (
	defaultWikiRoot   = ".qoder/repowiki/zh"
	defaultWikiMount  = "/wiki"
	defaultWikiHome   = "主页.md"
	defaultContentDir = "content"

	mermaidCDN = "https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"
	katexCSS   = "https://cdn.jsdelivr.net/npm/katex@0.16.9/dist/katex.min.css"
	katexJS    = "https://cdn.jsdelivr.net/npm/katex@0.16.9/dist/katex.min.js"
	katexAuto  = "https://cdn.jsdelivr.net/npm/katex@0.16.9/dist/contrib/auto-render.min.js"
)

func NewHandler(root string, gitCommit string, repoURL string, logger *zap.Logger) (fiber.Handler, string, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	mount := defaultWikiMount
	if root == "" {
		root = defaultWikiRoot
	}
	resolved := resolveWikiRoot(root)
	if resolved == "" {
		return nil, mount, fmt.Errorf("wiki root not found: %s", root)
	}
	cfg := repowiki.Config{
		FS:         os.DirFS(resolved),
		Root:       ".",
		ContentDir: defaultContentDir,
		Home:       defaultWikiHome,
		Git: repowiki.GitSource{
			RepoURL: repoURL,
			Ref:     gitCommit,
		},
		Assets: repowiki.AssetConfig{
			Mermaid: repowiki.MermaidConfig{
				UseCDN: true,
				CDNURL: mermaidCDN,
			},
			KaTeX: repowiki.KaTeXConfig{
				Enabled: true,
				UseCDN:  true,
				CDN: repowiki.KaTeXCDNConfig{
					CSS:          katexCSS,
					JS:           katexJS,
					AutoRenderJS: katexAuto,
				},
			},
		},
	}

	handler, err := repowiki.New(cfg)
	if err != nil {
		return nil, mount, fmt.Errorf("init wiki handler: %w", err)
	}
	return fiberadapter.Wrap(handler, mount), mount, nil
}

func resolveWikiRoot(root string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return ""
	}
	if filepath.IsAbs(root) && dirExists(root) {
		return root
	}
	if dirExists(root) {
		return root
	}
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, root)
		if dirExists(candidate) {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
