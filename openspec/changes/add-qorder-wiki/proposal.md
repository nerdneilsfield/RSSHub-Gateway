## Why
We need to serve the project’s Qoder RepoWiki content directly from the gateway so users can browse documentation at `/wiki` without gateway auth friction.

## What Changes
- Embed or mount the local `.qoder/repowiki/zh` content and serve it at `/wiki`.
- Bypass gateway auth for the wiki route.
- Configure RepoWiki assets to load Mermaid/KaTeX via CDN.
- Wire GitHub link rewriting to a pinned commit ref (not a branch) using the build git hash.
- Ensure release packaging (Dockerfile.goreleaser / GoReleaser) includes the wiki assets.

## Impact
- Affected specs: gateway-wiki (new capability).
- Affected code: proxy routing, new wiki handler integration, config examples, build packaging.
