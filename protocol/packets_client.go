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

type ClientKeepAlive struct {
	KeepAliveID int32
}

type ChatMessage struct {
	Message string
}

type UseEntity struct {
	Target int32
	Mouse  int8
}

type ClientPlayer struct {
	OnGround bool
}

type ClientPlayerPosition struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	OnGround bool
}

type ClientPlayerLook struct {
	Yaw      float32
	Pitch    float32
	OnGround bool
}

type ClientPlayerPositionLook struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	Yaw      float32
	Pitch    float32
	OnGround bool
}

type PlayerDigging struct {
	Status byte
	X      int32
	Y      byte
	Z      int32
	Face   byte
}

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

type ClientHeldItemChange struct {
	SlotID int16
}

type ClientAnimation struct {
	EntityID  int32
	Animation int8
}

type EntityAction struct {
	EntityID  int32
	ActionID  int8
	JumpBoost int32
}

type SteerVehicle struct {
	Sideways float32
	Forward  float32
	Jump     bool
	Unmount  bool
}

type ClientWindowClose struct {
	WindowID int8
}

type WindowClick struct {
	WindowID     int8
	Slot         int16
	Button       int8
	ActionNumber int16
	Mode         int8
	Item         Slot
}

type ClientWindowTransactionConfirm struct {
	WindowID     int8
	ActionNumber int16
	Accepted     bool
}

type CreativeInventoryAction struct {
	Slot int16
	Item Slot
}

type EnchantItem struct {
	WindowID    int8
	Enchantment int8
}

type ClientUpdateSign struct {
	X     int32
	Y     int16
	Z     int32
	Line1 string
	Line2 string
	Line3 string
	Line4 string
}

type ClientPlayerAbilities struct {
	Flags        byte
	FlyingSpeed  float32
	WalkingSpeed float32
}

type ClientTabComplete struct {
	Text string
}

type ClientSettings struct {
	Locale       string
	ViewDistance int8
	ChatFlags    byte
	Difficulty   int8
	ShowCape     bool
}

type ClientStatuses struct {
	Payload byte
}

type ClientPluginMessage struct {
	Channel string
	Data    []byte `ltype:"int16"`
}
