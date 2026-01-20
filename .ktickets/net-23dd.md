---
id: net-23dd
status: closed
created: "2026-01-16T16:52:08Z"
type: bug
priority: 1
assignee: kostyay
parent: net-4aaa
tests_passed: false
---
# Fix nil panic in RenderJSON

internal/output/json.go:57 will panic if ioStats is nil; add nil check before iterating
