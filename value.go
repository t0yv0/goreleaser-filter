package main

type value struct {
	inner    interface{}
	setInner func(interface{})
}

func toValue(x interface{}) *value {
	return &value{
		inner: x,
		setInner: func(interface{}) {
			panic("attempt to set read-only value")
		},
	}
}

func arrayValue(values ...*value) *value {
	x := []interface{}{}
	for _, v := range values {
		x = append(x, v.toInterface())
	}
	return toValue(x)
}

func (v *value) set(x *value) {
	v.setInner(x.inner)
}

func (v *value) toInterface() interface{} {
	return v.inner
}

func (v *value) at(key string) *value {
	if v == nil {
		return toValue(nil)
	}
	m, ok := v.toInterface().(map[interface{}]interface{})
	if !ok {
		return toValue(nil)
	}
	x, ok := m[key]
	if !ok {
		return toValue(nil)
	}
	return &value{
		inner: x,
		setInner: func(v interface{}) {
			m[key] = v
		},
	}
}

func (v *value) elements() []*value {
	var result []*value
	if slice, ok := v.toInterface().([]interface{}); ok {
		for i, s := range slice {
			result = append(result, &value{
				inner: s,
				setInner: func(v interface{}) {
					slice[i] = v
				},
			})
		}
	}
	return result
}

func (v *value) containsString(x string) bool {
	if v == nil {
		return false
	}
	elements, ok := v.toInterface().([]interface{})
	if !ok {
		return false
	}
	for _, e := range elements {
		if es, ok := e.(string); ok {
			if es == x {
				return true
			}
		}
	}
	return false
}
