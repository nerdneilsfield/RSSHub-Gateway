## 1. Implementation
- [x] 1.1 Extend group config with backend type and strip_prefix validation/defaults.
- [x] 1.2 Add proxy path rewrite using strip_prefix and backend-specific query rewrite.
- [x] 1.3 Update routing to select groups for the chosen backend and pass rewritten path to proxy.
- [x] 1.4 Add tests for prefix stripping and Upvote query passthrough (no code injection).
- [x] 1.5 Update config examples and README/MVP/TDD docs for multi-backend routing.
