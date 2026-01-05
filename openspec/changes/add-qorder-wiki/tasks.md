## 1. Implementation
- [x] 1.1 Add RepoWiki handler integration for `/wiki` using go-embed-qorder-wiki.
- [x] 1.2 Mount wiki before gateway auth and add `/wiki` to bypass paths in config examples.
- [x] 1.3 Wire GitSource ref to the build git commit (cobra/ldflags).
- [x] 1.4 Configure CDN assets for Mermaid/KaTeX in the wiki handler.
- [x] 1.5 Update Dockerfile.goreleaser and GoReleaser config to include `.qoder/repowiki/zh` assets.
- [x] 1.6 Update README files with `/wiki` access info.
- [x] 1.7 Add/adjust tests for wiki route availability (if feasible).
