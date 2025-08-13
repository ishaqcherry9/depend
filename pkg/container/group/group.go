package group

import "sync"

type Group struct {
	new  func() interface{}
	vals map[string]interface{}
	sync.RWMutex
}

func NewGroup(fn func() interface{}) *Group {
	if fn == nil {
		panic("container.group: can't assign a nil to the new function")
	}
	return &Group{
		new:  fn,
		vals: make(map[string]interface{}),
	}
}

func (g *Group) Get(key string) interface{} {
	g.RLock()
	v, ok := g.vals[key]
	if ok {
		g.RUnlock()
		return v
	}
	g.RUnlock()

	g.Lock()
	defer g.Unlock()
	v, ok = g.vals[key]
	if ok {
		return v
	}
	v = g.new()
	g.vals[key] = v
	return v
}

func (g *Group) Reset(fn func() interface{}) {
	if fn == nil {
		panic("container.group: can't assign a nil to the new function")
	}
	g.Lock()
	g.new = fn
	g.Unlock()
	g.Clear()
}

func (g *Group) Clear() {
	g.Lock()
	g.vals = make(map[string]interface{})
	g.Unlock()
}
