package topology

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_YAML(t *testing.T) {
	data := []byte(`
name: test-cluster
nodes:
  - name: db
    host: 10.0.0.1
    roles:
      - tool: lay
        entities:
          - type: amadla.org/entity/package@v1.0.0
  - name: app
    host: 10.0.0.2
    depends_on:
      - db
    roles:
      - tool: waiter
`)

	topo, err := Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "test-cluster", topo.Name)
	assert.Len(t, topo.Nodes, 2)
	assert.Equal(t, "db", topo.Nodes[0].Name)
	assert.Equal(t, "root", topo.Nodes[0].User) // default
	assert.Equal(t, 22, topo.Nodes[0].Port)     // default
	assert.Equal(t, "app", topo.Nodes[1].Name)
	assert.Equal(t, []string{"db"}, topo.Nodes[1].DependsOn)
}

func TestParse_JSON(t *testing.T) {
	data := []byte(`{"name":"test","nodes":[{"name":"web","host":"1.2.3.4"}]}`)
	topo, err := Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "test", topo.Name)
	assert.Len(t, topo.Nodes, 1)
}

func TestParse_MissingName(t *testing.T) {
	data := []byte(`nodes: [{name: web, host: 1.2.3.4}]`)
	_, err := Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "topology name is required")
}

func TestParse_NoNodes(t *testing.T) {
	data := []byte(`name: test`)
	_, err := Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one node")
}

func TestParse_MissingHost(t *testing.T) {
	data := []byte(`name: test
nodes:
  - name: web`)
	_, err := Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host is required")
}

func TestParse_DuplicateNode(t *testing.T) {
	data := []byte(`name: test
nodes:
  - name: web
    host: 1.2.3.4
  - name: web
    host: 5.6.7.8`)
	_, err := Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate node name")
}

func TestParse_UnknownDependency(t *testing.T) {
	data := []byte(`name: test
nodes:
  - name: app
    host: 1.2.3.4
    depends_on: [missing]`)
	_, err := Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown node")
}

func TestParse_SelfDependency(t *testing.T) {
	data := []byte(`name: test
nodes:
  - name: app
    host: 1.2.3.4
    depends_on: [app]`)
	_, err := Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot depend on itself")
}

func TestParse_InvalidInput(t *testing.T) {
	_, err := Parse([]byte("{invalid json"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "neither valid JSON nor YAML")
}

func TestParse_EmptyInput(t *testing.T) {
	_, err := Parse([]byte(":::"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "topology name is required")
}

func TestResolveOrder_Simple(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "app", Host: "2", DependsOn: []string{"db"}},
			{Name: "db", Host: "1"},
		},
	}

	order, err := topo.ResolveOrder()
	require.NoError(t, err)
	require.Len(t, order, 2)
	assert.Equal(t, "db", order[0].Name)
	assert.Equal(t, "app", order[1].Name)
}

func TestResolveOrder_NoDeps(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "a", Host: "1"},
			{Name: "b", Host: "2"},
		},
	}

	order, err := topo.ResolveOrder()
	require.NoError(t, err)
	assert.Len(t, order, 2)
}

func TestResolveOrder_Chain(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "c", Host: "3", DependsOn: []string{"b"}},
			{Name: "b", Host: "2", DependsOn: []string{"a"}},
			{Name: "a", Host: "1"},
		},
	}

	order, err := topo.ResolveOrder()
	require.NoError(t, err)
	require.Len(t, order, 3)
	assert.Equal(t, "a", order[0].Name)
	assert.Equal(t, "b", order[1].Name)
	assert.Equal(t, "c", order[2].Name)
}

func TestResolveOrder_CircularDep(t *testing.T) {
	topo := &Topology{
		Name: "test",
		Nodes: []Node{
			{Name: "a", Host: "1", DependsOn: []string{"b"}},
			{Name: "b", Host: "2", DependsOn: []string{"a"}},
		},
	}

	_, err := topo.ResolveOrder()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestGetNode(t *testing.T) {
	topo := &Topology{
		Name:  "test",
		Nodes: []Node{{Name: "web", Host: "1.2.3.4"}},
	}

	node, err := topo.GetNode("web")
	require.NoError(t, err)
	assert.Equal(t, "1.2.3.4", node.Host)
}

func TestGetNode_NotFound(t *testing.T) {
	topo := &Topology{Name: "test", Nodes: []Node{}}
	_, err := topo.GetNode("missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
