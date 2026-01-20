---
id: net-c9dd
status: closed
created: "2026-01-16T16:52:08Z"
type: task
priority: 2
assignee: kostyay
parent: net-4aaa
tests_passed: false
---
# Consolidate duplicate extractPort functions

Duplicate in cmd/netmon/kill.go:151 and internal/ui/update.go:577; move to internal/model/network.go as ExtractPort()
