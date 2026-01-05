## Why
Short subscriptions currently only support HTTP 301 redirects. We need flexible behaviors (301/302 redirects and direct proxying) to support both browser-friendly links and in-place proxy delivery.

## What Changes
- Add `method` to short entries: `301`, `302`, or `proxy` (default `301`).
- Support proxy mode for both internal targets (`/rsshub/`, `/upvote/`) and external `https://` targets.
- Adjust query handling: internal proxy keeps `key`/`code`, external proxy strips them; external redirects also strip `key`/`code`.
- Update short auth behavior: redirect methods bypass gateway auth; internal proxy flows through normal gateway auth.

## Impact
- Affected specs: gateway-short.
- Affected code: short config schema/validation, short handler, proxy routing, docs/tests.
