package entity

import (

)

func (e *Entity) GetPositionSync() (float64, float64, float64) {
	return e.Position.X, e.Position.Y, e.Position.Z
}

func (e *Entity) SetPositionSync(x, y, z float64) {
	e.Position.X, e.Position.Y, e.Position.Z = x, y, z
}