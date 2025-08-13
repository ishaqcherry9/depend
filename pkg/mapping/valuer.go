package mapping

type (
	Valuer interface {
		Value(key string) (any, bool)
	}

	valuerWithParent interface {
		Valuer
		Parent() valuerWithParent
	}

	node struct {
		current Valuer
		parent  valuerWithParent
	}

	valueWithParent struct {
		value  any
		parent valuerWithParent
	}

	mapValuer       map[string]any
	simpleValuer    node
	recursiveValuer node
)

func (mv mapValuer) Value(key string) (any, bool) {
	v, ok := mv[key]
	return v, ok
}

func (sv simpleValuer) Value(key string) (any, bool) {
	v, ok := sv.current.Value(key)
	return v, ok
}

func (sv simpleValuer) Parent() valuerWithParent {
	if sv.parent == nil {
		return nil
	}

	return recursiveValuer{
		current: sv.parent,
		parent:  sv.parent.Parent(),
	}
}

func (rv recursiveValuer) Value(key string) (any, bool) {
	val, ok := rv.current.Value(key)
	if !ok {
		if parent := rv.Parent(); parent != nil {
			return parent.Value(key)
		}

		return nil, false
	}

	vm, ok := val.(map[string]any)
	if !ok {
		return val, true
	}

	parent := rv.Parent()
	if parent == nil {
		return val, true
	}

	pv, ok := parent.Value(key)
	if !ok {
		return val, true
	}

	pm, ok := pv.(map[string]any)
	if !ok {
		return val, true
	}

	for k, v := range pm {
		if _, ok := vm[k]; !ok {
			vm[k] = v
		}
	}

	return vm, true
}

func (rv recursiveValuer) Parent() valuerWithParent {
	if rv.parent == nil {
		return nil
	}

	return recursiveValuer{
		current: rv.parent,
		parent:  rv.parent.Parent(),
	}
}
