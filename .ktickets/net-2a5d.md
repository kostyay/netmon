---
id: net-2a5d
status: closed
deps:
- net-3e94
- net-7690
- net-5440
created: "2026-01-15T00:25:13Z"
type: task
priority: 2
assignee: kostyay
parent: net-5bad
tests_passed: false
---
# Implement runJSONMode() in root.go

Wire it together: collect once, render JSON, print stdout, exit 0. Errors → JSON error object, exit 1
