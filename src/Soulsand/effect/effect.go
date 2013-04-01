package effect

import ()

type Type int32

const (
	//Sound
	RandomClick        Type = 1000
	RandomClick2       Type = 1001
	RandomBow          Type = 1002
	RandomDoor         Type = 1003
	RandomFizz         Type = 1004
	MusicDisc          Type = 1005
	MobGhastCharge     Type = 1007
	MobGhastFireball   Type = 1008
	MobZombieWood      Type = 1010
	MobZombieMetal     Type = 1011
	MobZombieWoodbreak Type = 1012
	MobWitherSpawn     Type = 1013

	//Particle
	Smoke        Type = 2000
	BlockBreak   Type = 2001
	SplashPotion Type = 2002
	EyeOfEnder   Type = 2003
	MobSpawn     Type = 2004
)
