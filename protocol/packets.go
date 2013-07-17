package protocol

import (
	"bytes"
	"github.com/NetherrackDev/netherrack/entity/metadata"
	"github.com/NetherrackDev/netherrack/nbt"
	"github.com/NetherrackDev/soulsand"
	"runtime"
)

//Keep Alive (0x00)
const KeepAlive = 0x00

//Keep Alive (0x00)
//Server <--> Client
func (c *Conn) WriteKeepAlive(id int32) {
	out := NewByteWriter(1 + 4)
	out.WriteUByte(KeepAlive)
	out.WriteInt(id)
	c.Write(out.Bytes())
}

//Keep Alive (0x00)
//Server <--> Client
func (c *Conn) ReadKeepAlive() int32 {
	return c.readInt()
}

//Login Request (0x01)
const LoginRequest = 0x01

//Login Request (0x01)
//Server --> Client
func (c *Conn) WriteLoginRequest(eId int32, levelType string, gamemode, dimension, difficuly, maxPlayers int8) {
	levelTypeRunes := []rune(levelType)
	out := NewByteWriter(1 + 4 + 2 + len(levelTypeRunes)*2 + 1 + 1 + 1 + 1 + 1)
	out.WriteUByte(LoginRequest)
	out.WriteInt(eId)
	out.WriteString(levelTypeRunes)
	out.WriteByte(gamemode)
	out.WriteByte(dimension)
	out.WriteByte(difficuly)
	out.WriteByte(0)
	out.WriteByte(maxPlayers)
	c.Write(out.Bytes())
}

//Login Request (0x01)
//Server --> Client
func (c *Conn) ReadLoginRequest() (eId int32, levelType string, gamemode, dimension, difficuly, maxPlayers int8) {
	eId = c.readInt()
	levelType = c.readString()
	gamemode = c.readByte()
	dimension = c.readByte()
	difficuly = c.readByte()
	c.readByte()
	maxPlayers = c.readByte()
	return
}

//Handshake (0x02)
const Handshake = 0x02

//Handshake (0x02)
//Server <-- Client
func (c *Conn) WriteHandshake(protoVersion byte, username, host string, port int32) {
	usernameRunes := []rune(username)
	hostRunes := []rune(host)
	out := NewByteWriter(1 + 1 + 2 + len(usernameRunes)*2 + 2 + len(hostRunes)*2 + 4)
	out.WriteUByte(Handshake)
	out.WriteUByte(protoVersion)
	out.WriteString(usernameRunes)
	out.WriteString(hostRunes)
	out.WriteInt(port)
}

//Handshake (0x02)
//Server <-- Client
func (c *Conn) ReadHandshake() (protoVersion byte, username, host string, port int32) {
	protoVersion = c.readUByte()
	username = c.readString()
	host = c.readString()
	port = c.readInt()
	return
}

//Chat Message (0x03)
const ChatMessage = 0x03

//Chat Message (0x03)
//Server <--> Client
func (c *Conn) WriteChatMessage(message string) {
	messageRunes := []rune(string(message))
	out := NewByteWriter(1 + 2 + len(messageRunes)*2)
	out.WriteUByte(ChatMessage)
	out.WriteString(messageRunes)
	c.Write(out.Bytes())
}

//Chat Message (0x03)
//Server <--> Client
func (c *Conn) ReadChatMessage() string {
	return c.readString()
}

//Time Update (0x04)
const TimeUpdate = 0x04

//Time Update (0x04)
//Server --> Client
func (c *Conn) WriteTimeUpdate(age, time int64) {
	out := NewByteWriter(1 + 8 + 8)
	out.WriteUByte(TimeUpdate)
	out.WriteLong(age)
	out.WriteLong(time)
	c.Write(out.Bytes())
}

//Time Update (0x04)
//Server --> Client
func (c *Conn) ReadTimeUpdate() (age, time int64) {
	age = c.readLong()
	time = c.readLong()
	return
}

//Entity Equipment (0x05)
const EntityEquipment = 0x05

//Entity Equipment (0x05)
//Server --> Client
func (c *Conn) WriteEntityEquipment(eId int32, slot int16, itemstack soulsand.ItemStack) {
	out := NewByteWriter(1 + 4 + 2)
	out.WriteUByte(EntityEquipment)
	out.WriteInt(eId)
	out.WriteShort(slot)
	c.Write(out.Bytes())
	c.writeItemstack(itemstack)
}

//Entity Equipment (0x05)
//Server --> Client
func (c *Conn) ReadEntityEquipment() (eId int32, slot int16, itemstack soulsand.ItemStack) {
	eId = c.readInt()
	slot = c.readShort()
	itemstack = c.readSlot()
	return
}

//Spawn Position (0x06)
const SpawnPosition = 0x06

//Spawn Position (0x06)
//Server --> Client
func (c *Conn) WriteSpawnPosition(x, y, z int32) {
	out := NewByteWriter(1 + 4 + 4 + 4)
	out.WriteUByte(SpawnPosition)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	c.Write(out.Bytes())
}

//Spawn Position (0x06)
//Server --> Client
func (c *Conn) ReadSpawnPosition() (x, y, z int32) {
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	return
}

//Use Entity (0x07)
const UseEntity = 0x07

//Use Entity (0x07)
//Server <-- Client
func (c *Conn) WriteUseEntity(user, target int32, mouseButton bool) {
	out := NewByteWriter(1 + 4 + 4 + 1)
	out.WriteUByte(UseEntity)
	out.WriteInt(user)
	out.WriteInt(target)
	out.WriteBool(mouseButton)
	c.Write(out.Bytes())
}

//Use Entity (0x07)
//Server <-- Client
func (c *Conn) ReadUseEntity() (user, target int32, mouseButton bool) {
	user = c.readInt()
	target = c.readInt()
	mouseButton = c.readBool()
	return
}

//Update Health (0x08)
const UpdateHealth = 0x08

//Update Health (0x08)
//Server --> Client
func (c *Conn) WriteUpdateHealth(health float32, food int16, foodSatiration float32) {
	out := NewByteWriter(1 + 4 + 2 + 4)
	out.WriteUByte(UpdateHealth)
	out.WriteFloat(health)
	out.WriteShort(food)
	out.WriteFloat(foodSatiration)
	c.Write(out.Bytes())
}

//Update Health (0x08)
//Server --> Client
func (c *Conn) ReadUpdateHealth() (health float32, food int16, foodSatiration float32) {
	health = c.readFloat()
	food = c.readShort()
	foodSatiration = c.readFloat()
	return
}

//Respawn (0x09)
const Respawn = 0x09

//Respawn (0x09)
//Server --> Client
func (c *Conn) WriteRespawn(dimension int32, difficulty, gamemode int8, worldHeight int16, levelType string) {
	levelTypeRune := []rune(levelType)
	out := NewByteWriter(1 + 4 + 1 + 1 + 2 + 2 + len(levelTypeRune)*2)
	out.WriteUByte(0x09)
	out.WriteInt(dimension)
	out.WriteByte(difficulty)
	out.WriteByte(gamemode)
	out.WriteShort(worldHeight)
	out.WriteString(levelTypeRune)
	c.Write(out.Bytes())
}

//Respawn (0x09)
//Server --> Client
func (c *Conn) ReadRespawn() (dimension int32, difficulty, gamemode int8, worldHeight int16, levelType string) {
	dimension = c.readInt()
	difficulty = c.readByte()
	gamemode = c.readByte()
	worldHeight = c.readShort()
	levelType = c.readString()
	return
}

//Player (0x0A)
const Player = 0x0A

//Player (0x0A)
//Server <-- Client
func (c *Conn) ReadPlayer() bool {
	return c.readBool()
}

//Player (0x0A)
//Server <-- Client
func (c *Conn) WritePlayer(onGround bool) {
	out := NewByteWriter(1 + 1)
	out.WriteUByte(Player)
	out.WriteBool(onGround)
	c.Write(out.Bytes())
}

//Player Position (0x0B)
const PlayerPosition = 0x0B

//Player Position (0x0B)
//Server <-- Client
func (c *Conn) WritePlayerPosition(x, y, stance, z float64, onGround bool) {
	out := NewByteWriter(1 + 8 + 8 + 8 + 8 + 1)
	out.WriteUByte(PlayerPosition)
	out.WriteDouble(x)
	out.WriteDouble(y)
	out.WriteDouble(stance)
	out.WriteDouble(z)
	out.WriteBool(onGround)
	c.Write(out.Bytes())
}

//Player Position (0x0B)
//Server <-- Client
func (c *Conn) ReadPlayerPosition() (x, y, stance, z float64, onGround bool) {
	x = c.readDouble()
	y = c.readDouble()
	stance = c.readDouble()
	z = c.readDouble()
	onGround = c.readBool()
	return
}

//Player Look (0x0C)
const PlayerLook = 0x0C

//Player Look (0x0C)
//Server <-- Client
func (c *Conn) WritePlayerLook(yaw, pitch float32, onGround bool) {
	out := NewByteWriter(1 + 4 + 4 + 1)
	out.WriteUByte(PlayerLook)
	out.WriteFloat(yaw)
	out.WriteFloat(pitch)
	out.WriteBool(onGround)
	c.Write(out.Bytes())
}

