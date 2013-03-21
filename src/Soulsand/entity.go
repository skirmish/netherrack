package Soulsand

import ()

type Entity interface {
	//Returns the ID of the entity
	GetID() int32
	//Sets the entity's location
	SetPosition(x, y, z float64)
	//Returns the entity's location
	GetPosition() (float64, float64, float64)
	//Sets the entity's velocity
	SetVelocity(x, y, z float64)
	//Returns the entity's velocity
	GetVelocity() (float64, float64, float64)
	//Returns if the entity is alive/active
	IsAlive() bool
	//Remove the entity from the game
	Remove()
}

type SyncEntity interface {
	//Returns the entity's location
	GetPositionSync() (float64, float64, float64)
	//Sets the entity's location
	SetPositionSync(x, y, z float64)
}