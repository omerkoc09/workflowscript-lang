package eval

import "fmt"

type Env struct {
	store  map[string]interface{}
	parent *Env
}

func NewEnv(parent *Env) *Env {
	return &Env{store: map[string]interface{}{}, parent: parent}
}

func (e *Env) Set(name string, val interface{}) {
	e.store[name] = val
}

func (e *Env) Get(name string) (interface{}, error) {
	if v, ok := e.store[name]; ok {
		return v, nil
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, fmt.Errorf("undeclared variable %q", name)
}

func (e *Env) Assign(name string, val interface{}) error {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return nil
	}
	if e.parent != nil {
		return e.parent.Assign(name, val)
	}
	return fmt.Errorf("undeclared variable %q", name)
}
