## 1. Implementation
- [ ] 1.1 Extend group config with backend type and strip_prefix validation/defaults.
- [ ] 1.2 Add proxy path rewrite using strip_prefix and backend-specific query rewrite.
- [ ] 1.3 Update routing to select groups for the chosen backend and pass rewritten path to proxy.
- [ ] 1.4 Add tests for prefix stripping and Upvote query passthrough (no code injection).
- [ ] 1.5 Update config examples and README/MVP/TDD docs for multi-backend routing.
