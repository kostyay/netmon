---
id: net-d394
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
# Update rendering in view.go

Compute cursorIdx once via resolveSelectionIndex(). Use isSelected := i == cursorIdx for highlighting.