//Player Look (0x0C)
//Server <-- Client
func (c *Conn) ReadPlayerLook() (yaw, pitch float32, onGround bool) {
	yaw = c.readFloat()
	pitch = c.readFloat()
	onGround = c.readBool()
	return
}

//Player Position and Look (0x0D)
const PlayerPositionAndLook = 0x0D

//Player Position and Look (0x0D)
//Server <--> Client
func (c *Conn) WritePlayerPositionLook(x, y, z, stance float64, yaw, pitch float32, onGround bool) {
	out := NewByteWriter(1 + 8 + 8 + 8 + 8 + 4 + 4 + 1)
	out.WriteUByte(0x0D)
	out.WriteDouble(x)
	out.WriteDouble(y)
	out.WriteDouble(stance)
	out.WriteDouble(z)
	out.WriteFloat(yaw)
	out.WriteFloat(pitch)
	out.WriteBool(onGround)
	c.Write(out.Bytes())
}

//Player Position and Look (0x0D)
//Server <--> Client
func (c *Conn) ReadPlayerPositionLook() (x, y, stance, z float64, yaw, pitch float32, onGround bool) {
	x = c.readDouble()
	y = c.readDouble()
	stance = c.readDouble()
	z = c.readDouble()
	yaw = c.readFloat()
	pitch = c.readFloat()
	onGround = c.readBool()
	return
}

//Player Digging (0x0E)
const PlayerDigging = 0x0E

//Player Digging (0x0E)
//Server <-- Client
func (c *Conn) WritePlayerDigging(status int8, x int32, y byte, z int32, face int8) {
	out := NewByteWriter(1 + 1 + 4 + 1 + 4 + 1)
	out.WriteUByte(PlayerDigging)
	out.WriteInt(x)
	out.WriteUByte(y)
	out.WriteInt(z)
	out.WriteByte(face)
	c.Write(out.Bytes())
}

//Player Digging (0x0E)
//Server <-- Client
func (c *Conn) ReadPlayerDigging() (status int8, x int32, y byte, z int32, face int8) {
	status = c.readByte()
	x = c.readInt()
	y = c.readUByte()
	z = c.readInt()
	face = c.readByte()
	return
}

//Player Block Placement (0x0F)
const PlayerBlockPlacement = 0x0F

//Player Block Placement (0x0F)
//Server <-- Client
func (c *Conn) WritePlayerBlockPlacement(x int32, y byte, z int32, direction, cursorX, cursorY, cursorZ byte) {
	out := NewByteWriter(1 + 4 + 1 + 4 + 1 + 2 + 1 + 1 + 1)
	out.WriteUByte(PlayerBlockPlacement)
	out.WriteInt(x)
	out.WriteUByte(y)
	out.WriteInt(z)
	out.WriteUByte(direction)
	out.WriteShort(-1) //Itemstack is not needed
	out.WriteUByte(cursorX)
	out.WriteUByte(cursorY)
	out.WriteUByte(cursorZ)
	c.Write(out.Bytes())
}

//Player Block Placement (0x0F)
//Server <-- Client
//Doesn't return the heldItem as it could be changed by the client
func (c *Conn) ReadPlayerBlockPlacement() (x int32, y byte, z int32, direction byte, cursorX, cursorY, cursorZ byte) {
	x = c.readInt()
	y = c.readUByte()
	z = c.readInt()
	direction = c.readUByte()
	c.readSlot()
	cursorX = c.readUByte()
	cursorY = c.readUByte()
	cursorZ = c.readUByte()
	return
}

//Held Item Change (0x10)
const HeldItemChange = 0x10

//Held Item Change (0x10)
//Server <--> Client
func (c *Conn) WriteHeldItemChange(slotID int16) {
	out := NewByteWriter(1 + 2)
	out.WriteUByte(HeldItemChange)
	out.WriteShort(slotID)
	c.Write(out.Bytes())
}

//Held Item Change (0x10)
//Server <--> Client
func (c *Conn) ReadHeldItemChange() int16 {
	return c.readShort()
}

//Use Bed (0x11)
const UseBed = 0x11

//Use Bed (0x11)
//Server --> Client
func (c *Conn) WriteUseBed(eId, x int32, y byte, z int32) {
	out := NewByteWriter(1 + 4 + 1 + 4 + 1 + 4)
	out.WriteUByte(UseBed)
	out.WriteInt(eId)
	out.WriteUByte(0)
	out.WriteInt(x)
	out.WriteUByte(y)
	out.WriteInt(z)
	c.Write(out.Bytes())
}

//Use Bed (0x11)
//Server --> Client
func (c *Conn) ReadUseBed() (eId, x int32, y byte, z int32) {
	eId = c.readInt()
	c.readUByte()
	x = c.readInt()
	y = c.readUByte()
	z = c.readInt()
	return
}

//Animation (0x12)
const Animation = 0x12

//Animation (0x12)
//Server <--> Client
func (c *Conn) WriteAnimation(eId int32, animation int8) {
	out := NewByteWriter(1 + 4 + 1)
	out.WriteUByte(Animation)
	out.WriteInt(eId)
	out.WriteByte(animation)
	c.Write(out.Bytes())
}

//Animation (0x12)
//Server <--> Client
func (c *Conn) ReadAnimation() (eId int32, animation int8) {
	eId = c.readInt()
	animation = c.readByte()
	return
}

//Entity Action (0x13)
const EntityAction = 0x13

//Entity Action (0x13)
//Server <-- Client
func (c *Conn) WriteEntityAction(eId int32, action int8, jumpBoost int32) {
	out := NewByteWriter(1 + 4 + 1 + 4)
	out.WriteUByte(EntityAction)
	out.WriteInt(eId)
	out.WriteByte(action)
	out.WriteInt(jumpBoost)
	c.Write(out.Bytes())
}

//Entity Action (0x13)
//Server <-- Client
func (c *Conn) ReadEntityAction() (eId int32, actionID int8, jumpBoost int32) {
	eId = c.readInt()
	actionID = c.readByte()
	jumpBoost = c.readInt()
	return
}

//Spawn Named Entity (0x14)
const SpawnNamedEntity = 0x14

//Spawn Named Entity (0x14)
//Server --> Client
func (c *Conn) WriteSpawnNamedEntity(eID int32, playerName string, x, y, z int32, yaw, pitch int8, currentItem int16, meta soulsand.EntityMetadata) {
	playerNameRunes := []rune(playerName)
	out := NewByteWriter(1 + 4 + 2 + len(playerNameRunes)*2 + 4 + 4 + 4 + 1 + 1 + 2)
	out.WriteUByte(SpawnNamedEntity)
	out.WriteInt(eID)
	out.WriteString(playerNameRunes)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteByte(yaw)
	out.WriteByte(pitch)
	out.WriteShort(currentItem)
	c.Write(out.Bytes())
	(meta.(metadata.Type)).WriteTo(c)
}

//Spawn Named Entity (0x14)
//Server --> Client
func (c *Conn) ReadSpawnNamedEntity() (eID int32, playerName string, x, y, z int32, yaw, pitch int8, currentItem int16, meta soulsand.EntityMetadata) {
	eID = c.readInt()
	playerName = c.readString()
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	yaw = c.readByte()
	pitch = c.readByte()
	currentItem = c.readShort()
	meta = metadata.ReadForm(c)
	return
}

//Collect Item (0x16)
const CollectItem = 0x16

//Collect Item (0x16)
//Server --> Client
func (c *Conn) WriteCollectItem(collectedEId, collectorEId int32) {
	out := NewByteWriter(1 + 4 + 4)
	out.WriteUByte(CollectItem)
	out.WriteInt(collectedEId)
	out.WriteInt(collectorEId)
	c.Write(out.Bytes())
}

//Collect Item (0x16)
//Server --> Client
func (c *Conn) ReadCollectItem() (collectedEId, collectorEId int32) {
	collectedEId = c.readInt()
	collectorEId = c.readInt()
	return
}

//Spawn Object/Vehicle (0x17)
const SpawnObjectVehicle = 0x17

//Spawn Object/Vehicle (0x17)
//Server --> Client
func (c *Conn) WriteSpawnObjectVehicle(eId int32, t int8, x, y, z int32, pitch, yaw int8, data int32, speedX, speedY, speedZ int16) {
	var out *byteWriter
	if data != 0 {
		out = NewByteWriter(1 + 4 + 1 + 4 + 4 + 4 + 1 + 1 + 4 + 2 + 2 + 2)
	} else {
		out = NewByteWriter(1 + 4 + 1 + 4 + 4 + 4 + 1 + 1 + 4)
	}
	out.WriteUByte(SpawnObjectVehicle)
	out.WriteInt(eId)
	out.WriteByte(t)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteByte(pitch)
	out.WriteByte(yaw)
	out.WriteInt(data)
	if data != 0 {
		out.WriteShort(speedX)
		out.WriteShort(speedY)
		out.WriteShort(speedZ)
	}
	c.Write(out.Bytes())
}

