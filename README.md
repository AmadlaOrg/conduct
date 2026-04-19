# conduct

Multi-server orchestrator for the Amadla ecosystem.

Coordinates tool execution (lay, enjoin, weaver, waiter) across multiple remote nodes via SSH, respecting dependency order and injecting cross-node data.

## Usage

```bash
# Deploy from topology file
conduct deploy -f topology.yaml

# Preview without executing
conduct deploy -f topology.yaml --dry-run

# Show deployment status
conduct status wordpress-cluster

# Run command on a specific node
conduct exec wordpress-cluster db-server -- systemctl status postgresql

# Remove deployment record
conduct destroy wordpress-cluster
```

## Topology Format

```yaml
name: wordpress-cluster
nodes:
  - name: db-server
    host: 10.0.0.1
    user: root
    key: ~/.ssh/id_rsa
    roles:
      - tool: lay
      - tool: enjoin

  - name: app-server
    host: 10.0.0.2
    user: root
    key: ~/.ssh/id_rsa
    depends_on:
      - db-server
    vars:
      db_host: "{{ db-server.host }}"
    roles:
      - tool: lay
      - tool: weaver
      - tool: waiter
```

## How It Works

1. Parses topology YAML/JSON defining nodes and their roles
2. Resolves node dependencies via topological sort
3. Interpolates cross-node variables (e.g., `{{ db-server.host }}` -> `10.0.0.1`)
4. Executes tools on each node via SSH in dependency order
5. Tracks deployment state for status/destroy commands

## License

MIT
