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

package nbt

import (
	"errors"
	"io"
)

func skip(r io.Reader, de *msgDecoder) error {
	_, err := fallbackRead(r, de)
	return err
}

//If the type isn't known at read time then this will manually check the type
//and return the value
func fallbackRead(r io.Reader, de *msgDecoder) (interface{}, error) {
	name, t, err := readPrefix(r, de)
	_, _, _ = name, t, err
	panic("NYI")
	return nil, ErrorUnknownType
}

var ErrorUnknownType = errors.New("format/msgpack: Unknown type")
