---
id: net-1ca6
status: closed
created: "2026-01-16T16:52:16Z"
type: bug
priority: 3
assignee: kostyay
parent: net-4aaa
tests_passed: false
---
# Fix inconsistent error output to stderr

cmd/netmon/root.go:59 uses fmt.Printf for errors; should use fmt.Fprintf(os.Stderr,...)
