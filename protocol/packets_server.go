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

import ()

type Packet interface {
	ID() byte
}

//Contains infomation on an item
type Slot struct {
	ID     int16
	Count  int8   `if:"ID,!=,-1"`
	Damage int16  `if:"ID,!=,-1"`
	Tag    []byte `if:"ID,!=,-1" nil:"-1" ltype:"int16"`
}

type KeepAlive struct {
	KeepAliveID int32
}

func (KeepAlive) ID() byte { return 0x00 }

type JoinGame struct {
	EntityID   int32
	LevelType  string
	Gamemode   int8
	Dimension  int8
	Difficulty int8
	NotUsed    int8
	MaxPlayers int8
}

func (JoinGame) ID() byte { return 0x01 }

type ServerMessage struct {
	Message string
}

func (ServerMessage) ID() byte { return 0x02 }

type TimeUpdate struct {
	AgeOfTheWorld int64
	TimeOfDay     int64
}

func (TimeUpdate) ID() byte { return 0x03 }

type EntityEquipment struct {
	EntityID int32
	Slot     int16
	Item     Slot
}

func (EntityEquipment) ID() byte { return 0x04 }

type SpawnPosition struct {
	X int32
	Y int32
	Z int32
}

func (SpawnPosition) ID() byte { return 0x05 }

type UpdateHealth struct {
	Health         float32
	Food           int16
	FoodSaturation float32
}

func (UpdateHealth) ID() byte { return 0x06 }

type Respawn struct {
	Dimension   int32
	Difficulty  int8
	Gamemode    int8
	WorldHeight int16
	LevelType   string
}

func (Respawn) ID() byte { return 0x07 }

type Player struct {
	OnGround bool
}

func (Player) ID() byte { return 0x08 }

type PlayerPosition struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	OnGround bool
}

func (PlayerPosition) ID() byte { return 0x09 }

