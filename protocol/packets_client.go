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

type ChatMessage struct {
	Message string
}

func (ChatMessage) ID() byte { return 0x01 }

type Unknown struct {
	A int32
	B int8
}

func (Unknown) ID() byte { return 0x02 }

type ClientPlayer struct {
	OnGround bool
}

func (ClientPlayer) ID() byte { return 0x03 }

type ClientPlayerPosition struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	OnGround bool
}

func (ClientPlayerPosition) ID() byte { return 0x04 }

type ClientPlayerLook struct {
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (ClientPlayerLook) ID() byte { return 0x05 }

type ClientPlayerPositionLook struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (ClientPlayerPositionLook) ID() byte { return 0x06 }

type PlayerDigging struct {
	Status byte
	X      int32
	Y      byte
	Z      int32
	Face   byte
}

func (PlayerDigging) ID() byte { return 0x07 }

type PlayerBlockPlacement struct {
	X               int32
	Y               byte
	Z               int32
	Direction       int8
	HeldItem        Slot
	CursorPositionX int8
	CursorPositionY int8
	CursorPositionZ int8
}

func (PlayerBlockPlacement) ID() byte { return 0x08 }

type ClientHeldItemChange struct {
	SlotID int16
}

func (ClientHeldItemChange) ID() byte { return 0x09 }

type ClientAnimation struct {
	EntityID  int32
	Animation int8
}

func (ClientAnimation) ID() byte { return 0x0A }

type EntityAction struct {
	EntityID  int32
	ActionID  int8
	JumpBoost int32
}

func (EntityAction) ID() byte { return 0x0B }

type SteerVehicle struct {
	Sideways float32
	Forward  float32
	Jump     bool
	Unmount  bool
}

func (SteerVehicle) ID() byte { return 0x0C }

type ClientWindowClose struct {
	WindowID int8
}

func (ClientWindowClose) ID() byte { return 0x0D }

type WindowClick struct {
	WindowID     int8
	Slot         int16
	Button       int8
	ActionNumber int16
	Mode         int8
	Item         Slot
}

func (WindowClick) ID() byte { return 0x0E }

type ClientWindowTransactionConfirm struct {
	WindowID     int8
	ActionNumber int16
	Accepted     bool
}

func (ClientWindowTransactionConfirm) ID() byte { return 0x0F }

type CreativeInventoryAction struct {
	Slot int16
	Item Slot
}

func (CreativeInventoryAction) ID() byte { return 0x10 }

type EnchantItem struct {
	WindowID    int8
	Enchantment int8
}

func (EnchantItem) ID() byte { return 0x11 }

type ClientUpdateSign struct {
	X     int32
	Y     int16
	Z     int32
	Line1 string
	Line2 string
	Line3 string
	Line4 string
}

func (ClientUpdateSign) ID() byte { return 0x12 }

type ClientPlayerAbilities struct {
	Flags        byte
	FlyingSpeed  float32
	WalkingSpeed float32
}

func (ClientPlayerAbilities) ID() byte { return 0x13 }

type ClientTabComplete struct {
	Text string
}

func (ClientTabComplete) ID() byte { return 0x14 }

type ClientSettings struct {
	Locale       string
	ViewDistance int8
	ChatFlags    byte
	Difficulty   int8
	ShowCape     bool
}

func (ClientSettings) ID() byte { return 0x15 }

type ClientStatuses struct {
	Payload byte
}

func (ClientStatuses) ID() byte { return 0x16 }

type ClientPluginMessage struct {
	Channel string
	Data    []byte `ltype:"int16"`
}

func (ClientPluginMessage) ID() byte { return 0x17 }

/*MIA


type UseEntity struct {
	User        int32
	Target      int32
	MouseButton bool
}

func (UseEntity) ID() byte { return 0x07 }
*/
