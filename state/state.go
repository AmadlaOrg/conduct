package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// DeploymentState represents the state of a multi-node deployment.
type DeploymentState struct {
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	Nodes     []NodeState `json:"nodes"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

// NodeState represents the state of a single node within a deployment.
type NodeState struct {
	Name   string `json:"name"`
	Host   string `json:"host"`
	Status string `json:"status"`
}

// Manager defines the interface for managing deployment state persistence.
type Manager interface {
	Load(name string) (*DeploymentState, error)
	Save(state *DeploymentState) error
	Remove(name string) error
	List() ([]DeploymentState, error)
}

type manager struct {
	stateDir string
	mu       sync.Mutex
}

// New creates a new state manager with the default state directory.
func New() Manager {
	return &manager{stateDir: defaultStateDir()}
}

// NewWithPath creates a new state manager with a custom state directory.
func NewWithPath(dir string) Manager {
	return &manager{stateDir: dir}
}

func defaultStateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "conduct")
}

func (m *manager) Load(name string) (*DeploymentState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.statePath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no deployment state for %q", name)
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state DeploymentState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

func (m *manager) Save(state *DeploymentState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.stateDir, 0o755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	path := m.statePath(state.Name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (m *manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.statePath(name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no deployment state for %q", name)
		}
		return fmt.Errorf("failed to remove state file: %w", err)
	}

	return nil
}

func (m *manager) List() ([]DeploymentState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := os.ReadDir(m.stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []DeploymentState{}, nil
		}
		return nil, fmt.Errorf("failed to read state directory: %w", err)
	}

	var states []DeploymentState
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(m.stateDir, entry.Name()))
		if err != nil {
			continue
		}

		var state DeploymentState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}

		states = append(states, state)
	}

	return states, nil
}

func (m *manager) statePath(name string) string {
	return filepath.Join(m.stateDir, name+".json")
}
