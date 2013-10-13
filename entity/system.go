/*

type System interface {
	Update(e interface{})
	Valid(e interface{}) bool
}
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

package entity

import ()

type System interface {
	Update(e interface{})
	Valid(e interface{}) bool
}

var systems = []System{}

//Shouldn't be called after init
func RegisterSystem(sys System) {
	systems = append(systems, sys)
}

func Systems(e interface{}) []System {
	matches := make([]System, 0, 5)
	for _, sys := range systems {
		if sys.Valid(e) {
			matches = append(matches, sys)
		}
	}
	return matches
}
