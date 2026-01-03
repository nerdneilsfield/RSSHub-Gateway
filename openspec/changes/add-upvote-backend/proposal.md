## Why
The gateway needs to support Upvote RSS alongside RSSHub while keeping a single shared access key and clear path separation.

## What Changes
- Add backend-aware routing for RSSHub and Upvote RSS with `/rsshub/` and `/upvote/` prefixes.
- Add explicit prefix stripping before proxying to upstreams.
- Apply backend-specific query rewrite: RSSHub injects upstream code, Upvote passes through without code injection.
- Update MVP/TDD and config guidance to document multi-backend routing.

## Impact
- Affected specs: gateway-routing (new)
- Affected code: config schema, router/proxy path rewrite, auth injection, tests, docs
