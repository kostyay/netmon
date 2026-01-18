---
id: net-d718
status: closed
deps:
- net-0055
created: "2026-01-17T22:33:00Z"
type: task
priority: 2
assignee: kostyay
parent: net-1e6b
tests_passed: false
---
# Simplify kill.go to use SelectedID directly

Use SelectedID.ProcessName or SelectedID.ConnectionKey directly instead of cursor lookup. No filter/sort needed anymore.
