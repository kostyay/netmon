---
id: net-3e94
status: closed
deps:
- net-3412
created: "2026-01-15T00:25:13Z"
type: task
priority: 2
assignee: kostyay
parent: net-5bad
tests_passed: false
---
# Implement TTY detection in root PreRunE

Check isatty.IsTerminal(os.Stdout.Fd()); if no TTY or --json, run JSON mode
