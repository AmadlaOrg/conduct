package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Load_NotFound(t *testing.T) {
	mgr := NewWithPath(t.TempDir())
	_, err := mgr.Load("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no deployment state")
}

func TestManager_SaveAndLoad(t *testing.T) {
	mgr := NewWithPath(t.TempDir())

	state := &DeploymentState{
		Name:   "test-cluster",
		Status: "deployed",
		Nodes: []NodeState{
			{Name: "db", Host: "10.0.0.1", Status: "ready"},
			{Name: "app", Host: "10.0.0.2", Status: "ready"},
		},
		CreatedAt: "2026-03-21T00:00:00Z",
		UpdatedAt: "2026-03-21T00:00:00Z",
	}

	require.NoError(t, mgr.Save(state))

	loaded, err := mgr.Load("test-cluster")
	require.NoError(t, err)
	assert.Equal(t, "test-cluster", loaded.Name)
	assert.Equal(t, "deployed", loaded.Status)
	assert.Len(t, loaded.Nodes, 2)
}

func TestManager_Save_Overwrite(t *testing.T) {
	mgr := NewWithPath(t.TempDir())

	state := &DeploymentState{Name: "test", Status: "deploying", CreatedAt: "2026-03-21T00:00:00Z", UpdatedAt: "2026-03-21T00:00:00Z"}
	require.NoError(t, mgr.Save(state))

	state.Status = "deployed"
	require.NoError(t, mgr.Save(state))

	loaded, err := mgr.Load("test")
	require.NoError(t, err)
	assert.Equal(t, "deployed", loaded.Status)
}

func TestManager_Remove_Success(t *testing.T) {
	dir := t.TempDir()
	mgr := NewWithPath(dir)

	state := &DeploymentState{Name: "test", Status: "deployed", CreatedAt: "2026-03-21T00:00:00Z", UpdatedAt: "2026-03-21T00:00:00Z"}
	require.NoError(t, mgr.Save(state))

	require.NoError(t, mgr.Remove("test"))

	_, err := mgr.Load("test")
	assert.Error(t, err)
}

func TestManager_Remove_NotFound(t *testing.T) {
	mgr := NewWithPath(t.TempDir())
	err := mgr.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no deployment state")
}

func TestManager_List_Empty(t *testing.T) {
	mgr := NewWithPath(t.TempDir())
	states, err := mgr.List()
	require.NoError(t, err)
	assert.Empty(t, states)
}

func TestManager_List_Multiple(t *testing.T) {
	dir := t.TempDir()
	mgr := NewWithPath(dir)

	require.NoError(t, mgr.Save(&DeploymentState{Name: "cluster-a", Status: "deployed", CreatedAt: "2026-03-21T00:00:00Z", UpdatedAt: "2026-03-21T00:00:00Z"}))
	require.NoError(t, mgr.Save(&DeploymentState{Name: "cluster-b", Status: "deployed", CreatedAt: "2026-03-21T01:00:00Z", UpdatedAt: "2026-03-21T01:00:00Z"}))

	states, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, states, 2)
}

func TestManager_List_NoDirectory(t *testing.T) {
	mgr := NewWithPath(filepath.Join(t.TempDir(), "nonexistent"))
	states, err := mgr.List()
	require.NoError(t, err)
	assert.Empty(t, states)
}

func TestManager_Load_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bad.json"), []byte("not json"), 0o644))

	mgr := NewWithPath(dir)
	_, err := mgr.Load("bad")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}