//Spawn Object/Vehicle (0x17)
//Server --> Client
func (c *Conn) ReadSpawnObjectVehicle() (eId int32, t int8, x, y, z int32, pitch, yaw int8, data int32, speedX, speedY, speedZ int16) {
	eId = c.readInt()
	t = c.readByte()
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	pitch = c.readByte()
	yaw = c.readByte()
	data = c.readInt()
	if data != 0 {
		speedX = c.readShort()
		speedY = c.readShort()
		speedZ = c.readShort()
	}
	return
}

//Spawn Mob (0x18)
const SpawnMob = 0x18

//Spawn Mob (0x18)
//Server --> Client
func (c *Conn) WriteSpawnMob(eId int32, t int8, x, y, z int32, pitch, headPitch, yaw int8, velocityX, velocityY, velocityZ int16, meta soulsand.EntityMetadata) {
	out := NewByteWriter(1 + 4 + 1 + 4 + 4 + 4 + 1 + 1 + 1 + 2 + 2 + 2)
	out.WriteUByte(SpawnMob)
	out.WriteInt(eId)
	out.WriteByte(t)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteByte(pitch)
	out.WriteByte(headPitch)
	out.WriteByte(yaw)
	out.WriteShort(velocityX)
	out.WriteShort(velocityY)
	out.WriteShort(velocityZ)
	c.Write(out.Bytes())
	(meta.(metadata.Type)).WriteTo(c)
}

//Spawn Mob (0x18)
//Server --> Client
func (c *Conn) ReadSpawnMob() (eId int32, t int8, x, y, z int32, pitch, headPitch, yaw int8, velocityX, velocityY, velocityZ int16, meta soulsand.EntityMetadata) {
	eId = c.readInt()
	t = c.readByte()
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	pitch = c.readByte()
	headPitch = c.readByte()
	yaw = c.readByte()
	velocityX = c.readShort()
	velocityY = c.readShort()
	velocityZ = c.readShort()
	meta = metadata.ReadForm(c)
	return
}

//Spawn Painting (0x19)
const SpawnPainting = 0x19

//Spawn Painting (0x19)
//Server --> Client
func (c *Conn) WriteSpawnPainting(eId int32, title string, x, y, z, direction int32) {
	titleRunes := []rune(title)
	out := NewByteWriter(1 + 4 + 2 + len(titleRunes)*2 + 4 + 4 + 4 + 4)
	out.WriteUByte(SpawnPainting)
	out.WriteString(titleRunes)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteInt(direction)
	c.Write(out.Bytes())
}

//Spawn Painting (0x19)
//Server --> Client
func (c *Conn) ReadSpawnPainting() (eId int32, title string, x, y, z, direction int32) {
	eId = c.readInt()
	title = c.readString()
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	direction = c.readInt()
	return
}

//Spawn Experience Orb (0x1A)
const SpawnExperienceOrb = 0x1A

//Spawn Experience Orb (0x1A)
//Server --> Client
func (c *Conn) WriteSpawnExperienceOrb(eId, x, y, z int32, count int16) {
	out := NewByteWriter(1 + 4 + 4 + 4 + 4 + 2)
	out.WriteUByte(SpawnExperienceOrb)
	out.WriteInt(eId)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteShort(count)
	c.Write(out.Bytes())
}

//Spawn Experience Orb (0x1A)
//Server --> Client
func (c *Conn) ReadSpawnExperienceOrb() (eId, x, y, z int32, count int16) {
	eId = c.readInt()
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	count = c.readShort()
	return
}

//Steer Vehicle (0x1B)
const SteerVehicle = 0x1B

//Steer Vehicle (0x1B)
//Server <-- Client
func (c *Conn) WriteSteerVehicle(sideways, forward float32, jump, unmount bool) {
	out := NewByteWriter(1 + 4 + 4 + 1 + 1)
	out.WriteUByte(SteerVehicle)
	out.WriteFloat(sideways)
	out.WriteFloat(forward)
	out.WriteBool(jump)
	out.WriteBool(unmount)
	c.Write(out.Bytes())
}

//Steer Vehicle (0x1B)
//Server <-- Client
func (c *Conn) ReadSteerVehicle() (sideways, forward float32, jump, unmount bool) {
	sideways = c.readFloat()
	forward = c.readFloat()
	jump = c.readBool()
	unmount = c.readBool()
	return
}

//Entity Velocity (0x1C)
const EntityVelocity = 0x1C

//Entity Velocity (0x1C)
//Server --> Client
func (c *Conn) WriteEntityVelocity(eId int32, velocityX, velocityY, velocityZ int16) {
	out := NewByteWriter(1 + 4 + 2 + 2 + 2)
	out.WriteUByte(EntityVelocity)
	out.WriteInt(eId)
	out.WriteShort(velocityX)
	out.WriteShort(velocityY)
	out.WriteShort(velocityZ)
	c.Write(out.Bytes())
}

//Entity Velocity (0x1C)
//Server --> Client
func (c *Conn) ReadEntityVelocity() (eId int32, velocityX, velocityY, velocityZ int16) {
	eId = c.readInt()
	velocityX = c.readShort()
	velocityY = c.readShort()
	velocityZ = c.readShort()
	return
}

//Destroy Entity (0x1D)
const DestroyEntity = 0x1D

//Destroy Entity (0x1D)
//Server --> Client
func (c *Conn) WriteDestroyEntity(ids []int32) {
	out := NewByteWriter(1 + 1 + len(ids)*4)
	out.WriteUByte(DestroyEntity)
	out.WriteByte(int8(len(ids)))
	for _, i := range ids {
		out.WriteInt(i)
	}
	c.Write(out.Bytes())
}

//Destroy Entity (0x1D)
//Server --> Client
func (c *Conn) ReadDestroyEntity() (ids []int32) {
	length := c.readByte()
	ids = make([]int32, length)
	for i := 0; i < int(length); i++ {
		ids[i] = c.readInt()
	}
	return
}

//Entity (0x1E)
const Entity = 0x1E

//Entity (0x1E)
//Server --> Client
func (c *Conn) WriteEntity(eId int32) {
	out := NewByteWriter(1 + 4)
	out.WriteUByte(Entity)
	out.WriteInt(eId)
	c.Write(out.Bytes())
}

//Entity (0x1E)
//Server --> Client
func (c *Conn) ReadEntity() (eId int32) {
	return c.readInt()
}

//Entity Relative Move (0x1F)
const EntityRelativeMove = 0x1F

//Entity Relative Move (0x1F)
//Server --> Client
func (c *Conn) WriteEntityRelativeMove(eId int32, dX, dY, dZ int8) {
	out := NewByteWriter(1 + 4 + 1 + 1 + 1)
	out.WriteUByte(EntityRelativeMove)
	out.WriteInt(eId)
	out.WriteByte(dX)
	out.WriteByte(dY)
	out.WriteByte(dZ)
	c.Write(out.Bytes())
}

//Entity Relative Move (0x1F)
//Server --> Client
func (c *Conn) ReadEntityRelativeMove() (eId int32, dX, dY, dZ int8) {
	eId = c.readInt()
	dX = c.readByte()
	dY = c.readByte()
	dZ = c.readByte()
	return
}

//Entity Look (0x20)
const EntityLook = 0x20

//Entity Look (0x20)
//Server --> Client
func (c *Conn) WriteEntityLook(eId int32, yaw, pitch int8) {
	out := NewByteWriter(1 + 4 + 1 + 1)
	out.WriteUByte(EntityLook)
	out.WriteInt(eId)
	out.WriteByte(yaw)
	out.WriteByte(pitch)
	c.Write(out.Bytes())
}

//Entity Look (0x20)
//Server --> Client
func (c *Conn) ReadEntityLook() (eId int32, yaw, pitch int8) {
	eId = c.readInt()
	yaw = c.readByte()
	pitch = c.readByte()
	return
}

//Entity Look and Relative Move (0x21)
const EntityLookAndRelativeMove = 0x21

//Entity Look and Relative Move (0x21)
//Server --> Client
func (c *Conn) WriteEntityLookAndRelativeMove(eId int32, dX, dY, dZ, yaw, pitch int8) {
	out := NewByteWriter(1 + 4 + 1 + 1 + 1 + 1 + 1)
	out.WriteUByte(EntityLookAndRelativeMove)
	out.WriteInt(eId)
	out.WriteByte(dX)
	out.WriteByte(dY)
	out.WriteByte(dZ)
	out.WriteByte(yaw)
	out.WriteByte(pitch)
	c.Write(out.Bytes())
}

//Entity Look and Relative Move (0x21)
//Server --> Client
func (c *Conn) ReadEntityLookAndRelativeMove() (eId int32, dX, dY, dZ, yaw, pitch int8) {
	eId = c.readInt()
	dX = c.readByte()
	dY = c.readByte()
	dZ = c.readByte()
	yaw = c.readByte()
	pitch = c.readByte()
	return
}

//Entity Teleport (0x22)
const EntityTeleport = 0x22

