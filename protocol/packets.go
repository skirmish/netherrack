package protocol

import (
	"bytes"
	"compress/gzip"
	"github.com/thinkofdeath/netherrack/entity/metadata"
	"github.com/thinkofdeath/netherrack/items"
	"github.com/thinkofdeath/netherrack/nbt"
	"github.com/thinkofdeath/soulsand"
)

//Keep Alive (0x00)

func (c *Conn) WriteKeepAlive(id int32) {
	out := NewByteWriter(1 + 4)
	out.WriteUByte(0x00)
	out.WriteInt(id)
	c.Write(out.Bytes())
}

func (c *Conn) ReadKeepAlive() int32 {
	return c.readInt()
}

//Login Request (0x01)

func (c *Conn) WriteLoginRequest(eId int32, levelType string, gamemode, dimension, difficuly, maxPlayers int8) {
	levelTypeRunes := []rune(levelType)
	out := NewByteWriter(1 + 4 + 2 + len(levelTypeRunes)*2 + 1 + 1 + 1 + 1 + 1)
	out.WriteUByte(0x01)
	out.WriteInt(eId)
	out.WriteString(levelTypeRunes)
	out.WriteByte(gamemode)
	out.WriteByte(dimension)
	out.WriteByte(difficuly)
	out.WriteByte(0)
	out.WriteByte(maxPlayers)
	c.Write(out.Bytes())
}

//Handshake (0x02)

func (c *Conn) ReadHandshake() (protoVersion byte, username, host string, port int32) {
	protoVersion = c.readUByte()
	username = c.readString()
	host = c.readString()
	port = c.readInt()
	return
}

//Chat Message (0x03)

func (c *Conn) WriteChatMessage(message string) {
	messageRunes := []rune(message)
	out := NewByteWriter(1 + 2 + len(messageRunes)*2)
	out.WriteUByte(0x03)
	out.WriteString(messageRunes)
	c.Write(out.Bytes())
}

func (c *Conn) ReadChatMessage() string {
	return c.readString()
}

//Time Update (0x04)

func (c *Conn) WriteTimeUpdate(age, time int64) {
	out := NewByteWriter(1 + 8 + 8)
	out.WriteUByte(0x04)
	out.WriteLong(age)
	out.WriteLong(time)
	c.Write(out.Bytes())
}

//Entity Equipment (0x05)

func (c *Conn) WriteEntityEquipment(eId int32, slot int16, itemstack soulsand.ItemStack) {
	var out *byteWriter
	slotData := itemstack.(*items.ItemStack)
	if slotData.ID == -1 {
		out = NewByteWriter(1 + 4 + 2 + 2)
	} else {
		out = NewByteWriter(1 + 4 + 2 + 2 + 1 + 2 + 2)
	}
	dataLength := int16(-1)
	var data []byte
	if slotData.Tag != nil {
		var buf bytes.Buffer
		gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
		slotData.Tag.WriteTo(gz, false)
		gz.Close()
		data = buf.Bytes()
	}
	out.WriteUByte(0x05)
	out.WriteInt(eId)
	out.WriteShort(slot)
	out.WriteShort(slotData.ID)
	if slotData.ID != -1 {
		out.WriteUByte(slotData.Count)
		out.WriteShort(slotData.Damage)
		out.WriteShort(dataLength)
	}
	c.Write(out.Bytes())
	if slotData.Tag != nil && slotData.ID != -1 {
		c.Write(data)
	}
}

//Spawn Position (0x06)

func (c *Conn) WriteSpawnPosition(x, y, z int32) {
	out := NewByteWriter(1 + 4 + 4 + 4)
	out.WriteUByte(0x06)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	c.Write(out.Bytes())
}

//Use Entity (0x07)

func (c *Conn) ReadUseEntity() (user, target int32, mouseButton bool) {
	user = c.readInt()
	target = c.readInt()
	mouseButton = c.readBool()
	return
}

//Update Health (0x08)

func (c *Conn) WriteUpdateHealth(health, food int16, foodSatiration float32) {
	out := NewByteWriter(1 + 2 + 2 + 4)
	out.WriteUByte(0x08)
	out.WriteShort(health)
	out.WriteShort(food)
	out.WriteFloat(foodSatiration)
	c.Write(out.Bytes())
}

//Respawn (0x09)

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

//Player (0x0A)

func (c *Conn) ReadPlayer() bool {
	return c.readBool()
}

//Player Position (0x0B)

