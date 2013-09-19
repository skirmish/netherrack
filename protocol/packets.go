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

//Keep Alive (0x00)
//
//This normally contains an random ID. The client should respond
//with the same ID.
//Server <--> Client
type KeepAlive struct {
	KeepAliveID int32
}

func (KeepAlive) ID() byte { return 0x00 }

//Login Request (0x01)
//
//Sent by the server once login is complete. Contains infomation about the
//player's entity ID and the world the player spawned into.
//Server --> Client
type LoginRequest struct {
	EntityID   int32
	LevelType  string
	Gamemode   int8
	Dimension  int8
	Difficulty int8
	NotUsed    int8
	MaxPlayers int8
}

func (LoginRequest) ID() byte { return 0x01 }

//Handshake (0x02)
//
//Sent by the client apon connecting
//Server <-- Client
type Handshake struct {
	ProtocolVersion byte
	Username        string
	ServerHost      string
	ServerPort      int32
}

func (Handshake) ID() byte { return 0x02 }

//Chat Message (0x02)
//
//When sent by the server the message should be json encoded
//Server <--> Client
type ChatMessage struct {
	Message string
}

func (ChatMessage) ID() byte { return 0x03 }

//Time Update (0x04)
//
//Age of the world is time in ticks not changed by server commands (e.g: /time set).
//Time of day controls the position of the sun and the sky light levels
//Server --> Client
type TimeUpdate struct {
	AgeOfTheWorld int64
	TimeOfDay     int64
}

func (TimeUpdate) ID() byte { return 0x04 }

//Entity Equipment
//
//Server --> Client
type EntityEquipment struct {
	EntityID int32
	Slot     int16
	Item     Slot
}

func (EntityEquipment) ID() byte { return 0x05 }

//Spawn Position (0x06)
//
//Server --> Client
type SpawnPosition struct {
	X int32
	Y int32
	Z int32
}

func (SpawnPosition) ID() byte { return 0x06 }

//Use Entity (0x07)
//
//Server <-- Client
type UseEntity struct {
	User        int32
	Target      int32
	MouseButton bool
}

func (UseEntity) ID() byte { return 0x07 }

//Update Health (0x08)
//
//Server --> Client
type UpdateHealth struct {
	Health         float32
	Food           int16
	FoodSaturation float32
}

func (UpdateHealth) ID() byte { return 0x08 }

//Respawn (0x09)
//
//Server --> Client
type Respawn struct {
	Dimension   int32
	Difficulty  int8
	Gamemode    int8
	WorldHeight int16
	LevelType   string
}

func (Respawn) ID() byte { return 0x09 }

//Player (0x0A)
//
//Server <-- Client
type Player struct {
	OnGround bool
}

func (Player) ID() byte { return 0x0A }

//Player Position (0x0B)
//
//Server <-- Client
type PlayerPosition struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	OnGround bool
}

func (PlayerPosition) ID() byte { return 0x0B }