//Entity Teleport (0x22)
//Server --> Client
func (c *Conn) WriteEntityTeleport(eId, x, y, z int32, yaw, pitch int8) {
	out := NewByteWriter(1 + 4 + 4 + 4 + 4 + 1 + 1)
	out.WriteUByte(EntityTeleport)
	out.WriteInt(eId)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteByte(yaw)
	out.WriteByte(pitch)
	c.Write(out.Bytes())
}

//Entity Teleport (0x22)
//Server --> Client
func (c *Conn) ReadEntityTeleport() (eId, x, y, z int32, yaw, pitch int8) {
	eId = c.readInt()
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	yaw = c.readByte()
	pitch = c.readByte()
	return
}

//Entity Head Look (0x23)
const EntityHeadLook = 0x23

//Entity Head Look (0x23)
//Server --> Client
func (c *Conn) WriteEntityHeadLook(eId int32, headYaw int8) {
	out := NewByteWriter(1 + 4 + 1)
	out.WriteUByte(EntityHeadLook)
	out.WriteInt(eId)
	out.WriteByte(headYaw)
	c.Write(out.Bytes())
}

//Entity Head Look (0x23)
//Server --> Client
func (c *Conn) ReadEntityHeadLook() (eId int32, headYaw int8) {
	eId = c.readInt()
	headYaw = c.readByte()
	return
}

//Entity Status (0x26)
const EntityStatus = 0x26

//Entity Status (0x26)
//Server --> Client
func (c *Conn) WriteEntityStatus(eId int32, entityStatus int8) {
	out := NewByteWriter(1 + 4 + 1)
	out.WriteUByte(EntityStatus)
	out.WriteInt(eId)
	out.WriteByte(entityStatus)
	c.Write(out.Bytes())
}

//Entity Status (0x26)
//Server --> Client
func (c *Conn) ReadEntityStatus() (eId int32, entityStatus int8) {
	eId = c.readInt()
	entityStatus = c.readByte()
	return
}

//Attach Entity (0x27)
const AttachEntity = 0x27

//Attach Entity (0x27)
//Server --> Client
func (c *Conn) WriteAttachEntity(eId, vId int32, leash bool) {
	out := NewByteWriter(1 + 4 + 4 + 1)
	out.WriteUByte(AttachEntity)
	out.WriteInt(eId)
	out.WriteInt(vId)
	out.WriteBool(leash)
	c.Write(out.Bytes())
}

//Attach Entity (0x27)
//Server --> Client
func (c *Conn) ReadAttachEntity() (eId, vId int32, leash bool) {
	eId = c.readInt()
	vId = c.readInt()
	leash = c.readBool()
	return
}

//Entity Metadata (0x28)
const EntityMetadata = 0x28

//Entity Metadata (0x28)
//Server --> Client
func (c *Conn) WriteEntityMetadata(eId int32, meta soulsand.EntityMetadata) {
	out := NewByteWriter(1 + 4)
	out.WriteUByte(EntityMetadata)
	out.WriteInt(eId)
	c.Write(out.Bytes())
	(meta.(metadata.Type)).WriteTo(c)
}

//Entity Metadata (0x28)
//Server --> Client
func (c *Conn) ReadEntityMetadata() (eId int32, meta soulsand.EntityMetadata) {
	eId = c.readInt()
	meta = metadata.ReadForm(c)
	return
}

//Entity Effect (0x29)
const EntityEffect = 0x29

//Entity Effect (0x29)
//Server --> Client
func (c *Conn) WriteEntityEffect(eId int32, effectId, amplifier int8, duration int16) {
	out := NewByteWriter(1 + 4 + 1 + 1 + 2)
	out.WriteUByte(EntityEffect)
	out.WriteInt(eId)
	out.WriteByte(effectId)
	out.WriteByte(amplifier)
	out.WriteShort(duration)
	c.Write(out.Bytes())
}

//Entity Effect (0x29)
//Server --> Client
func (c *Conn) ReadEntityEffect() (eId int32, effectId, amplifier int8, duration int16) {
	eId = c.readInt()
	effectId = c.readByte()
	amplifier = c.readByte()
	duration = c.readShort()
	return
}

//Remove Entity Effect (0x2A)
const RemoveEntityEffect = 0x2A

//Remove Entity Effect (0x2A)
//Server --> Client
func (c *Conn) WriteRemoveEntityEffect(eId int32, effectId int8) {
	out := NewByteWriter(1 + 4 + 1)
	out.WriteUByte(RemoveEntityEffect)
	out.WriteInt(eId)
	out.WriteByte(effectId)
	c.Write(out.Bytes())
}

//Remove Entity Effect (0x2A)
//Server --> Client
func (c *Conn) ReadRemoveEntityEffect() (eId int32, effectId int8) {
	eId = c.readInt()
	effectId = c.readByte()
	return
}

//Set Experience (0x2B)
const SetExperience = 0x2B

//Set Experience (0x2B)
//Server --> Client
func (c *Conn) WriteSetExperience(experienceBar float32, level, totalExperience int16) {
	out := NewByteWriter(1 + 4 + 2 + 2)
	out.WriteUByte(SetExperience)
	out.WriteFloat(experienceBar)
	out.WriteShort(level)
	out.WriteShort(totalExperience)
	c.Write(out.Bytes())
}

//Set Experience (0x2B)
//Server --> Client
func (c *Conn) ReadSetExperience() (experienceBar float32, level, totalExperience int16) {
	experienceBar = c.readFloat()
	level = c.readShort()
	totalExperience = c.readShort()
	return
}

//Entity Properties (0x2C)
const EntityProperties = 0x2C

//Entity Properties (0x2C)
//Server --> Client
func (c *Conn) WriteEntityProperties(eId int32, na map[string]float64) {
	out := NewByteWriter(1 + 4 + 4)
	out.WriteUByte(EntityProperties)
	out.WriteInt(eId)
	out.WriteInt(int32(len(na)))
	c.Write(out.Bytes())
	for key, value := range na {
		keyRunes := []rune(key)
		out := NewByteWriter(2 + len(keyRunes)*2 + 8 + 2)
		out.WriteString(keyRunes)
		out.WriteDouble(value)
		out.WriteShort(0) // Ignore for now
		c.Write(out.Bytes())
	}
}

//Entity Properties (0x2C)
//Server --> Client
func (c *Conn) ReadEntityProperties() (eId int32, na map[string]float64) {
	eId = c.readInt()
	count := int(c.readInt())
	na = map[string]float64{}
	for i := 0; i < count; i++ {
		na[c.readString()] = c.readDouble()
		co := int(c.readShort())
		for j := 0; j < co; j++ {
			c.readLong()
			c.readLong()
			c.readDouble()
			c.readByte()
		}
	}
	return
}

//Chunk Data (0x33)
const ChunkData = 0x33

//Chunk Data (0x33)
//Server --> Client
func (c *Conn) WriteChunkDataUnload(x, z int32) {
	out := NewByteWriter(1 + 4 + 4 + 1 + 2 + 2 + 4)
	out.WriteUByte(ChunkData)
	out.WriteInt(x)
	out.WriteInt(z)
	out.WriteBool(true)
	c.Write(out.Bytes())
}

//Chunk Data (0x33)
//Server --> Client
func (c *Conn) WriteChunkData(x, z int32, groundUp bool, primaryBit, addBit uint16, data []byte) {
	out := NewByteWriter(1 + 4 + 4 + 1 + 2 + 2 + 4)
	out.WriteUByte(ChunkData)
	out.WriteInt(x)
	out.WriteInt(z)
	out.WriteBool(groundUp)
	out.WriteUShort(primaryBit)
	out.WriteUShort(addBit)
	out.WriteInt(int32(len(data)))
	c.Write(out.Bytes())
	c.Write(data)
}

//Chunk Data (0x33)
//Server --> Client
func (c *Conn) ReadChunkData() (x, z int32, groundUp bool, primaryBit, addBit uint16, data []byte) {
	x = c.readInt()
	z = c.readInt()
	groundUp = c.readBool()
	primaryBit = c.readUShort()
	addBit = c.readUShort()
	size := c.readInt()
	data = make([]byte, size)
	c.Read(data)
	return
}

//Multi Block Change (0x34)
const MultiBlockChange = 0x34

//Multi Block Change (0x34)
//Server --> Client
func (c *Conn) WriteMultiBlockChange(chunkX, chunkZ int32, data []uint32) {
	out := NewByteWriter(1 + 4 + 4 + 2 + 4 + len(data)*4)
	out.WriteUByte(MultiBlockChange)
	out.WriteInt(chunkX)
	out.WriteInt(chunkZ)
	out.WriteShort(int16(len(data)))
	out.WriteInt(int32(len(data) * 4))
	for _, d := range data {
		out.WriteInt(int32(d))
	}
	c.Write(out.Bytes())
}

//Multi Block Change (0x34)
//Server --> Client
func (c *Conn) ReadMultiBlockChange() (chunkX, chunkZ int32, data []uint32) {
	chunkX = c.readInt()
	chunkZ = c.readInt()
	length := int(c.readShort())
	c.readInt()
	data = make([]uint32, length)
	for i := 0; i < length; i++ {
		data[i] = uint32(c.readInt())
	}
	return
}