func (c *Conn) ReadPlayerPosition() (x, y, stance, z float64, onGround bool) {
	x = c.readDouble()
	y = c.readDouble()
	stance = c.readDouble()
	z = c.readDouble()
	onGround = c.readBool()
	return
}

//Player Look (0x0C)

func (c *Conn) ReadPlayerLook() (yaw, pitch float32, onGround bool) {
	yaw = c.readFloat()
	pitch = c.readFloat()
	onGround = c.readBool()
	return
}

//Player Position and Look (0x0D)

func (c *Conn) WritePlayerPositionLook(x, y, z, stance float64, yaw, pitch float32, onGround bool) {
	out := NewByteWriter(1 + 8 + 8 + 8 + 8 + 4 + 4 + 1)
	out.WriteUByte(0x0D)
	out.WriteDouble(x)
	out.WriteDouble(stance)
	out.WriteDouble(y)
	out.WriteDouble(z)
	out.WriteFloat(yaw)
	out.WriteFloat(pitch)
	out.WriteBool(onGround)
	c.Write(out.Bytes())
}

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

func (c *Conn) ReadPlayerDigging() (status int8, x int32, y byte, z int32, face int8) {
	status = c.readByte()
	x = c.readInt()
	y = c.readUByte()
	z = c.readInt()
	face = c.readByte()
	return
}

//Player Block Placement (0x0F)
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

func (c *Conn) WriteHeldItemChange(slotID int16) {
	out := NewByteWriter(1 + 2)
	out.WriteUByte(0x10)
	out.WriteShort(slotID)
	c.Write(out.Bytes())
}

func (c *Conn) ReadHeldItemChange() int16 {
	return c.readShort()
}

//Use Bed (0x11)

func (c *Conn) WriteUseBed(eId, x int32, y byte, z int32) {
	out := NewByteWriter(1 + 4 + 1 + 4 + 1 + 4)
	out.WriteUByte(0x11)
	out.WriteInt(eId)
	out.WriteUByte(0)
	out.WriteInt(x)
	out.WriteUByte(y)
	out.WriteInt(z)
	c.Write(out.Bytes())
}

//Animation (0x12)

func (c *Conn) WriteAnimation(eId int32, animation int8) {
	out := NewByteWriter(1 + 4 + 1)
	out.WriteUByte(0x12)
	out.WriteInt(eId)
	out.WriteByte(animation)
	c.Write(out.Bytes())
}

func (c *Conn) ReadAnimation() (eId int32, animation int8) {
	eId = c.readInt()
	animation = c.readByte()
	return
}

//Entity Action (0x13)

func (c *Conn) ReadEntityAction() (eId int32, actionID int8) {
	eId = c.readInt()
	actionID = c.readByte()
	return
}

//Spawn Named Entity (0x14)

