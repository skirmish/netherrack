package event

import (
	"Soulsand"
)

var _ Soulsand.EventPlayerMessage = &Message{}

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