//Player Look (0x0C)
//
//Server <-- Client
type PlayerLook struct {
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (PlayerLook) ID() byte { return 0x0C }

//Player Position and Look (0x0D)
//
//Server <--> Client
type PlayerPositionLook struct {
	X        float64
	Y        float64
	Stance   float64
	Z        float64
	Yaw      float32
	Pitch    float32
	OnGround bool
}

func (PlayerPositionLook) ID() byte { return 0x0D }

//Player Digging (0x0E)
//
//Server <-- Client
type PlayerDigging struct {
	Status int8
	X      int32
	Y      byte
	Z      int32
	Face   int8
}

func (PlayerDigging) ID() byte { return 0x0E }

//Player Block Placement (0x0F)
//
//Server <-- Client
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

func (PlayerBlockPlacement) ID() byte { return 0x0F }

//Held Item Change (0x10)
//
//Server <--> Client
type HeldItemChange struct {
	SlotID int16
}

func (HeldItemChange) ID() byte { return 0x10 }

//Use Bed (0x11)
//
//Server --> Client
type UseBed struct {
	EntityID int32
	unknown  byte
	X        int32
	Y        byte
	Z        int32
}

func (UseBed) ID() byte { return 0x11 }

//Animation (0x12)
//
//Server <--> Client
type Animation struct {
	EntityID  int32
	Animation int8
}

func (Animation) ID() byte { return 0x12 }

//Entity Action (0x13)
//
//Server <-- Client
type EntityAction struct {
	EntityID  int32
	ActionID  int8
	JumpBoost int32
}

func (EntityAction) ID() byte { return 0x13 }

//Spawn Named Entity (0x14)
//
//Server --> Client
type SpawnNamedEntity struct {
	EntityID    int32
	PlayerName  string
	X           int32
	Y           int32
	Z           int32
	Yaw         int8
	Pitch       int8
	CurrentItem int16
	Metadata    map[byte]interface{} `metadata:"true"`
}

func (SpawnNamedEntity) ID() byte { return 0x14 }

//Collect Item (0x16)
//
//Server --> Client
type CollectItem struct {
	CollectedEntityID int32
	CollectorEntityID int32
}

func (CollectItem) ID() byte { return 0x16 }

//Spawn Object (0x17)
//
//Server --> Client
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

func (SpawnObject) ID() byte { return 0x17 }

//Spawn Mob (0x18)
//
//Server --> Client
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

func (SpawnMob) ID() byte { return 0x18 }

//Spawn Painting (0x19)
//
//Server --> Client
type SpawnPainting struct {
	EntityID  int32
	Title     string
	X         int32
	Y         int32
	Z         int32
	Direction int32
}

func (SpawnPainting) ID() byte { return 0x19 }

//Spawn Experience Orb (0x1A)
//
//Server --> Client
type SpawnExperienceOrb struct {
	EntityID int32
	X        int32
	Y        int32
	Z        int32
	Count    int16
}

func (SpawnExperienceOrb) ID() byte { return 0x1A }

//Steer Vehicle (0x1B)
//
//Server <-- Client
type SteerVehicle struct {
	Sideways float32
	Forward  float32
	Jump     bool
	Unmount  bool
}

func (SteerVehicle) ID() byte { return 0x1B }

//Entity Velocity (0x1C)
//
//Server --> Client
type EntityVelocity struct {
	EntityID  int32
	VelocityX int16
	VelocityY int16
	VelocityZ int16
}

func (EntityVelocity) ID() byte { return 0x1C }

//Entity Destroy (0x1D)
//
//Server --> Client
type EntityDestroy struct {
	EntityIDs []int32 `ltype:"int8"`
}

func (EntityDestroy) ID() byte { return 0x1D }

//Entity (0x1E)
//
//Server --> Client
type Entity struct {
	EntityID int32
}

func (Entity) ID() byte { return 0x1E }

//Entity Move (0x1F)
//
//Server --> Client
type EntityMove struct {
	EntityID int32
	DX       int8
	DY       int8
	DZ       int8
}

func (EntityMove) ID() byte { return 0x1F }

//Entity Look (0x20)
//
//Server --> Client
type EntityLook struct {
	EntityID int32
	Yaw      int8
	Pitch    int8
}

func (EntityLook) ID() byte { return 0x20 }

//Entity Look and Move (0x21)
//
//Server --> Client
type EntityLookMove struct {
	EntityID int32
	DX       int8
	DY       int8
	DZ       int8
	Yaw      int8
	Pitch    int8
}

func (EntityLookMove) ID() byte { return 0x21 }

//Entity Teleport (0x22)
//
//Server --> Client
type EntityTeleport struct {
	EntityID int32
	X        int32
	Y        int32
	Z        int32
	Yaw      int8
	Pitch    int8
}

func (EntityTeleport) ID() byte { return 0x22 }

//Entity Head Look (0x23)
//
//Server --> Client
type EntityHeadLook struct {
	EntityID int32
	HeadYaw  int8
}

func (EntityHeadLook) ID() byte { return 0x23 }

//Entity Status (0x26)
//
//Server --> Client
type EntityStatus struct {
	EntityID int32
	Status   int8
}

func (EntityStatus) ID() byte { return 0x26 }

//Entity Attach (0x27)
//
//Server --> Client
type EntityAttach struct {
	EntityID  int32
	VehicleID int32
	Leash     bool
}

func (EntityAttach) ID() byte { return 0x27 }

//Entity Metadata (0x28)
//
//Server --> Client
type EntityMetadata struct {
	EntityID int32
	Metadata map[byte]interface{} `metadata:"true"`
}

func (EntityMetadata) ID() byte { return 0x28 }

//Entity Effect (0x29)
//
//Server --> Client
type EntityEffect struct {
	EntityID  int32
	EffectID  int8
	Amplifier int8
	Duration  int16
}

func (EntityEffect) ID() byte { return 0x29 }

//Entity Effect Remove (0x2A)
//
//Server --> Client
type EntityEffectRemove struct {
	EntityID int32
	EffectID int8
}

func (EntityEffectRemove) ID() byte { return 0x2A }

//Set Experience (0x2B)
//
//Server --> Client
type SetExperience struct {
	ExperienceBar   float32
	Level           int16
	TotalExperience int16
}

func (SetExperience) ID() byte { return 0x2B }

//Entity Properties (0x2C)
//
//Server --> Client
type EntityProperties struct {
	EntityID   int32
	Properties []Property `ltype:"int32"`
}

//Part of Entity Properties (0x2C)
type Property struct {
	Key       string
	Value     float64
	Modifiers []Modifier `ltype:"int16"`
}

//Part of Entity Properties (0x2C)
type Modifier struct {
	UUIDHigh  int64
	UUIDLow   int64
	Amount    float64
	Operation int8
}

func (EntityProperties) ID() byte { return 0x2C }

//Chunk Data (0x33)
//
//Server --> Client
type ChunkData struct {
	X              int32
	Z              int32
	GroundUp       bool
	PrimaryBitMap  uint16
	AddBitMap      uint16
	CompressedData []byte `ltype:"int32"`
}

func (ChunkData) ID() byte { return 0x33 }

//Multi Block Change (0x34)
//
//Server --> Client
type MultiBlockChange struct {
	X           int32
	Z           int32
	RecordCount int16
	Data        []int32 `ltype:"int32"`
}

func (MultiBlockChange) ID() byte { return 0x34 }

//Block Change (0x35)
//
//Server --> Client
type BlockChange struct {
	X    int32
	Y    byte
	Z    int32
	Type int16
	Data byte
}

func (BlockChange) ID() byte { return 0x35 }

//Block Action (0x36)
//
//Server --> Client
type BlockAction struct {
	X            int32
	Y            int16
	Z            int32
	Byte1, Byte2 byte
	BlockID      int16
}

func (BlockAction) ID() byte { return 0x36 }

//Block Break Animation
//
//Server --> Client
type BlockBreakAnimation struct {
	EntityID     int32
	X            int32
	Y            int32
	Z            int32
	DestroyStage int8
}

func (BlockBreakAnimation) ID() byte { return 0x37 }

//Map Chunk Bulk
//
//Server --> Client
//Currently not supported

//Explosion (0x3C)
//
//Server --> Client
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

//Part of Explosion (0x3C)
type Record struct {
	X byte
	Y byte
	Z byte
}

func (Explosion) ID() byte { return 0x3C }

//Effect (0x3D)
//
//Server --> Client
type Effect struct {
	EffectID        int32
	X               int32
	Y               byte
	Z               int32
	Data            int32
	DisableRelative bool
}

func (Effect) ID() byte { return 0x3D }

//Sound Effect (0x3E)
//
//Server --> Client
type SoundEffect struct {
	Name          string
	X             int32
	Y             int32
	Z             int32
	Volume        float32
	Pitch         int8
	SoundCategory byte
}

func (SoundEffect) ID() byte { return 0x3E }

//Particle (0x3F)
//
//Server --> Client
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

func (Particle) ID() byte { return 0x3F }

//Game State (0x46)
//
//Server --> Client
type GameState struct {
	Reason   int8
	Gamemode int8
}

func (GameState) ID() byte { return 0x46 }

//Spawn Global Entity (0x47)
//
//Server --> Client
type SpawnGlobalEntity struct {
	EntityID int32
	Type     int8
	X        int32
	Y        int32
	Z        int32
}

func (SpawnGlobalEntity) ID() byte { return 0x47 }

//Window Open (0x64)
//
//Server --> Client
type WindowOpen struct {
	WindowID int8
	Type     int8
	Title    string
	Slots    int8
	UseTitle bool
	EntityID int32 `if:"Type,==,11"`
}

func (WindowOpen) ID() byte { return 0x64 }

//Window Close (0x65)
//
//Serve<r --> Client
type WindowClose struct {
	WindowID int8
}

func (WindowClose) ID() byte { return 0x65 }

//Window Click (0x66)
//
//Server <-- Client
type WindowClick struct {
	WindowID     int8
	Slot         int16
	Button       int8
	ActionNumber int16
	Mode         int8
	Item         Slot
}

func (WindowClick) ID() byte { return 0x66 }

//Window Set Slot
//
//Server --> Client
type WindowSetSlot struct {
	WindowID int8
	Slot     int16
	Item     Slot
}

func (WindowSetSlot) ID() byte { return 0x67 }

//Window Set Slots
//
//Server --> Client
type WindowSetSlots struct {
	WindowID int8
	Slots    []Slot `ltype:"int16"`
}

func (WindowSetSlots) ID() byte { return 0x68 }

//Window Update Property (0x69)
//
//Server --> Client
type WindowUpdateProperty struct {
	WindowID int8
	Property int16
	Value    int16
}

func (WindowUpdateProperty) ID() byte { return 0x69 }

//Confirm Transaction (0x6A)
//
//Server <--> Client
type WindowTransactionConfirm struct {
	WindowID     int8
	ActionNumber int16
	Accepted     bool
}

func (WindowTransactionConfirm) ID() byte { return 0x6A }

//Creative Inventory Action (0x6B)
//
//Server <--> Client
type CreativeInventoryAction struct {
	Slot int16
	Item Slot
}

func (CreativeInventoryAction) ID() byte { return 0x6B }

//Enchant Item (0x6C)
//
//Server <-- Client
type EnchantItem struct {
	WindowID    int8
	Enchantment int8
}

func (EnchantItem) ID() byte { return 0x6C }

//Update Sign (0x82)
//
//Server <--> Client
type UpdateSign struct {
	X     int32
	Y     int16
	Z     int32
	Line1 string
	Line2 string
	Line3 string
	Line4 string
}

func (UpdateSign) ID() byte { return 0x82 }

//Item Data (0x83)
//
//Server --> Client
type ItemData struct {
	ItemType int16
	ItemData int16
	Data     []byte `ltype:"int16"`
}

func (ItemData) ID() byte { return 0x83 }

//Update Tile Entity (0x84)
//
//Server --> Client
type UpdateTileEntity struct {
	X      int32
	Y      int16
	Z      int32
	Action int8
	Data   []byte `ltype:"int16"`
}

func (UpdateTileEntity) ID() byte { return 0x84 }

//Tile Editor Open (0x85)
//
//Server --> Client
type TileEditorOpen struct {
	TileEntityID int8
	X            int32
	Y            int32
	Z            int32
}

func (TileEditorOpen) ID() byte { return 0x85 }

//Increment Statistic (0xC8)
//
//Server --> Client
type IncrementStatistic struct {
	Statistics []Statistic `ltype:"int32"`
}

type Statistic struct {
	Name   string
	Amount int32
}

func (IncrementStatistic) ID() byte { return 0xC8 }

//Player List Item (0xC9)
//
//Server --> Client
type PlayerListItem struct {
	PlayerName string
	Online     bool
	Ping       int16
}

func (PlayerListItem) ID() byte { return 0xC9 }

//Player Abilities (0xCA)
//
//Server <--> Client
type PlayerAbilities struct {
	Flags        byte
	FlyingSpeed  float32
	WalkingSpeed float32
}

func (PlayerAbilities) ID() byte { return 0xCA }

//Tab-complete (0xCB)
//
//Server <--> Client
type TabComplete struct {
	Text string
}

func (TabComplete) ID() byte { return 0xCB }

//Client Settings (0xCC)
//
//Server <-- Client
type ClientSettings struct {
	Locale       string
	ViewDistance int8
	ChatFlags    byte
	Difficulty   int8
	ShowCape     bool
}

func (ClientSettings) ID() byte { return 0xCC }

//Client Statuses (0xCD)
//
//Server <-- Client
type ClientStatuses struct {
	Payload byte
}

func (ClientStatuses) ID() byte { return 0xCD }

//Scoreboard Objective (0xCE)
//
//Server --> Client
type ScoreboardObjective struct {
	Name  string
	Value string
	Mode  int8
}

func (ScoreboardObjective) ID() byte { return 0xCE }

//Update Score
//
//Server --> Client
type UpdateScore struct {
	ObjectiveName string
	Mode          int8
	Name          string
	Value         int32
}

func (UpdateScore) ID() byte { return 0xCF }

//Display Scoreboard (0xD0)
//
//Server --> Client
type DisplayScoreboard struct {
	Position      int8
	ObjectiveName string
}

func (DisplayScoreboard) ID() byte { return 0xD0 }

//Teams (0xD1)
//
//Server --> Client
type Teams struct {
	Name        string
	Mode        byte
	DisplayName string   `if:"Mode,==,0|2"`
	Prefix      string   `if:"Mode,==,0|2"`
	Suffix      string   `if:"Mode,==,0|2"`
	Flags       byte     `if:"Mode,==,0|2"`
	Players     []string `if:"Mode,==,0|3|4" ltype:"int16"`
}

func (Teams) ID() byte { return 0xD1 }

//Plugin Message (0xFA)
//
//Server <--> Client
type PluginMessage struct {
	Channel string
	Data    []byte `ltype:"int16"`
}

func (PluginMessage) ID() byte { return 0xFA }

//Encryption Key Response (0xFC)
//
//Server <--> Client
type EncryptionKeyResponse struct {
	SharedSecret []byte `ltype:"int16"`
	VerifyToken  []byte `ltype:"int16"`
}

func (EncryptionKeyResponse) ID() byte { return 0xFC }

//Encryption Key Request (0xFD)
//
//Server --> Client
type EncryptionKeyRequest struct {
	ServerID    string
	PublicKey   []byte `ltype:"int16"`
	VerifyToken []byte `ltype:"int16"`
}

func (EncryptionKeyRequest) ID() byte { return 0xFD }

//Server List Ping (0xFE)
//
//Server <-- Client
type ServerListPing struct {
	Magic byte
}

func (ServerListPing) ID() byte { return 0xFE }

//Disconnect (0xFF)
//
//Server <--> Client
type Disconnect struct {
	Reason string
}

func (Disconnect) ID() byte { return 0xFF }
