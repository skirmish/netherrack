package metadata

import (
	"fmt"
	"github.com/NetherrackDev/netherrack/nbt"
	"github.com/NetherrackDev/soulsand"
	"reflect"
	"sync"
)

var metadataValueType = reflect.TypeOf((*soulsand.MetadataValue)(nil)).Elem()
var metadataValues = map[string]reflect.Type{}

func RegisterType(t reflect.Type) {
	rt := t
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if !rt.AssignableTo(metadataValueType) {
		panic(fmt.Errorf("%s is not a soulsand.MetadataValue", t.Name()))
	}
	metadataValues[t.Name()] = rt
}

type Storage struct {
	data map[string]interface{}
	sync.RWMutex
}

func (s *Storage) SetMetadata(key string, value interface{}) {
	s.Lock()
	defer s.Unlock()
	if s.data == nil {
		s.data = map[string]interface{}{}
	}
	s.data[key] = value
}

func (s *Storage) Metadata(key string) interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.data[key]
}

func (s *Storage) RemoveMetadata(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.data, key)
}

func (s *Storage) ToNBT() nbt.Type {
	s.RLock()
	defer s.RUnlock()
	out := nbt.NewNBT()
	for key, value := range s.data {
		if metadataValue, ok := value.(soulsand.MetadataValue); ok {
			out.Set(key, structToNBT(metadataValue))
			continue
		}
		out.Set(key, value)
	}
	return out
}

func structToNBT(value interface{}) nbt.Type {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	if _, ok := metadataValues[t.Name()]; !ok {
		panic(fmt.Errorf("%s hasn't been registered", t.Name()))
	}
	out := nbt.NewNBT()
	out.Set("$netherrackType", t.PkgPath()+":"+t.Name())
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64, reflect.String:
			out.Set(f.Type().Name(), f.Interface())
		case reflect.Struct, reflect.Ptr:
			out.Set(f.Type().Name(), structToNBT(f.Interface()))
		}
	}
	return out
}
