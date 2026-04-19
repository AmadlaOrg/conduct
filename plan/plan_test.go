package plan

import (
	"testing"

	"github.com/AmadlaOrg/conduct/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild_Simple(t *testing.T) {
	topo := &topology.Topology{
		Name: "test",
		Nodes: []topology.Node{
			{
				Name: "db",
				Host: "10.0.0.1",
				User: "root",
				Port: 22,
				Roles: []topology.Role{
					{Tool: "lay"},
					{Tool: "enjoin"},
				},
			},
		},
	}

	p, err := Build(topo)
	require.NoError(t, err)
	assert.Equal(t, "test", p.Name)
	assert.Len(t, p.Steps, 2)
	assert.Equal(t, "lay", p.Steps[0].Tool)
	assert.Equal(t, "enjoin", p.Steps[1].Tool)
	assert.Equal(t, 1, p.Steps[0].Order)
	assert.Equal(t, 2, p.Steps[1].Order)
}

func TestBuild_DependencyOrder(t *testing.T) {
	topo := &topology.Topology{
		Name: "test",
		Nodes: []topology.Node{
			{Name: "app", Host: "10.0.0.2", User: "root", Port: 22, DependsOn: []string{"db"}, Roles: []topology.Role{{Tool: "waiter"}}},
			{Name: "db", Host: "10.0.0.1", User: "root", Port: 22, Roles: []topology.Role{{Tool: "lay"}}},
		},
	}

	p, err := Build(topo)
	require.NoError(t, err)
	require.Len(t, p.Steps, 2)
	assert.Equal(t, "db", p.Steps[0].Node)
	assert.Equal(t, "app", p.Steps[1].Node)
}

func TestBuild_VarInterpolation(t *testing.T) {
	topo := &topology.Topology{
		Name: "test",
		Nodes: []topology.Node{
			{Name: "db", Host: "10.0.0.1", User: "root", Port: 22, Roles: []topology.Role{{Tool: "lay"}}},
			{
				Name:      "app",
				Host:      "10.0.0.2",
				User:      "root",
				Port:      22,
				DependsOn: []string{"db"},
				Vars:      map[string]string{"db_host": "{{ db.host }}"},
				Roles:     []topology.Role{{Tool: "waiter"}},
			},
		},
	}

	p, err := Build(topo)
	require.NoError(t, err)
	require.Len(t, p.Steps, 2)
	assert.Equal(t, "10.0.0.1", p.Steps[1].Vars["db_host"])
}

func TestBuild_CircularDep(t *testing.T) {
	topo := &topology.Topology{
		Name: "test",
		Nodes: []topology.Node{
			{Name: "a", Host: "1", User: "root", Port: 22, DependsOn: []string{"b"}},
			{Name: "b", Host: "2", User: "root", Port: 22, DependsOn: []string{"a"}},
		},
	}

	_, err := Build(topo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestBuild_NoRoles(t *testing.T) {
	topo := &topology.Topology{
		Name: "test",
		Nodes: []topology.Node{
			{Name: "empty", Host: "1.2.3.4", User: "root", Port: 22},
		},
	}

	p, err := Build(topo)
	require.NoError(t, err)
	assert.Empty(t, p.Steps)
}

func TestResolveVars_NoVars(t *testing.T) {
	result := resolveVars(nil, map[string]string{"db": "10.0.0.1"})
	assert.Nil(t, result)
}

func TestResolveVars_MultipleRefs(t *testing.T) {
	vars := map[string]string{
		"db_host":    "{{ db.host }}",
		"cache_host": "{{ cache.host }}",
		"static":     "literal",
	}
	hosts := map[string]string{"db": "10.0.0.1", "cache": "10.0.0.3"}

	result := resolveVars(vars, hosts)
	assert.Equal(t, "10.0.0.1", result["db_host"])
	assert.Equal(t, "10.0.0.3", result["cache_host"])
	assert.Equal(t, "literal", result["static"])
}