//Block Change (0x35)
const BlockChange = 0x35

//Block Change (0x35)
//Server --> Client
func (c *Conn) WriteBlockChange(x int32, y byte, z int32, blockType int16, blockMetadata byte) {
	out := NewByteWriter(1 + 4 + 1 + 4 + 2 + 1)
	out.WriteUByte(BlockChange)
	out.WriteInt(x)
	out.WriteUByte(y)
	out.WriteInt(z)
	out.WriteShort(blockType)
	out.WriteUByte(blockMetadata)
	c.Write(out.Bytes())
}

//Block Change (0x35)
//Server --> Client
func (c *Conn) ReadBlockChange() (x int32, y byte, z int32, blockType int16, blockMetadata byte) {
	x = c.readInt()
	y = c.readUByte()
	z = c.readInt()
	blockType = c.readShort()
	blockMetadata = c.readUByte()
	return
}

//Block Action (0x36)
const BlockAction = 0x36

//Block Action (0x36)
//Server --> Client
func (c *Conn) WriteBlockAction(x int32, y int16, z int32, byte1, byte2 byte, blockId int16) {
	out := NewByteWriter(1 + 4 + 2 + 4 + 1 + 1 + 2)
	out.WriteUByte(BlockAction)
	out.WriteInt(x)
	out.WriteShort(y)
	out.WriteInt(z)
	out.WriteUByte(byte1)
	out.WriteUByte(byte2)
	out.WriteShort(blockId)
	c.Write(out.Bytes())
}

//Block Action (0x36)
//Server --> Client
func (c *Conn) ReadBlockAction() (x int32, y int16, z int32, byte1, byte2 byte, blockId int16) {
	x = c.readInt()
	y = c.readShort()
	z = c.readInt()
	byte1 = c.readUByte()
	byte2 = c.readUByte()
	blockId = c.readShort()
	return
}

//Block Break Animation (0x37)
const BlockBreakAnimation = 0x37

//Block Break Animation (0x37)
//Server --> Client
func (c *Conn) WriteBlockBreakAnimation(eId, x, y, z int32, destroyStage int8) {
	out := NewByteWriter(1 + 4 + 4 + 4 + 4 + 1)
	out.WriteUByte(BlockBreakAnimation)
	out.WriteInt(eId)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteByte(destroyStage)
	c.Write(out.Bytes())
}

//Block Break Animation (0x37)
//Server --> Client
func (c *Conn) ReadBlockBreakAnimation() (eId, x, y, z int32, destroyStage int8) {
	eId = c.readInt()
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	destroyStage = c.readByte()
	return
}

//Map Chunk Bulk (0x38)
const MapChunkBulk = 0x38

type ChunkMeta struct {
	x, z         int32
	primary, add uint16
}

//Map Chunk Bulk (0x38)
//Server --> Client
func (c *Conn) WriteMapChunkBulk(skyLight bool, meta []ChunkMeta, data []byte) {
	out := NewByteWriter(1 + 2 + 4 + 1)
	out.WriteUByte(MapChunkBulk)
	out.WriteShort(int16(len(meta)))
	out.WriteInt(int32(len(data)))
	out.WriteBool(skyLight)
	c.Write(out.Bytes())
	c.Write(data)
	out = NewByteWriter(4 + 4 + 2 + 2)
	for _, m := range meta {
		out.WriteInt(m.x)
		out.WriteInt(m.z)
		out.WriteUShort(m.primary)
		out.WriteUShort(m.add)
	}
	c.Write(out.Bytes())
}

//Map Chunk Bulk (0x38)
//Server --> Client
func (c *Conn) ReadMapChunkBulk() (skyLight bool, meta []ChunkMeta, data []byte) {
	metaCount := int(c.readShort())
	length := c.readInt()
	skyLight = c.readBool()
	data = make([]byte, length)
	c.Read(data)
	meta = make([]ChunkMeta, metaCount)
	for i := 0; i < metaCount; i++ {
		meta[i] = ChunkMeta{
			c.readInt(),
			c.readInt(),
			c.readUShort(),
			c.readUShort(),
		}
	}
	return
}

//Explosion (0x3C)
const Explosion = 0x3C

//Explosion (0x3C)
//Server --> Client
func (c *Conn) WriteExplosion(x, y, z float64, radius float32, records []int8, playerMotionX, playerMotionY, playerMotionZ float32) {
	out := NewByteWriter(1 + 8 + 8 + 8 + 4 + 4 + len(records) + 4 + 4 + 4)
	out.WriteUByte(Explosion)
	out.WriteDouble(x)
	out.WriteDouble(y)
	out.WriteDouble(z)
	out.WriteFloat(radius)
	out.WriteInt(int32(len(records) / 3))
	for _, i := range records {
		out.WriteByte(i)
	}
	out.WriteFloat(playerMotionX)
	out.WriteFloat(playerMotionY)
	out.WriteFloat(playerMotionZ)
	c.Write(out.Bytes())
}

//Explosion (0x3C)
//Server --> Client
func (c *Conn) ReadExplosion() (x, y, z float64, radius float32, records []int8, playerMotionX, playerMotionY, playerMotionZ float32) {
	x = c.readDouble()
	y = c.readDouble()
	z = c.readDouble()
	radius = c.readFloat()
	count := int(c.readInt()) * 3
	records = make([]int8, count)
	for i := range records {
		records[i] = c.readByte()
	}
	playerMotionX = c.readFloat()
	playerMotionY = c.readFloat()
	playerMotionZ = c.readFloat()
	return
}

//Sound Or Particle Effect (0x3D)
const SoundOrParticleEffect = 0x3D

//Sound Or Particle Effect (0x3D)
//Server --> Client
func (c *Conn) WriteSoundOrParticleEffect(effectId, x int32, y byte, z, data int32, disableRelativeVolume bool) {
	out := NewByteWriter(1 + 4 + 4 + 1 + 4 + 4 + 1)
	out.WriteUByte(SoundOrParticleEffect)
	out.WriteInt(effectId)
	out.WriteInt(x)
	out.WriteUByte(y)
	out.WriteInt(z)
	out.WriteInt(data)
	out.WriteBool(disableRelativeVolume)
	c.Write(out.Bytes())
}

//Sound Or Particle Effect (0x3D)
//Server --> Client
func (c *Conn) ReadSoundOrParticleEffect() (effectId, x int32, y byte, z, data int32, disableRelativeVolume bool) {
	effectId = c.readInt()
	x = c.readInt()
	y = c.readUByte()
	z = c.readInt()
	data = c.readInt()
	disableRelativeVolume = c.readBool()
	return
}

//Named Sound Effect (0x3E)
const NamedSoundEffect = 0x3E

//Named Sound Effect (0x3E)
//Server --> Client
func (c *Conn) WriteNamedSoundEffect(soundName string, positionX, positionY, positionZ int32, volume float32, pitch int8) {
	soundNameRunes := []rune(soundName)
	out := NewByteWriter(1 + 2 + len(soundNameRunes)*2 + 4 + 4 + 4 + 4 + 1)
	out.WriteUByte(NamedSoundEffect)
	out.WriteString(soundNameRunes)
	out.WriteInt(positionX)
	out.WriteInt(positionY)
	out.WriteInt(positionZ)
	out.WriteFloat(volume)
	out.WriteByte(pitch)
	c.Write(out.Bytes())
}

//Named Sound Effect (0x3E)
//Server --> Client
func (c *Conn) ReadNamedSoundEffect() (soundName string, positionX, positionY, positionZ int32, volume float32, pitch int8) {
	soundName = c.readString()
	positionX = c.readInt()
	positionY = c.readInt()
	positionZ = c.readInt()
	volume = c.readFloat()
	pitch = c.readByte()
	return
}

//Particle (0x3F)
const Particle = 0x3F

//Particle (0x3F)
//Server --> Client
func (c *Conn) WriteParticle(particleName string, x, y, z, offsetX, offsetY, offsetZ, particleSpeed float32, numberOfParticles int32) {
	particleNameRunes := []rune(particleName)
	out := NewByteWriter(1 + 2 + len(particleNameRunes)*2 + 4 + 4 + 4 + 4 + 4 + 4 + 4 + 4)
	out.WriteUByte(Particle)
	out.WriteString(particleNameRunes)
	out.WriteFloat(x)
	out.WriteFloat(y)
	out.WriteFloat(z)
	out.WriteFloat(offsetX)
	out.WriteFloat(offsetY)
	out.WriteFloat(offsetZ)
	out.WriteFloat(particleSpeed)
	out.WriteInt(numberOfParticles)
	c.Write(out.Bytes())
}

//Particle (0x3F)
//Server --> Client
func (c *Conn) ReadParticle() (particleName string, x, y, z, offsetX, offsetY, offsetZ, particleSpeed float32, numberOfParticles int32) {
	particleName = c.readString()
	x = c.readFloat()
	y = c.readFloat()
	z = c.readFloat()
	offsetX = c.readFloat()
	offsetY = c.readFloat()
	offsetZ = c.readFloat()
	particleSpeed = c.readFloat()
	numberOfParticles = c.readInt()
	return
}

