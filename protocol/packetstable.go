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

package protocol

import (
	"reflect"
)

type State int

const (
	Handshaking State = iota
	Play
	Login
	Status
)

var packets = [4][]reflect.Type{
	0: []reflect.Type{
		0x00: reflect.TypeOf((*Handshake)(nil)).Elem(),
	},
	1: []reflect.Type{
		0x00: reflect.TypeOf((*KeepAlive)(nil)).Elem(),
		0x01: reflect.TypeOf((*ChatMessage)(nil)).Elem(),
		0x02: reflect.TypeOf((*Unknown)(nil)).Elem(),
		0x03: reflect.TypeOf((*ClientPlayer)(nil)).Elem(),
		0x04: reflect.TypeOf((*ClientPlayerPosition)(nil)).Elem(),
		0x05: reflect.TypeOf((*ClientPlayerLook)(nil)).Elem(),
		0x06: reflect.TypeOf((*ClientPlayerPositionLook)(nil)).Elem(),
		0x07: reflect.TypeOf((*PlayerDigging)(nil)).Elem(),
		0x08: reflect.TypeOf((*PlayerBlockPlacement)(nil)).Elem(),
		0x09: reflect.TypeOf((*ClientHeldItemChange)(nil)).Elem(),
		0x0A: reflect.TypeOf((*ClientAnimation)(nil)).Elem(),
		0x0B: reflect.TypeOf((*EntityAction)(nil)).Elem(),
		0x0C: reflect.TypeOf((*SteerVehicle)(nil)).Elem(),
		0x0D: reflect.TypeOf((*ClientWindowClose)(nil)).Elem(),
		0x0E: reflect.TypeOf((*WindowClick)(nil)).Elem(),
		0x0F: reflect.TypeOf((*ClientWindowTransactionConfirm)(nil)).Elem(),
		0x10: reflect.TypeOf((*CreativeInventoryAction)(nil)).Elem(),
		0x11: reflect.TypeOf((*EnchantItem)(nil)).Elem(),
		0x12: reflect.TypeOf((*ClientUpdateSign)(nil)).Elem(),
		0x13: reflect.TypeOf((*ClientPlayerAbilities)(nil)).Elem(),
		0x14: reflect.TypeOf((*ClientTabComplete)(nil)).Elem(),
		0x15: reflect.TypeOf((*ClientSettings)(nil)).Elem(),
		0x16: reflect.TypeOf((*ClientStatuses)(nil)).Elem(),
		0x17: reflect.TypeOf((*ClientPluginMessage)(nil)).Elem(),
	},
	2: []reflect.Type{
		0x00: reflect.TypeOf((*LoginStart)(nil)).Elem(),
		0x01: reflect.TypeOf((*EncryptionKeyResponse)(nil)).Elem(),
	},
	3: []reflect.Type{
		0x00: reflect.TypeOf((*StatusGet)(nil)).Elem(),
		0x01: reflect.TypeOf((*StatusPing)(nil)).Elem(),
	},
}
