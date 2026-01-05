## ADDED Requirements
### Requirement: Embedded Wiki Route
The gateway SHALL serve the RepoWiki content at the `/wiki` route.

#### Scenario: Serve wiki content
- **WHEN** a request targets `/wiki` or `/wiki/...`
- **THEN** the gateway responds with the wiki HTML/content rendered from `.qoder/repowiki/zh`.

### Requirement: Wiki Auth Bypass
The gateway SHALL allow wiki access without gateway auth.

#### Scenario: Wiki bypasses gateway auth
- **WHEN** a request targets `/wiki/...` without key/code
- **THEN** the gateway serves the wiki content instead of returning 403.

### Requirement: Wiki Asset Loading via CDN
The wiki renderer SHALL use CDN-based Mermaid and KaTeX assets.

#### Scenario: Mermaid/KaTeX load from CDN
- **WHEN** the wiki page includes Mermaid or KaTeX
- **THEN** the HTML references CDN URLs for these assets.

### Requirement: GitHub Link Rewriting
The wiki renderer SHALL rewrite `file://` links to GitHub blob URLs using a pinned commit ref.

#### Scenario: Links use build commit ref
- **WHEN** the wiki contains a `file://` link
- **THEN** the rendered link points to `https://github.com/nerdneilsfield/RSSHub-Gateway/blob/<gitCommit>/...`.

### Requirement: Release Packaging Includes Wiki Assets
Docker and GoReleaser artifacts SHALL include `.qoder/repowiki/zh` so the wiki can load at runtime.

#### Scenario: Docker image contains wiki assets
- **WHEN** the gateway runs in a Docker image produced by CI
- **THEN** `.qoder/repowiki/zh` exists and `/wiki` serves content.
