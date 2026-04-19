package plan

import (
	"fmt"
	"strings"

	"github.com/AmadlaOrg/conduct/topology"
)

// Step represents a single execution step in the deployment plan.
type Step struct {
	Order int               `json:"order"`
	Node  string            `json:"node"`
	Host  string            `json:"host"`
	User  string            `json:"user"`
	Port  int               `json:"port"`
	Key   string            `json:"key"`
	Tool  string            `json:"tool"`
	Args  []string          `json:"args,omitempty"`
	Vars  map[string]string `json:"vars,omitempty"`
}

// Plan represents an ordered list of execution steps.
type Plan struct {
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}

// Build creates an execution plan from a topology.
func Build(topo *topology.Topology) (*Plan, error) {
	orderedNodes, err := topo.ResolveOrder()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve node order: %w", err)
	}

	p := &Plan{Name: topo.Name}
	stepOrder := 1

	// Build a map of node name -> host for variable interpolation
	nodeHosts := make(map[string]string)
	for _, n := range orderedNodes {
		nodeHosts[n.Name] = n.Host
	}

	for _, node := range orderedNodes {
		// Resolve variable interpolation
		resolvedVars := resolveVars(node.Vars, nodeHosts)

		for _, role := range node.Roles {
			step := Step{
				Order: stepOrder,
				Node:  node.Name,
				Host:  node.Host,
				User:  node.User,
				Port:  node.Port,
				Key:   node.Key,
				Tool:  role.Tool,
				Args:  role.Args,
				Vars:  resolvedVars,
			}
			p.Steps = append(p.Steps, step)
			stepOrder++
		}
	}

	return p, nil
}

// resolveVars interpolates node host references in variable values.
// Supports {{ node-name.host }} syntax.
func resolveVars(vars map[string]string, nodeHosts map[string]string) map[string]string {
	if len(vars) == 0 {
		return nil
	}

	resolved := make(map[string]string)
	for k, v := range vars {
		for nodeName, host := range nodeHosts {
			placeholder := fmt.Sprintf("{{ %s.host }}", nodeName)
			v = strings.ReplaceAll(v, placeholder, host)
		}
		resolved[k] = v
	}
	return resolved
}
