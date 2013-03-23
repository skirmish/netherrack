package event

import (
	"Soulsand"
)

var _ Soulsand.EventPlayerMessage = &Message{}
var _ Soulsand.EventPlayerJoin = &Join{}

type Message struct {
	Event

	player  Soulsand.Player
	message string
}

func NewMessage(player Soulsand.Player, message string) *Message {
	return &Message{
		Event:   Event{},
		player:  player,
		message: message,
	}
}

func (m *Message) GetMessage() string {
	return m.message
}

func (m *Message) GetPlayer() Soulsand.Player {
	return m.player
}

type Join struct {
	Event

	player  Soulsand.SyncPlayer
}

func NewJoin(player Soulsand.SyncPlayer) *Join {
	return &Join{
		Event:   Event{},
		player:  player,
	}
}

func (j *Join) GetPlayer() Soulsand.SyncPlayer {
	return j.player
}