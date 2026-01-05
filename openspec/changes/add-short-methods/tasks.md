## 1. Implementation
- [ ] 1.1 Extend short config schema with `method` (default 301) and validation rules.
- [ ] 1.2 Update short runtime/resolve to return method + target details.
- [ ] 1.3 Implement redirect methods 301/302 with full query passthrough.
- [ ] 1.4 Implement proxy method for internal targets (route through gateway + auth) and external targets (strip key/code, direct proxy).
- [ ] 1.5 Add tests for redirect 302, proxy internal, proxy external, and auth behavior.
- [ ] 1.6 Update docs and config examples for short methods.
