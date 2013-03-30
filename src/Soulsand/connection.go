package Soulsand

import (
	"io"
)

type UnsafeConnection interface {
	GetInputStream() io.Reader
	GetOutputStream() io.Writer
	WriteDisconnect(reason string)
	WritePluginMessage(channel string, data []byte)
	WriteTeamRemovePlayers(name string, players []string)
	WriteTeamAddPlayers(name string, players []string)
	WriteUpdateTeam(name string, dName string, pre string, suf string, ff int8)
	WriteRemoveTeam(name string)
	WriteCreateTeam(name string, dName string, pre string, suf string, ff bool, players []string)
	WriteDisplayScoreboard(position int8, name string)
	WriteUpdateScore(name string, sName string, value int32)
	WriteRemoveScore(name string)
	WriteCreateScoreboard(name, display string, remove bool)
	WriteTabComplete(text string)
	WritePlayerAbilities(flags, fSpeed, wSpeed byte)
	WritePlayerListItem(name string, online bool, ping int16)
	WriteIncrementStatistic(statID int32, amount byte)
	//WriteUpdateTileEntity(x int32, y int16, z int32, action byte, data *nbt.Compound)
	WriteItemData(iType, iID int16, data []byte)
	WriteUpdateSign(x int32, y int16, z int32, text1, text2, text3, text4 string)
	WriteCreativeInventoryAction(slot int16, item ItemStack)
	WriteConfirmTransaction(wID byte, aNum int16, accepted bool)
	WriteUpdateWindowProperty(wID byte, prop, val int16)
	WriteSetWindowItems(wID byte, slots []ItemStack)
	WriteSetSlot(wID byte, slot int16, data ItemStack)
	WriteCloseWindow(wID byte)
	WriteOpenWindow(wID, iType byte, title string, slots byte, useTitle bool)
	WriteSpawnGlobalEntity(eID int32, t byte, x, y, z int32)
	WriteChangeGameState(res, gMode byte)
	WriteParticle(a string, b, cc, d, e, f, g, h float32, i int32)
	WriteNameSoundEffect(name string, posX, posY, posZ int32, vol float32, pitch byte)
	WriteSoundParticleEffect(eff, x int32, y byte, z, data int32, rVol bool)
	WriteExplosion(x, y, z float64, radius float32, records []byte, velX, velY, velZ float32)
	WriteBlockBreakAnimation(eID, x, y, z int32, stage byte)
	WriteBlockAction(x int32, y int16, z int32, b1, b2 byte, bID int16)
	WriteBlockChange(x int32, y byte, z int32, bType int16, bMeta byte)
	WriteMultiBlockChange(cx, cz int32, blocks []uint32)
	WriteChunkDataUnload(x, z int32)
	WriteSetExperience(bar float32, level, exp int16)
	WriteRemoveEntityEffect(eID int32, eff int8)
	WriteEntityEffect(eID int32, eff int8, amp int8, duration int16)
	//WriteEntityMetadata(eID int32, metadata map[byte]entity.MetadataItem)
	WriteAttachEntity(eID int32, vID int32)
	WriteEntityStatus(eID int32, status int8)
	WriteEntityHeadLook(eID int32, hYaw int8)
	WriteEntityTeleport(eID, x, y, z int32, yaw, pitch int8)
	WriteEntityLookRelativeMove(eID int32, dX, dY, dZ int8, yaw, pitch int8)
	WriteEntityLook(eID int32, yaw, pitch int8)
	WriteEntityRelativeMove(eID int32, dX, dY, dZ int8)
	WriteEntity(eID int32)
	WriteDestroyEntity(eIDS []int32)
	WriteEntityVelocity(eID int32, velX, velY, velZ int16)
	WriteSpawnExperienceOrb(eID, x, y, z int32, count int16)
	WriteSpawnPainting(eID int32, title string, x, y, z, dir int32)
	//WriteSpawnMob(eID int32, t int8, x, y, z int32, yaw, pitch, hYaw int8, velX, velY, velZ int16, metadata map[byte]entity.MetadataItem)
	WriteSpawnObjectSpeed(eID int32, t int8, x, y, z int32, yaw, pitch int8, data int32, speedX, speedY, speedZ int16)
	WriteSpawnObject(eID int32, t int8, x, y, z int32, yaw, pitch int8)
	WriteCollectItem(collectedEID, collectorEID int32)
	//WriteSpawnNamedEntity(eID int32, name string, x, y, z int32, yaw, pitch int8, curItem int16, metadata map[byte]entity.MetadataItem)
	WriteAnimation(eID int32, ani int8)
	WriteUseBed(eID int32, x int32, y byte, z int32)
	WriteHeldItemChange(slotID int16)
	WriteRespawn(dim int32, diff, gMode int8, height int16, lType string)
	WriteUpdateHealth(health, food int16, fSat float32)
	//WriteEntityEquipment(eID int32, slot int16, slotData *items.Slot)
	WriteTimeUpdate(age, time int64)
	WriteChatMessage(msg string)
	WriteKeepAlive(id int32)
	WritePlayerPositionLook(x, y, z, stance float64, yaw, pitch float32, onGround bool)
	WriteSpawnPosition(x, y, z int32)
	WriteLoginRequest(eID int32, lType string, gMode, dim, diff, mP int8)
	WriteKeyRequest(token []byte, serverID string)
}
