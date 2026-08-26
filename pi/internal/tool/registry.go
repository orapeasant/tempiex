package tool

import "sync"

type Registry struct {
	mu    sync.RWMutex
	tools map[string]AnyTool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]AnyTool)}
}

func (r *Registry) Register(t AnyTool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[t.Info().Name] = t
}

func (r *Registry) Get(name string) (AnyTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) List() []ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	infos := make([]ToolInfo, 0, len(r.tools))
	for _, t := range r.tools {
		infos = append(infos, t.Info())
	}
	return infos
}
