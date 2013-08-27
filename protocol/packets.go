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
	notUsed    int8
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
	Metadata    map[byte]interface{}
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
	Metadata  map[byte]interface{}
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
