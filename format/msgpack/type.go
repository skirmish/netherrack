/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package msgpack

import (
	"reflect"
	"sync"
	"unsafe"
)

var (
	engines     = map[reflect.Type]*engine{}
	enginesById = map[string]*engine{}
	engineLock  sync.RWMutex
)

//Registers the type for use in decoding interface types
func Register(e interface{}) {
	getEngine(reflect.TypeOf(e))
}

type encodeFunc func(enc *Encoder, p unsafe.Pointer) error
type encodeReflectFunc func(enc *Encoder, v reflect.Value) error
type decodeFunc func(dec *Decoder, p unsafe.Pointer) error
type decodeReflectFunc func(dec *Decoder, v reflect.Value) error

func init() {
	//Special case
	enginesById["[]byte"] = &engine{
		id:     "[]byte",
		_type:  reflect.TypeOf((*[]byte)(nil)).Elem(),
		encode: encodeFunc(_encodeByteSlice),
		decode: decodeFunc(_decodeByteSlice),
	}
	engines[reflect.TypeOf((*[]byte)(nil)).Elem()] = enginesById["[]byte"]
}

type engine struct {
	id    string
	_type reflect.Type

	encode interface{}
	decode interface{}
}

func getEngine(t reflect.Type) (eng *engine) {
	engineLock.RLock()
	eng = engines[t]
	engineLock.RUnlock()
	if eng != nil {
		return
	}

	engineLock.Lock()
	defer engineLock.Unlock()

	//Check again in case another goroutine just compiled it
	eng = engines[t]
	if eng != nil {
		return
	}

	eng = compileType(t)
	engines[t] = eng
	enginesById[eng.id] = eng

	return
}

func getEngineById(id string) (eng *engine) {
	engineLock.RLock()
	eng = enginesById[id]
	defer engineLock.RUnlock()
	return
}

func compileType(t reflect.Type) *engine {
	eng := &engine{
		_type: t,
	}
	if t.Kind() == reflect.Struct {
		eng.id = t.PkgPath() + "#" + t.Name()
	} else {
		eng.id = t.Kind().String()
	}

	eng.encode = getTypeEncoder(t)
	eng.decode = getTypeDecoder(t)

	return eng
}