type PlayerLook struct {
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (PlayerLook) ID() byte { return 0x0A }

type PlayerPositionLook struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (PlayerPositionLook) ID() byte { return 0x0B }

type HeldItemChange struct {
	SlotID int16
}

func (HeldItemChange) ID() byte { return 0x0C }

type UseBed struct {
	EntityID int32
	unknown  byte
	X        int32
	Y        byte
	Z        int32
}

func (UseBed) ID() byte { return 0x0D }

type Animation struct {
	EntityID  int32
	Animation int8
}

func (Animation) ID() byte { return 0x0E }

type SpawnPlayer struct {
	EntityID    int32
	PlayerUUID  string
	PlayerName  string
	X           int32
	Y           int32
	Z           int32
	Yaw         int8
	Pitch       int8
	CurrentItem int16
	Metadata    map[byte]interface{} `metadata:"true"`
}

func (SpawnPlayer) ID() byte { return 0x0F }

type CollectItem struct {
	CollectedEntityID int32
	CollectorEntityID int32
}

func (CollectItem) ID() byte { return 0x10 }

type SpawnObject struct {
	EntityID  int32
	Type      int8
	X         int32
	Y         int32
	Z         int32
	Pitch     int8
	Yaw       int8
	ExtraData int32
	SpeedX    int16 `if:"ExtraData,!=,0"`
	SpeedY    int16 `if:"ExtraData,!=,0"`
	SpeedZ    int16 `if:"ExtraData,!=,0"`
}

func (SpawnObject) ID() byte { return 0x11 }

type SpawnMob struct {
	EntityID  int32
	Type      int8
	X         int32
	Y         int32
	Z         int32
	Pitch     int8
	HeadPitch int8
	Yaw       int8
	VelocityX int16
	VelocityY int16
	VelocityZ int16
	Metadata  map[byte]interface{} `metadata:"true"`
}

func (SpawnMob) ID() byte { return 0x12 }

type SpawnPainting struct {
	EntityID  int32
	Title     string
	X         int32
	Y         int32
	Z         int32
	Direction int32
}

func (SpawnPainting) ID() byte { return 0x13 }

type SpawnExperienceOrb struct {
	EntityID int32
	X        int32
	Y        int32
	Z        int32
	Count    int16
}

func (SpawnExperienceOrb) ID() byte { return 0x14 }

type EntityVelocity struct {
	EntityID  int32
	VelocityX int16
	VelocityY int16
	VelocityZ int16
}

func (EntityVelocity) ID() byte { return 0x15 }

type EntityDestroy struct {
	EntityIDs []int32 `ltype:"int8"`
}

func (EntityDestroy) ID() byte { return 0x16 }

type Entity struct {
	EntityID int32
}

func (Entity) ID() byte { return 0x17 }

type EntityMove struct {
	EntityID int32
	DX       int8
	DY       int8
	DZ       int8
}

func (EntityMove) ID() byte { return 0x18 }

type EntityLook struct {
	EntityID int32
	Yaw      int8
	Pitch    int8
}

func (EntityLook) ID() byte { return 0x19 }

type EntityLookMove struct {
	EntityID int32
	DX       int8
	DY       int8
	DZ       int8
	Yaw      int8
	Pitch    int8
}

func (EntityLookMove) ID() byte { return 0x1A }

type EntityTeleport struct {
	EntityID int32
	X        int32
	Y        int32
	Z        int32
	Yaw      int8
	Pitch    int8
}

func (EntityTeleport) ID() byte { return 0x1B }

type EntityHeadLook struct {
	EntityID int32
	HeadYaw  int8
}

func (EntityHeadLook) ID() byte { return 0x1C }

type EntityStatus struct {
	EntityID int32
	Status   int8
}

func (EntityStatus) ID() byte { return 0x1D }

type EntityAttach struct {
	EntityID  int32
	VehicleID int32
	Leash     bool
}

func (EntityAttach) ID() byte { return 0x1E }

type EntityMetadata struct {
	EntityID int32
	Metadata map[byte]interface{} `metadata:"true"`
}

func (EntityMetadata) ID() byte { return 0x1F }

type EntityEffect struct {
	EntityID  int32
	EffectID  int8
	Amplifier int8
	Duration  int16
}

func (EntityEffect) ID() byte { return 0x20 }

type EntityEffectRemove struct {
	EntityID int32
	EffectID int8
}

func (EntityEffectRemove) ID() byte { return 0x21 }

type SetExperience struct {
	ExperienceBar   float32
	Level           int16
	TotalExperience int16
}

func (SetExperience) ID() byte { return 0x22 }

type EntityProperties struct {
	EntityID   int32
	Properties []Property `ltype:"int32"`
}

//Part of Entity Properties
type Property struct {
	Key       string
	Value     float64
	Modifiers []Modifier `ltype:"int16"`
}

//Part of Entity Properties
type Modifier struct {
	UUIDHigh  int64
	UUIDLow   int64
	Amount    float64
	Operation int8
}

func (EntityProperties) ID() byte { return 0x23 }

type ChunkData struct {
	X              int32
	Z              int32
	GroundUp       bool
	PrimaryBitMap  uint16
	AddBitMap      uint16
	CompressedData []byte `ltype:"int32"`
}

func (ChunkData) ID() byte { return 0x24 }

type MultiBlockChange struct {
	X           int32
	Z           int32
	RecordCount int16
	Data        []byte `ltype:"int32"`
}

func (MultiBlockChange) ID() byte { return 0x25 }

type BlockChange struct {
	X    int32
	Y    byte
	Z    int32
	Type int16
	Data byte
}

func (BlockChange) ID() byte { return 0x26 }

type BlockAction struct {
	X            int32
	Y            int16
	Z            int32
	Byte1, Byte2 byte
	BlockID      int16
}

func (BlockAction) ID() byte { return 0x27 }

type MapChunkBulk struct {
	ChunkCount int16
	DataLength int32
	SkyLight   byte
	Data       []byte      `ltype:"nil"`
	Meta       []ChunkMeta `ltype:"nil"`
}

type ChunkMeta struct {
	X, Z       int32
	PrimaryBit uint16
	AddBitmap  uint16
}

func (MapChunkBulk) ID() byte { return 0x29 }

type Explosion struct {
	X       float64
	Y       float64
	Z       float64
	Radius  float32
	Records []Record `ltype:"int32"`
	MotionX float32
	MotionY float32
	MotionZ float32
}

//Part of Explosion
type Record struct {
	X byte
	Y byte
	Z byte
}

func (Explosion) ID() byte { return 0x2A }

type Effect struct {
	EffectID        int32
	X               int32
	Y               byte
	Z               int32
	Data            int32
	DisableRelative bool
}

func (Effect) ID() byte { return 0x2B }

type SoundEffect struct {
	Name          string
	X             int32
	Y             int32
	Z             int32
	Volume        float32
	Pitch         int8
	SoundCategory byte
}

func (SoundEffect) ID() byte { return 0x2C }

type Particle struct {
	Name          string
	X             float32
	Y             float32
	Z             float32
	OffsetX       float32
	OffsetY       float32
	OffsetZ       float32
	ParticleSpeed float32
	Count         int32
}

func (Particle) ID() byte { return 0x2D }

type GameState struct {
	Reason int8
	Value  float32
}

func (GameState) ID() byte { return 0x2E }

type SpawnGlobalEntity struct {
	EntityID int32
	Type     int8
	X        int32
	Y        int32
	Z        int32
}

func (SpawnGlobalEntity) ID() byte { return 0x2F }

type WindowOpen struct {
	WindowID int8
	Type     int8
	Title    string
	Slots    int8
	UseTitle bool
	EntityID int32 `if:"Type,==,11"`
}

func (WindowOpen) ID() byte { return 0x30 }

type WindowClose struct {
	WindowID int8
}

func (WindowClose) ID() byte { return 0x31 }

type WindowSetSlot struct {
	WindowID int8
	Slot     int16
	Item     Slot
}

func (WindowSetSlot) ID() byte { return 0x32 }

type WindowItems struct {
	WindowID int8
	Slots    []Slot `ltype:"int16"`
}

func (WindowItems) ID() byte { return 0x33 }

type WindowUpdateProperty struct {
	WindowID int8
	Property int16
	Value    int16
}

func (WindowUpdateProperty) ID() byte { return 0x34 }

type WindowTransactionConfirm struct {
	WindowID     int8
	ActionNumber int16
	Accepted     bool
}

func (WindowTransactionConfirm) ID() byte { return 0x35 }

type UpdateSign struct {
	X     int32
	Y     int16
	Z     int32
	Line1 string
	Line2 string
	Line3 string
	Line4 string
}

func (UpdateSign) ID() byte { return 0x36 }

type Maps struct {
	ItemType int16
	ItemData int16
	Data     []byte `ltype:"int16"`
}

func (Maps) ID() byte { return 0x37 }

type UpdateBlockEntity struct {
	X      int32
	Y      int16
	Z      int32
	Action int8
	Data   []byte `ltype:"int16"`
}

func (UpdateBlockEntity) ID() byte { return 0x38 }

type BlockEditorOpen struct {
	TileEntityID int8
	X            int32
	Y            int32
	Z            int32
}

func (BlockEditorOpen) ID() byte { return 0x39 }

type Statistics struct {
	Statistics []Statistic `ltype:"int32"`
}

type Statistic struct {
	Name   string
	Amount int32
}

func (Statistics) ID() byte { return 0x3A }

type PlayerListItem struct {
	PlayerName string
	Online     bool
	Ping       int16
}

func (PlayerListItem) ID() byte { return 0x3B }

type PlayerAbilities struct {
	Flags        byte
	FlyingSpeed  float32
	WalkingSpeed float32
}

func (PlayerAbilities) ID() byte { return 0x3C }

type TabComplete struct {
	Text string
}

func (TabComplete) ID() byte { return 0x3D }

type ScoreboardObjective struct {
	Name  string
	Value string
	Mode  int8
}

func (ScoreboardObjective) ID() byte { return 0x3E }

type UpdateScore struct {
	ObjectiveName string
	Mode          int8
	Name          string
	Value         int32
}

func (UpdateScore) ID() byte { return 0x3F }

type DisplayScoreboard struct {
	Position      int8
	ObjectiveName string
}

func (DisplayScoreboard) ID() byte { return 0x40 }

type Teams struct {
	Name        string
	Mode        byte
	DisplayName string   `if:"Mode,==,0|2"`
	Prefix      string   `if:"Mode,==,0|2"`
	Suffix      string   `if:"Mode,==,0|2"`
	Flags       byte     `if:"Mode,==,0|2"`
	Players     []string `if:"Mode,==,0|3|4" ltype:"int16"`
}

func (Teams) ID() byte { return 0x41 }

type PluginMessage struct {
	Channel string
	Data    []byte `ltype:"int16"`
}

func (PluginMessage) ID() byte { return 0x42 }

type Disconnect struct {
	Reason string
}

func (Disconnect) ID() byte { return 0x43 }