//Change Game State (0x46)
const ChangeGameState = 0x46

//Change Game State (0x46)
//Server --> Client
func (c *Conn) WriteChangeGameState(reason, gamemode int8) {
	out := NewByteWriter(1 + 1 + 1)
	out.WriteUByte(ChangeGameState)
	out.WriteByte(reason)
	out.WriteByte(gamemode)
	c.Write(out.Bytes())
}

//Change Game State (0x46)
//Server --> Client
func (c *Conn) ReadChangeGameState() (reason, gamemode int8) {
	reason = c.readByte()
	gamemode = c.readByte()
	return
}

//Spawn Global Entity (0x47)
const SpawnGlobalEntity = 0x47

//Spawn Global Entity (0x47)
//Server --> Client
func (c *Conn) WriteSpawnGlobalEntity(eId int32, t int8, x, y, z int32) {
	out := NewByteWriter(1 + 4 + 1 + 4 + 4 + 4)
	out.WriteUByte(SpawnGlobalEntity)
	out.WriteInt(eId)
	out.WriteByte(t)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	c.Write(out.Bytes())
}

//Spawn Global Entity (0x47)
//Server --> Client
func (c *Conn) ReadSpawnGlobalEntity() (eId int32, t int8, x, y, z int32) {
	eId = c.readInt()
	t = c.readByte()
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	return
}

//Open Window (0x64)
const OpenWindow = 0x64

//Open Window (0x64)
//Server --> Client
func (c *Conn) WriteOpenWindow(windowId, inventoryType int8, windowTitle string, numberOfSlots int8, useTitle bool, entityId int32) {
	windowTitleRunes := []rune(windowTitle)
	out := NewByteWriter(1 + 1 + 1 + 2 + len(windowTitleRunes)*2 + 1 + 1)
	out.WriteUByte(OpenWindow)
	out.WriteByte(windowId)
	out.WriteByte(inventoryType)
	out.WriteString(windowTitleRunes)
	out.WriteByte(numberOfSlots)
	out.WriteBool(useTitle)
	if inventoryType == 0xb {
		out.WriteInt(entityId)
	}
	c.Write(out.Bytes())
}

//Open Window (0x64)
//Server --> Client
func (c *Conn) ReadOpenWindow() (windowId, inventoryType int8, windowTitle string, numberOfSlots int8, useTitle bool, entityId int32) {
	windowId = c.readByte()
	inventoryType = c.readByte()
	windowTitle = c.readString()
	numberOfSlots = c.readByte()
	useTitle = c.readBool()
	if inventoryType == 0xb {
		entityId = c.readInt()
	}
	return
}

//Close Window (0x65)
const CloseWindow = 0x65

//Close Window (0x65)
//Server --> Client
func (c *Conn) WriteCloseWindow(windowId int8) {
	out := NewByteWriter(1 + 1)
	out.WriteUByte(CloseWindow)
	out.WriteByte(windowId)
	c.Write(out.Bytes())
}

//Close Window (0x65)
//Server --> Client
func (c *Conn) ReadCloseWindow() int8 {
	return c.readByte()
}

//Click Window (0x66)
const ClickWindow = 0x66

//Click Window (0x66)
//Server <-- Client
func (c *Conn) WriteClickWindow(windowId int8, slot int16, button int8, actionNumber int16, mode int8) {
	out := NewByteWriter(1 + 1 + 2 + 1 + 2 + 1 + 2)
	out.WriteUByte(ClickWindow)
	out.WriteByte(windowId)
	out.WriteShort(slot)
	out.WriteByte(button)
	out.WriteShort(actionNumber)
	out.WriteByte(mode)
	out.WriteShort(-1)
	c.Write(out.Bytes())
}

//Click Window (0x66)
//Server <-- Client
//Doesn't return the clickedItem as it could be changed by the client
func (c *Conn) ReadClickWindow() (windowId int8, slot int16, button int8, actionNumber int16, mode int8) {
	windowId = c.readByte()
	slot = c.readShort()
	button = c.readByte()
	actionNumber = c.readShort()
	mode = c.readByte()
	c.readSlot()
	return
}

//Set Slot (0x67)
const SetSlot = 0x67

//Set Slot (0x67)
//Server --> Client
func (c *Conn) WriteSetSlot(windowId int8, slot int16, itemstack soulsand.ItemStack) {
	out := NewByteWriter(1 + 1 + 2)
	out.WriteUByte(SetSlot)
	out.WriteByte(windowId)
	out.WriteShort(slot)
	c.Write(out.Bytes())
	c.writeItemstack(itemstack)
}

//Set Slot (0x67)
//Server --> Client
func (c *Conn) ReadSetSlot() (windowId int8, slot int16, itemstack soulsand.ItemStack) {
	windowId = c.readByte()
	slot = c.readShort()
	itemstack = c.readSlot()
	return
}

//Set Window Items (0x68)
const SetWindowItems = 0x68

//Set Window Items (0x68)
//Server --> Client
func (c *Conn) WriteSetWindowItems(windowId int8, itemstacks []soulsand.ItemStack) {
	out := NewByteWriter(1 + 1 + 2)
	out.WriteUByte(SetWindowItems)
	out.WriteByte(windowId)
	out.WriteShort(int16(len(itemstacks)))
	c.Write(out.Bytes())
	for _, itemstack := range itemstacks {
		c.writeItemstack(itemstack)
	}
}

//Set Window Items (0x68)
//Server --> Client
func (c *Conn) ReadSetWindowItems() (windowId int8, itemstacks []soulsand.ItemStack) {
	windowId = c.readByte()
	count := int(c.readShort())
	itemstacks = make([]soulsand.ItemStack, count)
	for i := 0; i < count; i++ {
		itemstacks[i] = c.readSlot()
	}
	return
}

//Update Window Property (0x69)
const UpdateWindowProperty = 0x69

//Update Window Property (0x69)
//Server --> Client
func (c *Conn) WriteUpdateWindowProperty(windowId int8, property, value int16) {
	out := NewByteWriter(1 + 1 + 2 + 2)
	out.WriteUByte(UpdateWindowProperty)
	out.WriteByte(windowId)
	out.WriteShort(property)
	out.WriteShort(value)
	c.Write(out.Bytes())
}

//Update Window Property (0x69)
//Server --> Client
func (c *Conn) ReadUpdateWindowProperty() (windowId int8, property, value int16) {
	windowId = c.readByte()
	property = c.readShort()
	value = c.readShort()
	return
}

//Confirm Transaction (0x6A)
const ConfirmTransaction = 0x6A

//Confirm Transaction (0x6A)
//Server <--> Client
func (c *Conn) WriteConfirmTransaction(windowId int8, actionNumber int16, accepted bool) {
	out := NewByteWriter(1 + 1 + 2 + 1)
	out.WriteUByte(ConfirmTransaction)
	out.WriteByte(windowId)
	out.WriteShort(actionNumber)
	out.WriteBool(accepted)
	c.Write(out.Bytes())
}

//Confirm Transaction (0x6A)
//Server <--> Client
func (c *Conn) ReadConfirmTransaction() (windowId int8, actionNumber int16, accepted bool) {
	windowId = c.readByte()
	actionNumber = c.readShort()
	accepted = c.readBool()
	return
}

//Creative Inventory Action (0x6B)
const CreativeInventoryAction = 0x6B

//Creative Inventory Action (0x6B)
//Server <--> Client
func (c *Conn) WriteCreativeInventoryAction(slot int16, itemstack soulsand.ItemStack) {
	out := NewByteWriter(1 + 2)
	out.WriteUByte(CreativeInventoryAction)
	out.WriteShort(slot)
	c.Write(out.Bytes())
	c.writeItemstack(itemstack)
}

//Creative Inventory Action (0x6B)
//Server <--> Client
func (c *Conn) ReadCreativeInventoryAction() (slot int16, itemstack soulsand.ItemStack) {
	slot = c.readShort()
	itemstack = c.readSlot()
	return
}

//Enchant Item (0x6C)
const EnchantItem = 0x6C

//Enchant Item (0x6C)
//Server <-- Client
func (c *Conn) WriteEnchantItem(windowId, enchantment int8) {
	out := NewByteWriter(1 + 1 + 1)
	out.WriteUByte(EnchantItem)
	out.WriteByte(windowId)
	out.WriteByte(enchantment)
	c.Write(out.Bytes())
}

//Enchant Item (0x6C)
//Server <-- Client
func (c *Conn) ReadEnchantItem() (windowId, enchantment int8) {
	windowId = c.readByte()
	enchantment = c.readByte()
	return
}

//Update Sign (0x82)
const UpdateSign = 0x82

