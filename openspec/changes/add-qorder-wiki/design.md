## Context
RepoWiki content is generated under `.qoder/repowiki/zh`. We want the gateway to serve it at `/wiki` using the go-embed-qorder-wiki library and CDN assets. The gateway already has a single Fiber entry point and an auth bypass mechanism.

## Goals / Non-Goals
- Goals:
  - Serve `/wiki` from RepoWiki content with correct path handling.
  - Use CDN for Mermaid/KaTeX assets.
  - Rewrite file:// links to GitHub using a pinned commit ref (build git hash).
  - Package wiki assets in Docker/GoReleaser outputs.
- Non-Goals:
  - Multi-language wiki switching.
  - Runtime wiki regeneration.

## Decisions
- Serve wiki via the repo library’s Fiber adapter mounted at `/wiki`.
- Use `os.DirFS(".qoder/repowiki/zh")` as the source FS (runtime file mount), with Docker packaging ensuring files exist.
- Use GitSource with `RepoURL=https://github.com/nerdneilsfield/RSSHub-Gateway` and `Ref=<gitCommit>`.
- Use CDN defaults by enabling Mermaid and KaTeX CDN flags.
- Bypass gateway auth by handling `/wiki` before auth checks and adding `/wiki` to config examples.

## Risks / Trade-offs
- Runtime FS dependency means wiki files must be present in deployments; mitigate by copying `.qoder/repowiki/zh` into release artifacts/images.
- CDN availability can affect wiki rendering; use CDN defaults to reduce maintenance.

## Migration Plan
- Add wiki assets to build artifacts and images.
- Deploy with updated image containing `.qoder/repowiki/zh`.

## Open Questions
- Whether to make wiki enable/disable configurable (defaults to enabled).
