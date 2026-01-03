## Why
Users need a short, stable subscription entrypoint that can redirect to internal RSSHub/Upvote routes or external feed URLs while preserving query parameters.

## What Changes
- Add short subscription entries that redirect with HTTP 301 to internal `/rsshub/` or `/upvote/` targets, or external `https://` URLs.
- Add configurable short path prefix (default `/short`) and validation for entries and targets.
- Define query passthrough behavior to preserve `key`, `code`, and other parameters.

## Impact
- Affected specs: gateway-short
- Affected code: config schema/validation, routing/handler for short redirect, docs/tests