func (c *Conn) WriteSpawnNamedEntity(eID int32, playerName string, x, y, z int32, yaw, pitch int8, currentItem int16, meta soulsand.EntityMetadata) {
	playerNameRunes := []rune(playerName)
	out := NewByteWriter(1 + 4 + 2 + len(playerNameRunes)*2 + 4 + 4 + 4 + 1 + 1 + 2)
	out.WriteUByte(0x14)
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

//Collect Item (0x16)

func (c *Conn) WriteCollectItem(collectedEId, collectorEId int32) {
	out := NewByteWriter(1 + 4 + 4)
	out.WriteUByte(0x16)
	out.WriteInt(collectedEId)
	out.WriteInt(collectorEId)
	c.Write(out.Bytes())
}

//Spawn Object/Vehicle (0x17)

func (c *Conn) WriteSpawnObjectVehicle(eId int32, t int8, x, y, z int32, pitch, yaw int8, data int32, speedX, speedY, speedZ int16) {
	var out *byteWriter
	if data != 0 {
		out = NewByteWriter(1 + 4 + 1 + 4 + 4 + 4 + 1 + 1 + 4 + 2 + 2 + 2)
	} else {
		out = NewByteWriter(1 + 4 + 1 + 4 + 4 + 4 + 1 + 1 + 4)
	}
	out.WriteUByte(0x17)
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

//Spawn Mob (0x18)

func (c *Conn) WriteSpawnMob(eId int32, t int8, x, y, z int32, pitch, headPitch, yaw int8, velocityX, velocityY, velocityZ int16, meta soulsand.EntityMetadata) {
	out := NewByteWriter(1 + 4 + 1 + 4 + 4 + 4 + 1 + 1 + 1 + 2 + 2 + 2)
	out.WriteUByte(0x18)
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

//Spawn Painting (0x19)

func (c *Conn) WriteSpawnPainting(eId int32, title string, x, y, z, direction int32) {
	titleRunes := []rune(title)
	out := NewByteWriter(1 + 4 + 2 + len(titleRunes)*2 + 4 + 4 + 4 + 4)
	out.WriteUByte(0x19)
	out.WriteString(titleRunes)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteInt(direction)
	c.Write(out.Bytes())
}

//Spawn Experience Orb (0x1A)

func (c *Conn) WriteSpawnExperienceOrb(eId, x, y, z int32, count int16) {
	out := NewByteWriter(1 + 4 + 4 + 4 + 4 + 2)
	out.WriteUByte(0x1A)
	out.WriteInt(eId)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteShort(count)
	c.Write(out.Bytes())
}

//Entity Velocity (0x1C)

func (c *Conn) WriteEntityVelocity(eId int32, velocityX, velocityY, velocityZ int16) {
	out := NewByteWriter(1 + 4 + 2 + 2 + 2)
	out.WriteUByte(0x1C)
	out.WriteInt(eId)
	out.WriteShort(velocityX)
	out.WriteShort(velocityY)
	out.WriteShort(velocityZ)
	c.Write(out.Bytes())
}

//Destroy Entity (0x1D)

func (c *Conn) WriteDestroyEntity(ids []int32) {
	out := NewByteWriter(1 + 1 + len(ids)*4)
	out.WriteUByte(0x1D)
	out.WriteByte(int8(len(ids)))
	for _, i := range ids {
		out.WriteInt(i)
	}
	c.Write(out.Bytes())
}

//Entity (0x1E)

func (c *Conn) WriteEntity(eId int32) {
	out := NewByteWriter(1 + 4)
	out.WriteUByte(0x1E)
	out.WriteInt(eId)
	c.Write(out.Bytes())
}

//Entity Relative Move (0x1F)

func (c *Conn) WriteEntityRelativeMove(eId int32, dX, dY, dZ int8) {
	out := NewByteWriter(1 + 4 + 1 + 1 + 1)
	out.WriteUByte(0x1F)
	out.WriteInt(eId)
	out.WriteByte(dX)
	out.WriteByte(dY)
	out.WriteByte(dZ)
	c.Write(out.Bytes())
}

//Entity Look (0x20)

func (c *Conn) WriteEntityLook(eId int32, yaw, pitch int8) {
	out := NewByteWriter(1 + 4 + 1 + 1)
	out.WriteUByte(0x20)
	out.WriteInt(eId)
	out.WriteByte(yaw)
	out.WriteByte(pitch)
	c.Write(out.Bytes())
}

//Entity Look and Relative Move (0x21)

func (c *Conn) WriteEntityLookAndRelativeMove(eId int32, dX, dY, dZ, yaw, pitch int8) {
	out := NewByteWriter(1 + 4 + 1 + 1 + 1 + 1 + 1)
	out.WriteUByte(0x21)
	out.WriteInt(eId)
	out.WriteByte(dX)
	out.WriteByte(dY)
	out.WriteByte(dZ)
	out.WriteByte(yaw)
	out.WriteByte(pitch)
	c.Write(out.Bytes())
}

//Entity Teleport (0x22)

func (c *Conn) WriteEntityTeleport(eId, x, y, z int32, yaw, pitch int8) {
	out := NewByteWriter(1 + 4 + 4 + 4 + 4 + 1 + 1)
	out.WriteUByte(0x22)
	out.WriteInt(eId)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteByte(yaw)
	out.WriteByte(pitch)
	c.Write(out.Bytes())
}

//Entity Head Look (0x23)

func (c *Conn) WriteEntityHeadLook(eId int32, headYaw int8) {
	out := NewByteWriter(1 + 4 + 1)
	out.WriteUByte(0x23)
	out.WriteInt(eId)
	out.WriteByte(headYaw)
	c.Write(out.Bytes())
}

//Entity Status (0x26)

func (c *Conn) WriteEntityStatus(eId int32, entityStatus int8) {
	out := NewByteWriter(1 + 4 + 1)
	out.WriteUByte(0x26)
	out.WriteInt(eId)
	out.WriteByte(entityStatus)
	c.Write(out.Bytes())
}

//Attach Entity (0x27)

func (c *Conn) WriteAttachEntity(eId, vId int32) {
	out := NewByteWriter(1 + 4 + 4)
	out.WriteUByte(0x27)
	out.WriteInt(eId)
	out.WriteInt(vId)
	c.Write(out.Bytes())
}

//Entity Metadata (0x28)

func (c *Conn) WriteEntityMetadata(eId int32, meta soulsand.EntityMetadata) {
	out := NewByteWriter(1 + 4)
	out.WriteUByte(0x28)
	out.WriteInt(eId)
	c.Write(out.Bytes())
	(meta.(metadata.Type)).WriteTo(c)
}

//Entity Effect (0x29)

func (c *Conn) WriteEntityEffect(eId int32, effectId, amplifier int8, duration int16) {
	out := NewByteWriter(1 + 4 + 1 + 1 + 2)
	out.WriteUByte(0x29)
	out.WriteInt(eId)
	out.WriteByte(effectId)
	out.WriteByte(amplifier)
	out.WriteShort(duration)
	c.Write(out.Bytes())
}

//Remove Entity Effect (0x2A)

func (c *Conn) WriteRemoveEntityEffect(eId int32, effectId int8) {
	out := NewByteWriter(1 + 4 + 1)
	out.WriteUByte(0x2A)
	out.WriteInt(eId)
	out.WriteByte(effectId)
	c.Write(out.Bytes())
}

//Set Experience (0x2B)

func (c *Conn) WriteSetExperience(experienceBar float32, level, totalExperience int16) {
	out := NewByteWriter(1 + 4 + 2 + 2)
	out.WriteUByte(0x2B)
	out.WriteFloat(experienceBar)
	out.WriteShort(level)
	out.WriteShort(totalExperience)
	c.Write(out.Bytes())
}

//Chunk Data (0x33)

func (c *Conn) WriteChunkDataUnload(x, z int32) {
	out := NewByteWriter(1 + 4 + 4 + 1 + 2 + 2 + 4)
	out.WriteUByte(0x33)
	out.WriteInt(x)
	out.WriteInt(z)
	out.WriteBool(true)
	c.Write(out.Bytes())
}

//Multi Block Change (0x34)

func (c *Conn) WriteMultiBlockChange(chunkX, chunkZ int32, data []uint32) {
	out := NewByteWriter(1 + 4 + 4 + 2 + 4 + len(data)*4)
	out.WriteUByte(0x34)
	out.WriteInt(chunkX)
	out.WriteInt(chunkZ)
	out.WriteShort(int16(len(data)))
	out.WriteInt(int32(len(data) * 4))
	for _, d := range data {
		out.WriteInt(int32(d))
	}
	c.Write(out.Bytes())
}

//Block Change (0x35)

func (c *Conn) WriteBlockChange(x int32, y byte, z int32, blockType int16, blockMetadata byte) {
	out := NewByteWriter(1 + 4 + 1 + 4 + 2 + 1)
	out.WriteUByte(0x35)
	out.WriteInt(x)
	out.WriteUByte(y)
	out.WriteInt(z)
	out.WriteShort(blockType)
	out.WriteUByte(blockMetadata)
	c.Write(out.Bytes())
}

//Block Action (0x36)

func (c *Conn) WriteBlockAction(x int32, y int16, z int32, byte1, byte2 byte, blockId int16) {
	out := NewByteWriter(1 + 4 + 2 + 4 + 1 + 1 + 2)
	out.WriteUByte(0x36)
	out.WriteInt(x)
	out.WriteShort(y)
	out.WriteInt(z)
	out.WriteUByte(byte1)
	out.WriteUByte(byte2)
	out.WriteShort(blockId)
	c.Write(out.Bytes())
}

//Block Break Animation (0x37)

func (c *Conn) WriteBlockBreakAnimation(eId, x, y, z int32, destroyStage int8) {
	out := NewByteWriter(1 + 4 + 4 + 4 + 4 + 1)
	out.WriteUByte(0x37)
	out.WriteInt(eId)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	out.WriteByte(destroyStage)
	c.Write(out.Bytes())
}

//Map Chunk Bulk (0x38)

//Explosion (0x3C)

func (c *Conn) WriteExplosion(x, y, z float64, radius float32, records []int8, playerMotionX, playerMotionY, playerMotionZ float32) {
	out := NewByteWriter(1 + 8 + 8 + 8 + 4 + 4 + len(records) + 4 + 4 + 4)
	out.WriteUByte(0x3C)
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

//Sound Or Particle Effect (0x3D)

func (c *Conn) WriteSoundOrParticleEffect(effectId, x int32, y byte, z, data int32, disableRelativeVolume bool) {
	out := NewByteWriter(1 + 4 + 4 + 1 + 4 + 4 + 1)
	out.WriteUByte(0x3D)
	out.WriteInt(effectId)
	out.WriteInt(x)
	out.WriteUByte(y)
	out.WriteInt(z)
	out.WriteInt(data)
	out.WriteBool(disableRelativeVolume)
	c.Write(out.Bytes())
}

//Named Sound Effect (0x3E)

func (c *Conn) WriteNamedSoundEffect(soundName string, positionX, positionY, positionZ int32, volume float32, pitch int8) {
	soundNameRunes := []rune(soundName)
	out := NewByteWriter(1 + 2 + len(soundNameRunes)*2 + 4 + 4 + 4 + 4 + 1)
	out.WriteUByte(0x3E)
	out.WriteString(soundNameRunes)
	out.WriteInt(positionX)
	out.WriteInt(positionY)
	out.WriteInt(positionZ)
	out.WriteFloat(volume)
	out.WriteByte(pitch)
	c.Write(out.Bytes())
}

//Particle (0x3F)

func (c *Conn) WriteParticle(particleName string, x, y, z, offsetX, offsetY, offsetZ, particleSpeed float32, numberOfParticles int32) {
	particleNameRunes := []rune(particleName)
	out := NewByteWriter(1 + 2 + len(particleNameRunes)*2 + 4 + 4 + 4 + 4 + 4 + 4 + 4 + 4)
	out.WriteUByte(0x3F)
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

//Change Game State (0x46)

func (c *Conn) WriteChangeGameState(reason, gamemode int8) {
	out := NewByteWriter(1 + 1 + 1)
	out.WriteUByte(0x46)
	out.WriteByte(reason)
	out.WriteByte(gamemode)
	c.Write(out.Bytes())
}

//Spawn Global Entity (0x47)

func (c *Conn) WriteSpawnGlobalEntity(eId int32, t int8, x, y, z int32) {
	out := NewByteWriter(1 + 4 + 1 + 4 + 4 + 4)
	out.WriteUByte(0x47)
	out.WriteInt(eId)
	out.WriteByte(t)
	out.WriteInt(x)
	out.WriteInt(y)
	out.WriteInt(z)
	c.Write(out.Bytes())
}

//Open Window (0x64)

func (c *Conn) WriteOpenWindow(windowId, inventoryType int8, windowTitle string, numberOfSlots int8, useTitle bool) {
	windowTitleRunes := []rune(windowTitle)
	out := NewByteWriter(1 + 1 + 1 + 2 + len(windowTitleRunes)*2 + 1 + 1)
	out.WriteUByte(0x64)
	out.WriteByte(windowId)
	out.WriteByte(inventoryType)
	out.WriteString(windowTitleRunes)
	out.WriteByte(numberOfSlots)
	out.WriteBool(useTitle)
	c.Write(out.Bytes())
}

//Close Window (0x65)

func (c *Conn) WriteCloseWindow(windowId int8) {
	out := NewByteWriter(1 + 1)
	out.WriteUByte(0x65)
	out.WriteByte(windowId)
	c.Write(out.Bytes())
}

func (c *Conn) ReadCloseWindow() int8 {
	return c.readByte()
}

//Click Window (0x66)
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

func (c *Conn) WriteSetSlot(windowId int8, slot int16, itemstack soulsand.ItemStack) {
	var out *byteWriter
	slotData := itemstack.(*items.ItemStack)
	if slotData.ID == -1 {
		out = NewByteWriter(1 + 1 + 2 + 2)
	} else {
		out = NewByteWriter(1 + 1 + 2 + 2 + 1 + 2 + 2)
	}
	dataLength := int16(-1)
	var data []byte
	if slotData.Tag != nil {
		var buf bytes.Buffer
		gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
		slotData.Tag.WriteTo(gz, false)
		gz.Close()
		data = buf.Bytes()
	}
	out.WriteUByte(0x67)
	out.WriteByte(windowId)
	out.WriteShort(slot)
	out.WriteShort(slotData.ID)
	if slotData.ID != -1 {
		out.WriteUByte(slotData.Count)
		out.WriteShort(slotData.Damage)
		out.WriteShort(dataLength)
	}
	c.Write(out.Bytes())
	if slotData.Tag != nil && slotData.ID != -1 {
		c.Write(data)
	}
}

//Set Window Items (0x68)

func (c *Conn) WriteSetWindowItems(windowId int8, itemstacks []soulsand.ItemStack) {
	out := NewByteWriter(1 + 1 + 2)
	out.WriteUByte(0x68)
	out.WriteByte(windowId)
	out.WriteShort(int16(len(itemstacks)))
	c.Write(out.Bytes())
	for _, itemstack := range itemstacks {
		slotData := itemstack.(*items.ItemStack)
		if slotData.ID == -1 {
			out = NewByteWriter(2)
		} else {
			out = NewByteWriter(2 + 1 + 2 + 2)
		}
		dataLength := int16(-1)
		var data []byte
		if slotData.Tag != nil {
			var buf bytes.Buffer
			gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
			slotData.Tag.WriteTo(gz, false)
			gz.Close()
			data = buf.Bytes()
		}
		if slotData.ID != -1 {
			out.WriteUByte(slotData.Count)
			out.WriteShort(slotData.Damage)
			out.WriteShort(dataLength)
		}
		c.Write(out.Bytes())
		if slotData.Tag != nil && slotData.ID != -1 {
			c.Write(data)
		}
	}
}

//Update Window Property (0x69)

func (c *Conn) WriteUpdateWindowProperty(windowId int8, property, value int16) {
	out := NewByteWriter(1 + 1 + 2 + 2)
	out.WriteUByte(0x69)
	out.WriteByte(windowId)
	out.WriteShort(property)
	out.WriteShort(value)
	c.Write(out.Bytes())
}

//Confirm Transaction (0x6A)

func (c *Conn) WriteConfirmTransaction(windowId int8, actionNumber int16, accepted bool) {
	out := NewByteWriter(1 + 1 + 2 + 1)
	out.WriteUByte(0x6A)
	out.WriteByte(windowId)
	out.WriteShort(actionNumber)
	out.WriteBool(accepted)
	c.Write(out.Bytes())
}

func (c *Conn) ReadConfirmTransaction() (windowId int8, actionNumber int16, accepted bool) {
	windowId = c.readByte()
	actionNumber = c.readShort()
	accepted = c.readBool()
	return
}

//Creative Inventory Action (0x6B)

func (c *Conn) WriteCreativeInventoryAction(slot int16, itemstack soulsand.ItemStack) {
	var out *byteWriter
	slotData := itemstack.(*items.ItemStack)
	if slotData.ID == -1 {
		out = NewByteWriter(1 + 2 + 2)
	} else {
		out = NewByteWriter(1 + 2 + 2 + 1 + 2 + 2)
	}
	dataLength := int16(-1)
	var data []byte
	if slotData.Tag != nil {
		var buf bytes.Buffer
		gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
		slotData.Tag.WriteTo(gz, false)
		gz.Close()
		data = buf.Bytes()
	}
	out.WriteUByte(0x6B)
	out.WriteShort(slot)
	out.WriteShort(slotData.ID)
	if slotData.ID != -1 {
		out.WriteUByte(slotData.Count)
		out.WriteShort(slotData.Damage)
		out.WriteShort(dataLength)
	}
	c.Write(out.Bytes())
	if slotData.Tag != nil && slotData.ID != -1 {
		c.Write(data)
	}
}

func (c *Conn) ReadCreativeInventoryAction() (slot int16, itemstack soulsand.ItemStack) {
	slot = c.readShort()
	itemstack = c.readSlot()
	return
}

//Enchant Item (0x6C)

func (c *Conn) ReadEnchantItem() (windowId, enchantment int8) {
	windowId = c.readByte()
	enchantment = c.readByte()
	return
}

//Update Sign (0x82)

func (c *Conn) WriteUpdateSign(x int32, y int16, z int32, text1, text2, text3, text4 string) {
	text1Runes := []rune(text1)
	text2Runes := []rune(text2)
	text3Runes := []rune(text3)
	text4Runes := []rune(text4)
	out := NewByteWriter(1 + 4 + 2 + 4 + 2 + len(text1Runes)*2 + 2 + len(text2Runes)*2 + 2 + len(text3Runes)*2 + 2 + len(text4Runes)*2)
	out.WriteUByte(0x82)
	out.WriteInt(x)
	out.WriteShort(y)
	out.WriteInt(z)
	out.WriteString(text1Runes)
	out.WriteString(text2Runes)
	out.WriteString(text3Runes)
	out.WriteString(text4Runes)
	c.Write(out.Bytes())
}

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

func (c *Conn) WriteItemData(itemType, itemID int16, text []byte) {
	out := NewByteWriter(1 + 2 + 2 + 2)
	out.WriteUByte(0x83)
	out.WriteShort(itemType)
	out.WriteShort(itemID)
	out.WriteShort(int16(len(text)))
	c.Write(out.Bytes())
	c.Write(text)
}

//Update Tile Entity (0x84)

func (c *Conn) WriteUpdateTileEntity(x int32, y int16, z int32, action int8, data *nbt.Compound) {
	var buf bytes.Buffer
	if data != nil {
		data.WriteTo(&buf, false)
	}
	bytes := buf.Bytes()
	out := NewByteWriter(1 + 4 + 2 + 4 + 1 + 2 + len(bytes))
	out.WriteUByte(0x84)
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

//Increment Statistic (0xC8)

func (c *Conn) WriteIncrementStatistic(statisticID int32, amount int8) {
	out := NewByteWriter(1 + 4 + 1)
	out.WriteUByte(0xC8)
	out.WriteInt(statisticID)
	out.WriteByte(amount)
	c.Write(out.Bytes())
}

//Player List Item (0xC9)

func (c *Conn) WritePlayerListItem(playerName string, online bool, ping int16) {
	playerNameRunes := []rune(playerName)
	out := NewByteWriter(1 + 2 + len(playerNameRunes)*2 + 1 + 2)
	out.WriteUByte(0xC9)
	out.WriteString(playerNameRunes)
	out.WriteBool(online)
	out.WriteShort(ping)
	c.Write(out.Bytes())
}

//Player Abilities (0xCA)

func (c *Conn) WritePlayerAbilities(flags, flyingSpeed, walkingSpeed int8) {
	out := NewByteWriter(1 + 1 + 1 + 1)
	out.WriteUByte(0xCA)
	out.WriteByte(flags)
	out.WriteByte(flyingSpeed)
	out.WriteByte(walkingSpeed)
	c.Write(out.Bytes())
}

func (c *Conn) ReadPlayerAbilities() (flags, flyingSpeed, walkingSpeed int8) {
	flags = c.readByte()
	flyingSpeed = c.readByte()
	walkingSpeed = c.readByte()
	return
}

//Tab-complete (0xCB)

func (c *Conn) WriteTabComplete(text string) {
	textRunes := []rune(text)
	out := NewByteWriter(1 + 2 + len(text)*2)
	out.WriteUByte(0xCB)
	out.WriteString(textRunes)
	c.Write(out.Bytes())
}

func (c *Conn) ReadTabComplete() string {
	return c.readString()
}

//Client Settings (0xCC)

func (c *Conn) ReadClientSettings() (locale string, viewDistance, chatFlags, difficulty int8, showCape bool) {
	locale = c.readString()
	viewDistance = c.readByte()
	chatFlags = c.readByte()
	difficulty = c.readByte()
	showCape = c.readBool()
	return
}

//Client Statuses (0xCD)

func (c *Conn) ReadClientStatuses() int8 {
	return c.readByte()
}

//Scoreboard Objective (0xCE)

func (c *Conn) WriteScoreboardObjective(objectiveName, objectiveValue string, mode byte) {
	objectiveNameRunes := []rune(objectiveName)
	objectiveValueRunes := []rune(objectiveValue)
	out := NewByteWriter(1 + 2 + len(objectiveNameRunes)*2 + 2 + len(objectiveValueRunes)*2 + 1)
	out.WriteUByte(0xCE)
	out.WriteString(objectiveNameRunes)
	out.WriteString(objectiveValueRunes)
	out.WriteUByte(mode)
	c.Write(out.Bytes())
}

//Update Score (0xCF)

func (c *Conn) WriteUpdateScore(itemName string, mode byte, scoreName string, value int32) {
	itemNameRunes := []rune(itemName)
	scoreNameRunes := []rune(scoreName)
	out := NewByteWriter(1 + 2 + len(itemNameRunes)*2 + 1 + 2 + len(scoreNameRunes)*2 + 4)
	out.WriteUByte(0xCF)
	out.WriteString(itemNameRunes)
	out.WriteUByte(mode)
	out.WriteString(scoreNameRunes)
	out.WriteInt(value)
	c.Write(out.Bytes())
}

//Display Scoreboard (0xD0)

func (c *Conn) WriteDisplayScoreboard(position int8, scoreName string) {
	scoreNameRunes := []rune(scoreName)
	out := NewByteWriter(1 + 1 + 2 + len(scoreNameRunes)*2)
	out.WriteUByte(0xD0)
	out.WriteByte(position)
	out.WriteString(scoreNameRunes)
	c.Write(out.Bytes())
}

//Teams (0xD1)

func (c *Conn) WriteTeamCreate(teamName, teamDisplayName, teamPrefix, teamSuffix string, teamSettings byte, players []string) {
	teamNameRunes := []rune(teamName)
	teamDisplayNameRunes := []rune(teamDisplayName)
	teamPrefixRunes := []rune(teamPrefix)
	teamSuffixRunes := []rune(teamSuffix)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1 + 2 + len(teamDisplayNameRunes)*2 + 2 + len(teamPrefixRunes)*2 + 2 + len(teamSuffixRunes)*2 + 1 + 2)
	out.WriteUByte(0xD1)
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

func (c *Conn) WriteTeamRemove(teamName string) {
	teamNameRunes := []rune(teamName)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1)
	out.WriteUByte(0xD1)
	out.WriteString(teamNameRunes)
	out.WriteByte(1)
	c.Write(out.Bytes())
}

func (c *Conn) WriteTeamUpdate(teamName, teamDisplayName, teamPrefix, teamSuffix string, teamSettings byte) {
	teamNameRunes := []rune(teamName)
	teamDisplayNameRunes := []rune(teamDisplayName)
	teamPrefixRunes := []rune(teamPrefix)
	teamSuffixRunes := []rune(teamSuffix)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1 + 2 + len(teamDisplayNameRunes)*2 + 2 + len(teamPrefixRunes)*2 + 2 + len(teamSuffixRunes)*2 + 1)
	out.WriteUByte(0xD1)
	out.WriteString(teamNameRunes)
	out.WriteByte(2)
	out.WriteString(teamDisplayNameRunes)
	out.WriteString(teamPrefixRunes)
	out.WriteString(teamSuffixRunes)
	out.WriteUByte(teamSettings)
	c.Write(out.Bytes())
}

func (c *Conn) WriteTeamAddPlayers(teamName string, players []string) {
	teamNameRunes := []rune(teamName)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1 + 2)
	out.WriteUByte(0xD1)
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

func (c *Conn) WriteTeamRemovePlayers(teamName string, players []string) {
	teamNameRunes := []rune(teamName)
	out := NewByteWriter(1 + 2 + len(teamNameRunes)*2 + 1 + 2)
	out.WriteUByte(0xD1)
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

//Plugin Message (0xFA)

func (c *Conn) WritePluginMessage(channel string, data []byte) {
	channelRunes := []rune(channel)
	out := NewByteWriter(1 + 2 + len(channel)*2 + 2)
	out.WriteUByte(0xFA)
	out.WriteString(channelRunes)
	out.WriteShort(int16(len(data)))
	c.Write(out.Bytes())
	c.Write(data)
}

func (c *Conn) ReadPluginMessage() (channel string, data []byte) {
	channel = c.readString()
	l := c.readShort()
	data = make([]byte, l)
	c.Read(data)
	return
}

//Encryption Key Response (0xFC)

func (c *Conn) WriteEncryptionKeyResponse(sharedSecret, verifyTokenResponse []byte) {
	out := NewByteWriter(1 + 2)
	out.WriteUByte(0xFC)
	out.WriteShort(int16(len(sharedSecret)))
	c.Write(out.Bytes())
	c.Write(sharedSecret)
	out = NewByteWriter(2)
	out.WriteShort(int16(len(verifyTokenResponse)))
	c.Write(out.Bytes())
	c.Write(verifyTokenResponse)
}

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

func (c *Conn) WriteEncryptionKeyRequest(serverID string, publicKey []byte, verifyToken []byte) {
	serverIDRunes := []rune(serverID)
	out := NewByteWriter(1 + 2 + len(serverIDRunes)*2 + 2)
	out.WriteUByte(0xFD)
	out.WriteString(serverIDRunes)
	out.WriteShort(int16(len(publicKey)))
	c.Write(out.Bytes())
	c.Write(publicKey)
	out = NewByteWriter(2)
	out.WriteShort(int16(len(verifyToken)))
	c.Write(out.Bytes())
	c.Write(verifyToken)
}

//Disconnect/Kick (0xFF)

func (c *Conn) WriteDisconnect(reason string) {
	reasonRunes := []rune(reason)
	out := NewByteWriter(1 + 2 + len(reason)*2)
	out.WriteUByte(0xFF)
	out.WriteString(reasonRunes)
	c.Write(out.Bytes())
}

func (c *Conn) ReadDisconnect() string {
	return c.readString()
}
