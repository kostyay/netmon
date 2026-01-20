---
id: net-2aa7
status: closed
created: "2026-01-16T16:52:08Z"
type: task
priority: 2
assignee: kostyay
parent: net-4aaa
tests_passed: false
---
# Consolidate duplicate signalMap

Duplicate in cmd/netmon/kill.go:23 and internal/ui/update.go:588; create internal/process/signals.go with shared map
