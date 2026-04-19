package topology

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Topology represents a multi-node deployment specification.
type Topology struct {
	Name  string `json:"name" yaml:"name"`
	Nodes []Node `json:"nodes" yaml:"nodes"`
}

// Node represents a single server in the topology.
type Node struct {
	Name      string            `json:"name" yaml:"name"`
	Host      string            `json:"host" yaml:"host"`
	User      string            `json:"user" yaml:"user"`
	Port      int               `json:"port" yaml:"port"`
	Key       string            `json:"key" yaml:"key"`
	DependsOn []string          `json:"depends_on" yaml:"depends_on"`
	Vars      map[string]string `json:"vars" yaml:"vars"`
	Roles     []Role            `json:"roles" yaml:"roles"`
}

// Role defines which tool and entity types to execute on a node.
type Role struct {
	Tool     string      `json:"tool" yaml:"tool"`
	Entities []EntityRef `json:"entities" yaml:"entities"`
	Args     []string    `json:"args" yaml:"args"`
}

// EntityRef references an entity type to process.
type EntityRef struct {
	Type string `json:"type" yaml:"type"`
	File string `json:"file" yaml:"file"`
}

// Parse parses topology data from JSON or YAML bytes.
func Parse(data []byte) (*Topology, error) {
	var t Topology
	if err := json.Unmarshal(data, &t); err != nil {
		if err := yaml.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("input is neither valid JSON nor YAML: %w", err)
		}
	}

	if err := validate(&t); err != nil {
		return nil, err
	}

	// Apply defaults
	for i := range t.Nodes {
		if t.Nodes[i].User == "" {
			t.Nodes[i].User = "root"
		}
		if t.Nodes[i].Port <= 0 {
			t.Nodes[i].Port = 22
		}
	}

	return &t, nil
}

func validate(t *Topology) error {
	if t.Name == "" {
		return fmt.Errorf("topology name is required")
	}
	if len(t.Nodes) == 0 {
		return fmt.Errorf("topology must have at least one node")
	}

	names := make(map[string]bool)
	for _, n := range t.Nodes {
		if n.Name == "" {
			return fmt.Errorf("node name is required")
		}
		if n.Host == "" {
			return fmt.Errorf("node %q: host is required", n.Name)
		}
		if names[n.Name] {
			return fmt.Errorf("duplicate node name: %q", n.Name)
		}
		names[n.Name] = true
	}

	// Validate depends_on references
	for _, n := range t.Nodes {
		for _, dep := range n.DependsOn {
			if !names[dep] {
				return fmt.Errorf("node %q depends on unknown node %q", n.Name, dep)
			}
			if dep == n.Name {
				return fmt.Errorf("node %q cannot depend on itself", n.Name)
			}
		}
	}

	return nil
}

// ResolveOrder returns nodes in dependency order (topological sort).
func (t *Topology) ResolveOrder() ([]Node, error) {
	nodeMap := make(map[string]*Node)
	for i := range t.Nodes {
		nodeMap[t.Nodes[i].Name] = &t.Nodes[i]
	}

	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var order []Node

	var visit func(name string) error
	visit = func(name string) error {
		if inStack[name] {
			return fmt.Errorf("circular dependency detected: %s", name)
		}
		if visited[name] {
			return nil
		}

		inStack[name] = true
		node := nodeMap[name]

		for _, dep := range node.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}

		inStack[name] = false
		visited[name] = true
		order = append(order, *node)
		return nil
	}

	for _, n := range t.Nodes {
		if err := visit(n.Name); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// GetNode returns a node by name.
func (t *Topology) GetNode(name string) (*Node, error) {
	for i := range t.Nodes {
		if t.Nodes[i].Name == name {
			return &t.Nodes[i], nil
		}
	}
	return nil, fmt.Errorf("node %q not found", name)
}