//Update Sign (0x82)
//Server <--> Client
func (c *Conn) WriteUpdateSign(x int32, y int16, z int32, text1, text2, text3, text4 string) {
	text1Runes := []rune(text1)
	text2Runes := []rune(text2)
	text3Runes := []rune(text3)
	text4Runes := []rune(text4)
	out := NewByteWriter(1 + 4 + 2 + 4 + 2 + len(text1Runes)*2 + 2 + len(text2Runes)*2 + 2 + len(text3Runes)*2 + 2 + len(text4Runes)*2)
	out.WriteUByte(UpdateSign)
	out.WriteInt(x)
	out.WriteShort(y)
	out.WriteInt(z)
	out.WriteString(text1Runes)
	out.WriteString(text2Runes)
	out.WriteString(text3Runes)
	out.WriteString(text4Runes)
	c.Write(out.Bytes())
}

//Update Sign (0x82)
//Server <--> Client
func (c *Conn) ReadUpdateSign() (x int32, y int16, z int32, text1, text2, text3, text4 string) {
	x = c.readInt()
	y = c.readShort()
	z = c.readInt()
	text1 = c.readString()
	text2 = c.readString()
	text3 = c.readString()
	text4 = c.readString()
	return
}

//Item Data (0x83)
const ItemData = 0x83

//Item Data (0x83)
//Server --> Client
func (c *Conn) WriteItemData(itemType, itemID int16, text []byte) {
	out := NewByteWriter(1 + 2 + 2 + 2)
	out.WriteUByte(ItemData)
	out.WriteShort(itemType)
	out.WriteShort(itemID)
	out.WriteShort(int16(len(text)))
	c.Write(out.Bytes())
	c.Write(text)
}

//Item Data (0x83)
//Server --> Client
func (c *Conn) ReadItemData() (itemType, itemID int16, text []byte) {
	itemType = c.readShort()
	itemID = c.readShort()
	length := c.readShort()
	text = make([]byte, length)
	c.Read(text)
	return
}

//Update Tile Entity (0x84)
const UpdateTileEntity = 0x84

//Update Tile Entity (0x84)
//Server --> Client
func (c *Conn) WriteUpdateTileEntity(x int32, y int16, z int32, action int8, data nbt.Type) {
	var buf bytes.Buffer
	if data != nil {
		data.WriteTo(&buf, "")
	}
	bytes := buf.Bytes()
	out := NewByteWriter(1 + 4 + 2 + 4 + 1 + 2 + len(bytes))
	out.WriteUByte(UpdateTileEntity)
	out.WriteInt(x)
	out.WriteShort(y)
	out.WriteInt(z)
	out.WriteByte(action)
	out.WriteShort(int16(len(bytes)))
	for _, b := range bytes {
		out.WriteUByte(b)
	}
	c.Write(out.Bytes())
}

//Update Tile Entity (0x84)
//Server --> Client
func (c *Conn) ReadUpdateTileEntity() (x int32, y int16, z int32, action int8, data nbt.Type) {
	x = c.readInt()
	y = c.readShort()
	z = c.readInt()
	action = c.readByte()
	c.readShort()
	d, err := nbt.Parse(c)
	if err != nil {
		runtime.Goexit()
	}
	data = d
	return
}

//Tile Editor Open (0x85)
const TileEditorOpen = 0x85

//Tile Editor Open (0x85)
//Server --> Client
func (c *Conn) WriteTileEditorOpen(tileId int8, x, y, z int32) {
	out := NewByteWriter(1 + 1 + 4 + 4 + 4)
	out.WriteUByte(TileEditorOpen)
	out.WriteByte(tileId)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	c.Write(out.Bytes())
}

//Tile Editor Open (0x85)
//Server --> Client
func (c *Conn) ReadTileEditorOpen() (tileId int8, x, y, z int32) {
	tileId = c.readByte()
	x = c.readInt()
	y = c.readInt()
	z = c.readInt()
	return
}

//Increment Statistic (0xC8)
const IncrementStatistic = 0xC8

//Increment Statistic (0xC8)
//Server --> Client
func (c *Conn) WriteIncrementStatistic(statisticID, amount int32) {
	out := NewByteWriter(1 + 4 + 4)
	out.WriteUByte(IncrementStatistic)
	out.WriteInt(statisticID)
	out.WriteInt(amount)
	c.Write(out.Bytes())
}

//Increment Statistic (0xC8)
//Server --> Client
func (c *Conn) ReadIncrementStatistic() (statisticID, amount int32) {
	statisticID = c.readInt()
	amount = c.readInt()
	return
}

//Player List Item (0xC9)
const PlayerListItem = 0xC9

//Player List Item (0xC9)
//Server --> Client
func (c *Conn) WritePlayerListItem(playerName string, online bool, ping int16) {
	playerNameRunes := []rune(playerName)
	out := NewByteWriter(1 + 2 + len(playerNameRunes)*2 + 1 + 2)
	out.WriteUByte(PlayerListItem)
	out.WriteString(playerNameRunes)
	out.WriteBool(online)
	out.WriteShort(ping)
	c.Write(out.Bytes())
}

//Player List Item (0xC9)
//Server --> Client
func (c *Conn) ReadPlayerListItem() (playerName string, online bool, ping int16) {
	playerName = c.readString()
	online = c.readBool()
	ping = c.readShort()
	return
}

//Player Abilities (0xCA)
const PlayerAbilities = 0xCA

//Player Abilities (0xCA)
//Server <--> Client
func (c *Conn) WritePlayerAbilities(flags int8, flyingSpeed, walkingSpeed float32) {
	out := NewByteWriter(1 + 1 + 4 + 4)
	out.WriteUByte(PlayerAbilities)
	out.WriteByte(flags)
	out.WriteFloat(flyingSpeed)
	out.WriteFloat(walkingSpeed)
	c.Write(out.Bytes())
}

//Player Abilities (0xCA)
//Server <--> Client
func (c *Conn) ReadPlayerAbilities() (flags int8, flyingSpeed, walkingSpeed float32) {
	flags = c.readByte()
	flyingSpeed = c.readFloat()
	walkingSpeed = c.readFloat()
	return
}

//Tab-complete (0xCB)
const TabComplete = 0xCB

//Tab-complete (0xCB)
//Server <--> Client
func (c *Conn) WriteTabComplete(text string) {
	textRunes := []rune(text)
	out := NewByteWriter(1 + 2 + len(text)*2)
	out.WriteUByte(TabComplete)
	out.WriteString(textRunes)
	c.Write(out.Bytes())
}

//Tab-complete (0xCB)
//Server <--> Client
func (c *Conn) ReadTabComplete() string {
	return c.readString()
}

//Client Settings (0xCC)
const ClientSettings = 0xCC

//Client Settings (0xCC)
//Server <-- Client
func (c *Conn) WriteClientSettings(locale string, viewDistance, chatFlags, difficulty int8, showCape bool) {
	localeRunes := []rune(locale)
	out := NewByteWriter(1 + 2 + len(localeRunes)*2 + 1 + 1 + 1 + 1)
	out.WriteUByte(ClientSettings)
	out.WriteString(localeRunes)
	out.WriteByte(viewDistance)
	out.WriteByte(chatFlags)
	out.WriteByte(difficulty)
	out.WriteBool(showCape)
	c.Write(out.Bytes())
}

//Client Settings (0xCC)
//Server <-- Client
func (c *Conn) ReadClientSettings() (locale string, viewDistance, chatFlags, difficulty int8, showCape bool) {
	locale = c.readString()
	viewDistance = c.readByte()
	chatFlags = c.readByte()
	difficulty = c.readByte()
	showCape = c.readBool()
	return
}

//Client Statuses (0xCD)
const ClientStatuses = 0xCD

//Client Statuses (0xCD)
//Server <-- Client
func (c *Conn) WriteClientStatuses(payload int8) {
	out := NewByteWriter(1 + 1)
	out.WriteUByte(ClientStatuses)
	out.WriteByte(payload)
	c.Write(out.Bytes())
}

//Client Statuses (0xCD)
//Server <-- Client
func (c *Conn) ReadClientStatuses() int8 {
	return c.readByte()
}

//Scoreboard Objective (0xCE)
const ScoreboardObjective = 0xCE

//Scoreboard Objective (0xCE)
//Server --> Client
func (c *Conn) WriteScoreboardObjective(objectiveName, objectiveValue string, mode byte) {
	objectiveNameRunes := []rune(objectiveName)
	objectiveValueRunes := []rune(objectiveValue)
	out := NewByteWriter(1 + 2 + len(objectiveNameRunes)*2 + 2 + len(objectiveValueRunes)*2 + 1)
	out.WriteUByte(ScoreboardObjective)
	out.WriteString(objectiveNameRunes)
	out.WriteString(objectiveValueRunes)
	out.WriteUByte(mode)
	c.Write(out.Bytes())
}

//Scoreboard Objective (0xCE)
//Server --> Client
func (c *Conn) ReadScoreboardObjective() (objectiveName, objectiveValue string, mode byte) {
	objectiveName = c.readString()
	objectiveValue = c.readString()
	mode = c.readUByte()
	return
}

//Update Score (0xCF)
const UpdateScore = 0xCF

