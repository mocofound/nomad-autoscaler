package apm

import (
	"fmt"
	"sync"

	"github.com/hashicorp/go-plugin"
)

// APM interface plugins must implement
type APM interface {
	Query(q string) (float64, error)
	SetConfig(config map[string]string) error
}

type Manager struct {
	lock            sync.RWMutex
	lockInternal    sync.RWMutex
	pluginClients   map[string]*plugin.Client
	internalPlugins map[string]*APM
}

func NewAPMManager() *Manager {
	return &Manager{
		pluginClients:   make(map[string]*plugin.Client),
		internalPlugins: make(map[string]*APM),
	}
}

func (m *Manager) RegisterPlugin(key string, p *plugin.ClientConfig) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	client := plugin.NewClient(p)
	m.pluginClients[key] = client
	return nil
}

func (m *Manager) RegisterInternalPlugin(key string, p *APM) {
	m.lockInternal.Lock()
	defer m.lockInternal.Unlock()

	m.internalPlugins[key] = p
}

func (m *Manager) Dispense(key string) (*APM, error) {
	// check if this is a local implementation
	m.lockInternal.RLock()
	if apm, ok := m.internalPlugins[key]; ok {
		m.lockInternal.RUnlock()
		return apm, nil
	}
	m.lockInternal.RUnlock()

	// otherwise dispense a plugin
	m.lock.RLock()
	client := m.pluginClients[key]
	m.lock.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("missing client %s", key)
	}

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %v", err)
	}

	raw, err := rpcClient.Dispense("apm")
	if err != nil {
		return nil, fmt.Errorf("failed to dispense plugin: %v", err)
	}
	apm, ok := raw.(APM)
	if !ok {
		return nil, fmt.Errorf("plugins %s is not APM\n", key)
	}

	return &apm, nil
}

func (m *Manager) Kill() {
	for _, c := range m.pluginClients {
		c.Kill()
	}
}
