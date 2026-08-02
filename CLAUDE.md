# CLAUDE.md

## AI Skills

Follow the practices defined in `~/Projects/SiteNetSoft/ai-skills/`:
- `dev-practices/golang/` — Go style, error handling, functions, testing, linting
- `dev-practices/git/` — Git authorship rules, multi-repo workspace patterns

## Project Overview

Conduct is the multi-server orchestrator in the Amadla ecosystem. It coordinates tool execution (lay, enjoin, weaver, waiter) across multiple remote nodes via SSH, handling dependency ordering and cross-node data injection.

## Build Commands

```bash
make build    # Build for current platform
make test     # Run tests
make clean    # Remove build artifacts
```

## Architecture

**Commands:**
```
conduct deploy -f topology.yaml         # Execute multi-node deployment
conduct deploy -f topology.yaml --dry-run  # Preview plan
conduct status [deployment]             # Show deployment status
conduct destroy <deployment>            # Remove deployment record
conduct exec <deployment> <node> -- <cmd>  # Run command on node
```

**Package Structure:**
- `main.go` - CLI entry point (Cobra)
- `topology/` - Topology parsing, validation, node ordering (topological sort)
- `executor/` - SSH-based remote command execution and SCP file copy
- `plan/` - Execution plan builder (topology -> ordered steps with variable interpolation)
- `state/` - Deployment state persistence (~/.local/share/conduct/)
- `cmd/` - CLI commands

**Cross-Node Data Flow:**
Variables in topology nodes support `{{ node-name.host }}` interpolation. Example: `db_host: "{{ db-server.host }}"` resolves to the db-server's IP. Injected as environment variables before tool execution on remote nodes.
