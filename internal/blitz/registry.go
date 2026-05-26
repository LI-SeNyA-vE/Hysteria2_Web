package blitz

import (
	"fmt"
	"sync"

	"hysteria2-web/internal/domain/server"
)

type Registry struct {
	mu      sync.RWMutex
	clients map[uint]*HTTPClient
}

func NewRegistry() *Registry {
	return &Registry{
		clients: make(map[uint]*HTTPClient),
	}
}

func (r *Registry) Register(s server.Server) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[s.ID] = NewClient(s.BaseURL, s.APIKey)
}

func (r *Registry) Unregister(id uint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, id)
}

func (r *Registry) Get(id uint) (*HTTPClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	client, ok := r.clients[id]
	if !ok {
		return nil, fmt.Errorf("blitz: server %d not registered", id)
	}
	return client, nil
}

func (r *Registry) LoadAll(servers []server.Server) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients = make(map[uint]*HTTPClient, len(servers))
	for _, s := range servers {
		if s.IsActive {
			r.clients[s.ID] = NewClient(s.BaseURL, s.APIKey)
		}
	}
}
