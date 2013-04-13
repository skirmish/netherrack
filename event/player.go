package event

import (
	"bitbucket.org/Thinkofdeath/soulsand"
)

var _ soulsand.EventPlayerMessage = &Message{}
var _ soulsand.EventPlayerJoin = &Join{}

type Message struct {
	Event

	player  soulsand.SyncPlayer
	message string
}

func NewMessage(player soulsand.SyncPlayer, message string) *Message {
	return &Message{
		Event:   Event{},
		player:  player,
		message: message,
	}
}

func (m *Message) SetMessage(msg string) {
	m.message = msg
}

func (m *Message) GetMessage() string {
	return m.message
}

func (m *Message) GetPlayer() soulsand.SyncPlayer {
	return m.player
}

type Join struct {
	Event

	player soulsand.SyncPlayer
	Reason string
}

func NewJoin(player soulsand.SyncPlayer, reason string) *Join {
	return &Join{
		Event:  Event{},
		player: player,
		Reason: reason,
	}
}

func (j *Join) GetPlayer() soulsand.SyncPlayer {
	return j.player
}

func (j *Join) Disconnect(reason string) {
	j.Reason = reason
	j.Cancel()
}

type Leave struct {
	Event

	player soulsand.SyncPlayer
}

func NewLeave(player soulsand.SyncPlayer) *Leave {
	return &Leave{
		Event:  Event{},
		player: player,
	}
}

func (l *Leave) GetPlayer() soulsand.SyncPlayer {
	return l.player
}
