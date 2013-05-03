package nbt

import ()

type Type map[string]interface{}

//Creates a new NBT compound
func NewNBT() Type {
	return make(Type)
}

func (t Type) Set(name string, val interface{}) {
	t[name] = val
}

func (t Type) Remove(name string) {
	delete(t, name)
}

func (t Type) GetByte(name string, def int8) (v int8, ok bool) {
	val, ok := t[name]
	if !ok {
		v = def
		return
	}
	v = val.(int8)
	return
}

func (t Type) GetShort(name string, def int16) (v int16, ok bool) {
	val, ok := t[name]
	if !ok {
		v = def
		return
	}
	v = val.(int16)
	return
}

func (t Type) GetInt(name string, def int32) (v int32, ok bool) {
	val, ok := t[name]
	if !ok {
		v = def
		return
	}
	v = val.(int32)
	return
}

func (t Type) GetLong(name string, def int64) (v int64, ok bool) {
	val, ok := t[name]
	if !ok {
		v = def
		return
	}
	v = val.(int64)
	return
}

func (t Type) GetFloat(name string, def float32) (v float32, ok bool) {
	val, ok := t[name]
	if !ok {
		v = def
		return
	}
	v = val.(float32)
	return
}

func (t Type) GetDouble(name string, def float64) (v float64, ok bool) {
	val, ok := t[name]
	if !ok {
		v = def
		return
	}
	v = val.(float64)
	return
}

func (t Type) GetByteArray(name string, def []byte) (v []byte, ok bool) {
	val, ok := t[name]
	if !ok {
		v = def
		return
	}
	v = val.([]byte)
	return
}

func (t Type) GetString(name string, def string) (v string, ok bool) {
	val, ok := t[name]
	if !ok {
		v = def
		return
	}
	v = val.(string)
	return
}

func (t Type) GetList(name string, create bool) (v []interface{}, ok bool) {
	val, ok := t[name]
	if !ok {
		if create {
			v = make([]interface{}, 0)
			t[name] = v
		}
		return
	}
	v = val.([]interface{})
	return
}

func (t Type) GetCompound(name string, create bool) (v Type, ok bool) {
	val, ok := t[name]
	if !ok {
		if create {
			v = NewNBT()
			t[name] = v
		}
		return
	}
	v = val.(Type)
	return
}

func (t Type) GetIntArray(name string, def []int32) (v []int32, ok bool) {
	val, ok := t[name]
	if !ok {
		v = def
		return
	}
	v = val.([]int32)
	return
}
