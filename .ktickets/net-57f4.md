---
id: net-57f4
status: closed
created: "2026-01-16T16:52:08Z"
type: task
priority: 2
assignee: kostyay
parent: net-4aaa
tests_passed: false
---
# Split view.go into smaller files

view.go 1003 LOC; split into: view_render.go (main View + layout), view_table.go (columns/headers/rows), view_sort.go (sorting funcs), format.go (formatBytes, truncateAddr, etc)
