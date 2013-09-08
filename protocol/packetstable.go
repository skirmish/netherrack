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

var packets = [256]reflect.Type{
	0x00: reflect.TypeOf((*KeepAlive)(nil)).Elem(),
	0x01: reflect.TypeOf((*LoginRequest)(nil)).Elem(),
	0x02: reflect.TypeOf((*Handshake)(nil)).Elem(),
	0x03: reflect.TypeOf((*ChatMessage)(nil)).Elem(),
	0x04: reflect.TypeOf((*TimeUpdate)(nil)).Elem(),
	0x05: reflect.TypeOf((*EntityEquipment)(nil)).Elem(),
	0x06: reflect.TypeOf((*SpawnPosition)(nil)).Elem(),
	0x07: reflect.TypeOf((*UseEntity)(nil)).Elem(),
	0x08: reflect.TypeOf((*UpdateHealth)(nil)).Elem(),
	0x09: reflect.TypeOf((*Respawn)(nil)).Elem(),
	0x0A: reflect.TypeOf((*Player)(nil)).Elem(),
	0x0B: reflect.TypeOf((*PlayerPosition)(nil)).Elem(),
	0x0C: reflect.TypeOf((*PlayerLook)(nil)).Elem(),
	0x0D: reflect.TypeOf((*PlayerPositionLook)(nil)).Elem(),
	0x0E: reflect.TypeOf((*PlayerDigging)(nil)).Elem(),
	0x0F: reflect.TypeOf((*PlayerBlockPlacement)(nil)).Elem(),
	0x10: reflect.TypeOf((*HeldItemChange)(nil)).Elem(),
	0x11: reflect.TypeOf((*UseBed)(nil)).Elem(),
	0x12: reflect.TypeOf((*Animation)(nil)).Elem(),
	0x13: reflect.TypeOf((*EntityAction)(nil)).Elem(),
	0x14: reflect.TypeOf((*SpawnNamedEntity)(nil)).Elem(),

	0x16: reflect.TypeOf((*CollectItem)(nil)).Elem(),
	0x17: reflect.TypeOf((*SpawnObject)(nil)).Elem(),
	0x18: reflect.TypeOf((*SpawnMob)(nil)).Elem(),
	0x19: reflect.TypeOf((*SpawnPainting)(nil)).Elem(),
	0x1A: reflect.TypeOf((*SpawnExperienceOrb)(nil)).Elem(),
	0x1B: reflect.TypeOf((*SteerVehicle)(nil)).Elem(),
	0x1C: reflect.TypeOf((*EntityVelocity)(nil)).Elem(),
	0x1D: reflect.TypeOf((*EntityDestroy)(nil)).Elem(),
	0x1E: reflect.TypeOf((*Entity)(nil)).Elem(),
	0x1F: reflect.TypeOf((*EntityMove)(nil)).Elem(),
	0x20: reflect.TypeOf((*EntityLook)(nil)).Elem(),
	0x21: reflect.TypeOf((*EntityLookMove)(nil)).Elem(),
	0x22: reflect.TypeOf((*EntityTeleport)(nil)).Elem(),
	0x23: reflect.TypeOf((*EntityHeadLook)(nil)).Elem(),

	0x26: reflect.TypeOf((*EntityStatus)(nil)).Elem(),
	0x27: reflect.TypeOf((*EntityAttach)(nil)).Elem(),
	0x28: reflect.TypeOf((*EntityMetadata)(nil)).Elem(),
	0x29: reflect.TypeOf((*EntityEffect)(nil)).Elem(),
	0x2A: reflect.TypeOf((*EntityEffectRemove)(nil)).Elem(),
	0x2B: reflect.TypeOf((*SetExperience)(nil)).Elem(),
	0x2C: reflect.TypeOf((*EntityProperties)(nil)).Elem(),

	0x33: reflect.TypeOf((*ChunkData)(nil)).Elem(),
	0x34: reflect.TypeOf((*MultiBlockChange)(nil)).Elem(),
	0x35: reflect.TypeOf((*BlockChange)(nil)).Elem(),
	0x36: reflect.TypeOf((*BlockAction)(nil)).Elem(),
	0x37: reflect.TypeOf((*BlockBreakAnimation)(nil)).Elem(),

	0x3C: reflect.TypeOf((*Explosion)(nil)).Elem(),
	0x3D: reflect.TypeOf((*Effect)(nil)).Elem(),
	0x3E: reflect.TypeOf((*SoundEffect)(nil)).Elem(),
	0x3F: reflect.TypeOf((*Particle)(nil)).Elem(),

	0x46: reflect.TypeOf((*GameState)(nil)).Elem(),
	0x47: reflect.TypeOf((*SpawnGlobalEntity)(nil)).Elem(),

	0x64: reflect.TypeOf((*WindowOpen)(nil)).Elem(),
	0x65: reflect.TypeOf((*WindowClose)(nil)).Elem(),
	0x66: reflect.TypeOf((*WindowClick)(nil)).Elem(),
	0x67: reflect.TypeOf((*WindowSetSlot)(nil)).Elem(),
	0x68: reflect.TypeOf((*WindowSetSlots)(nil)).Elem(),
	0x69: reflect.TypeOf((*WindowUpdateProperty)(nil)).Elem(),
	0x6A: reflect.TypeOf((*WindowTransactionConfirm)(nil)).Elem(),
	0x6B: reflect.TypeOf((*CreativeInventoryAction)(nil)).Elem(),
	0x6C: reflect.TypeOf((*EnchantItem)(nil)).Elem(),

	0x82: reflect.TypeOf((*UpdateSign)(nil)).Elem(),
	0x83: reflect.TypeOf((*ItemData)(nil)).Elem(),
	0x84: reflect.TypeOf((*UpdateTileEntity)(nil)).Elem(),
	0x85: reflect.TypeOf((*TileEditorOpen)(nil)).Elem(),

	0xC8: reflect.TypeOf((*IncrementStatistic)(nil)).Elem(),
	0xC9: reflect.TypeOf((*PlayerListItem)(nil)).Elem(),
	0xCA: reflect.TypeOf((*PlayerAbilities)(nil)).Elem(),
	0xCB: reflect.TypeOf((*TabComplete)(nil)).Elem(),
	0xCC: reflect.TypeOf((*ClientSettings)(nil)).Elem(),
	0xCD: reflect.TypeOf((*ClientStatuses)(nil)).Elem(),
	0xCE: reflect.TypeOf((*ScoreboardObjective)(nil)).Elem(),
	0xCF: reflect.TypeOf((*UpdateScore)(nil)).Elem(),
	0xD0: reflect.TypeOf((*DisplayScoreboard)(nil)).Elem(),
	0xD1: reflect.TypeOf((*Teams)(nil)).Elem(),

	0xFA: reflect.TypeOf((*PluginMessage)(nil)).Elem(),

	0xFC: reflect.TypeOf((*EncryptionKeyResponse)(nil)).Elem(),
	0xFD: reflect.TypeOf((*EncryptionKeyRequest)(nil)).Elem(),
	0xFE: reflect.TypeOf((*ServerListPing)(nil)).Elem(),
	0xFF: reflect.TypeOf((*Disconnect)(nil)).Elem(),
}