//Update Score (0xCF)
//Server --> Client
func (c *Conn) WriteUpdateScore(itemName string, mode byte, scoreName string, value int32) {
	itemNameRunes := []rune(itemName)
	scoreNameRunes := []rune(scoreName)
	out := NewByteWriter(1 + 2 + len(itemNameRunes)*2 + 1 + 2 + len(scoreNameRunes)*2 + 4)
	out.WriteUByte(UpdateScore)
	out.WriteString(itemNameRunes)
	out.WriteUByte(mode)
	out.WriteString(scoreNameRunes)
	out.WriteInt(value)
	c.Write(out.Bytes())
}

//Update Score (0xCF)
//Server --> Client
func (c *Conn) ReadUpdateScore() (itemName string, mode byte, scoreName string, value int32) {
	itemName = c.readString()
	mode = c.readUByte()
	scoreName = c.readString()
	value = c.readInt()
	return
}

//Display Scoreboard (0xD0)
const DisplayScoreboard = 0xD0

//Display Scoreboard (0xD0)
//Server --> Client
func (c *Conn) WriteDisplayScoreboard(position int8, scoreName string) {
	scoreNameRunes := []rune(scoreName)
	out := NewByteWriter(1 + 1 + 2 + len(scoreNameRunes)*2)
	out.WriteUByte(DisplayScoreboard)
	out.WriteByte(position)
	out.WriteString(scoreNameRunes)
	c.Write(out.Bytes())
}

//Display Scoreboard (0xD0)
//Server --> Client
func (c *Conn) ReadDisplayScoreboard() (position int8, scoreName string) {
	position = c.readByte()
	scoreName = c.readString()
	return
}

//Teams (0xD1)
const Teams = 0xD1

//Teams (0xD1)
//Server --> Client
func (c *Conn) WriteTeamCreate(teamName, teamDisplayName, teamPrefix, teamSuffix string, teamSettings byte, players []string) {
	teamNameRunes := []rune(teamName)
	teamDisplayNameRunes := []rune(teamDisplayName)
	teamPrefixRunes := []rune(teamPrefix)
	teamSuffixRunes := []rune(teamSuffix)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1 + 2 + len(teamDisplayNameRunes)*2 + 2 + len(teamPrefixRunes)*2 + 2 + len(teamSuffixRunes)*2 + 1 + 2)
	out.WriteUByte(Teams)
	out.WriteString(teamNameRunes)
	out.WriteByte(0)
	out.WriteString(teamDisplayNameRunes)
	out.WriteString(teamPrefixRunes)
	out.WriteString(teamSuffixRunes)
	out.WriteUByte(teamSettings)
	out.WriteShort(int16(len(players)))
	c.Write(out.Bytes())
	for _, p := range players {
		runes := []rune(p)
		out = NewByteWriter(2 + len(runes)*2)
		out.WriteString(runes)
		c.Write(out.Bytes())
	}
}

//Teams (0xD1)
//Server --> Client
func (c *Conn) WriteTeamRemove(teamName string) {
	teamNameRunes := []rune(teamName)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1)
	out.WriteUByte(Teams)
	out.WriteString(teamNameRunes)
	out.WriteByte(1)
	c.Write(out.Bytes())
}

//Teams (0xD1)
//Server --> Client
func (c *Conn) WriteTeamUpdate(teamName, teamDisplayName, teamPrefix, teamSuffix string, teamSettings byte) {
	teamNameRunes := []rune(teamName)
	teamDisplayNameRunes := []rune(teamDisplayName)
	teamPrefixRunes := []rune(teamPrefix)
	teamSuffixRunes := []rune(teamSuffix)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1 + 2 + len(teamDisplayNameRunes)*2 + 2 + len(teamPrefixRunes)*2 + 2 + len(teamSuffixRunes)*2 + 1)
	out.WriteUByte(Teams)
	out.WriteString(teamNameRunes)
	out.WriteByte(2)
	out.WriteString(teamDisplayNameRunes)
	out.WriteString(teamPrefixRunes)
	out.WriteString(teamSuffixRunes)
	out.WriteUByte(teamSettings)
	c.Write(out.Bytes())
}

//Teams (0xD1)
//Server --> Client
func (c *Conn) WriteTeamAddPlayers(teamName string, players []string) {
	teamNameRunes := []rune(teamName)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1 + 2)
	out.WriteUByte(Teams)
	out.WriteString(teamNameRunes)
	out.WriteByte(3)
	out.WriteShort(int16(len(players)))
	c.Write(out.Bytes())
	for _, p := range players {
		runes := []rune(p)
		out = NewByteWriter(2 + len(runes)*2)
		out.WriteString(runes)
		c.Write(out.Bytes())
	}
}

//Teams (0xD1)
//Server --> Client
func (c *Conn) WriteTeamRemovePlayers(teamName string, players []string) {
	teamNameRunes := []rune(teamName)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1 + 2)
	out.WriteUByte(Teams)
	out.WriteString(teamNameRunes)
	out.WriteByte(4)
	out.WriteShort(int16(len(players)))
	c.Write(out.Bytes())
	for _, p := range players {
		runes := []rune(p)
		out = NewByteWriter(2 + len(runes)*2)
		out.WriteString(runes)
		c.Write(out.Bytes())
	}
}

//Teams (0xD1)
//Server --> Client
func (c *Conn) ReadTeams() (teamName string, mode byte, displayName, prefix, suffix string, friendlyFire byte, players []string) {
	teamName = c.readString()
	mode = c.readUByte()
	if mode == 0 || mode == 2 {
		displayName = c.readString()
		prefix = c.readString()
		suffix = c.readString()
		friendlyFire = c.readUByte()
	}
	if mode == 0 || mode == 3 || mode == 4 {
		count := int(c.readShort())
		players = make([]string, count)
		for i := 0; i < count; i++ {
			players[i] = c.readString()
		}
	}
	return
}

//Plugin Message (0xFA)
const PluginMessage = 0xFA

//Plugin Message (0xFA)
//Server <--> Client
func (c *Conn) WritePluginMessage(channel string, data []byte) {
	channelRunes := []rune(channel)
	out := NewByteWriter(1 + 2 + len(channel)*2 + 2)
	out.WriteUByte(PluginMessage)
	out.WriteString(channelRunes)
	out.WriteShort(int16(len(data)))
	c.Write(out.Bytes())
	c.Write(data)
}

//Plugin Message (0xFA)
//Server <--> Client
func (c *Conn) ReadPluginMessage() (channel string, data []byte) {
	channel = c.readString()
	l := c.readShort()
	data = make([]byte, l)
	c.Read(data)
	return
}

//Encryption Key Response (0xFC)
const EncryptionKeyResponse = 0xFC

//Encryption Key Response (0xFC)
//Server <--> Client
func (c *Conn) WriteEncryptionKeyResponse(sharedSecret, verifyTokenResponse []byte) {
	out := NewByteWriter(1 + 2)
	out.WriteUByte(EncryptionKeyResponse)
	out.WriteShort(int16(len(sharedSecret)))
	c.Write(out.Bytes())
	c.Write(sharedSecret)
	out = NewByteWriter(2)
	out.WriteShort(int16(len(verifyTokenResponse)))
	c.Write(out.Bytes())
	c.Write(verifyTokenResponse)
}

//Encryption Key Response (0xFC)
//Server <--> Client
func (c *Conn) ReadEncryptionKeyResponse() (sharedSecret, verifyTokenResponse []byte) {
	l := c.readShort()
	sharedSecret = make([]byte, l)
	c.Read(sharedSecret)
	l = c.readShort()
	verifyTokenResponse = make([]byte, l)
	c.Read(verifyTokenResponse)
	return
}

//Encryption Key Request (0xFD)
const EncryptionKeyRequest = 0xFD

//Encryption Key Request (0xFD)
//Server --> Client
func (c *Conn) WriteEncryptionKeyRequest(serverID string, publicKey []byte, verifyToken []byte) {
	serverIDRunes := []rune(serverID)
	out := NewByteWriter(1 + 2 + len(serverIDRunes)*2 + 2)
	out.WriteUByte(EncryptionKeyRequest)
	out.WriteString(serverIDRunes)
	out.WriteShort(int16(len(publicKey)))
	c.Write(out.Bytes())
	c.Write(publicKey)
	out = NewByteWriter(2)
	out.WriteShort(int16(len(verifyToken)))
	c.Write(out.Bytes())
	c.Write(verifyToken)
}

//Encryption Key Request (0xFD)
//Server --> Client
func (c *Conn) ReadEncryptionKeyRequest() (serverID string, publicKey []byte, verifyToken []byte) {
	serverID = c.readString()
	publicKey = make([]byte, c.readShort())
	c.Read(publicKey)
	verifyToken = make([]byte, c.readShort())
	c.Read(verifyToken)
	return
}

//Disconnect/Kick (0xFF)
const Disconnect = 0xFF

//Disconnect/Kick (0xFF)
//Server <--> Client
func (c *Conn) WriteDisconnect(reason string) {
	reasonRunes := []rune(reason)
	out := NewByteWriter(1 + 2 + len(reason)*2)
	out.WriteUByte(Disconnect)
	out.WriteString(reasonRunes)
	c.Write(out.Bytes())
}

//Disconnect/Kick (0xFF)
//Server <--> Client
func (c *Conn) ReadDisconnect() string {
	return c.readString()
}
