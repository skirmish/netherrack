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
	"fmt"
)

type ErrorIncorrectType struct {
	Wanted string
	Got    byte
}

func (e ErrorIncorrectType) Error() string {
	return fmt.Sprintf("Wanted %s but instead got %02X", e.Wanted, e.Got)
}

type ErrorIncorrectLength struct {
	Wanted int
	Got    int
}

func (e ErrorIncorrectLength) Error() string {
	return fmt.Sprintf("Wanted array length of %d but instead got %d", e.Wanted, e.Got)
}

type ErrorUnknownInterface struct {
	Id string
}

func (e ErrorUnknownInterface) Error() string {
	return fmt.Sprintf("Unknown type %s", e.Id)
}
