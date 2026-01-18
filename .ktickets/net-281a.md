---
id: net-281a
status: closed
deps:
- net-1bb8
created: "2026-01-16T12:46:17Z"
type: task
priority: 2
assignee: kostyay
parent: net-6ff0
tests_passed: false
---
# Handle y/n confirmation in kill mode

y confirms kill via syscall.Kill; any other key cancels; set killResult message
