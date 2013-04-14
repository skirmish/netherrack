package event

import (
	"bitbucket.org/Thinkofdeath/soulsand"
)

var _ soulsand.EventPlayerMessage = &EventPlayerMessage{}
var _ soulsand.EventPlayerJoin = &EventPlayerJoin{}

type EventPlayerMessage struct {
	Event

	player  soulsand.SyncPlayer
	message string
}

func NewMessage(player soulsand.SyncPlayer, message string) *EventPlayerMessage {
	return &EventPlayerMessage{
		Event:   Event{},
		player:  player,
		message: message,
	}
}

func (m *EventPlayerMessage) SetMessage(msg string) {
	m.message = msg
}

func (m *EventPlayerMessage) GetMessage() string {
	return m.message
}

func (m *EventPlayerMessage) GetPlayer() soulsand.SyncPlayer {
	return m.player
}

type EventPlayerJoin struct {
	Event

	player soulsand.SyncPlayer
	Reason string
}

func NewJoin(player soulsand.SyncPlayer, reason string) *EventPlayerJoin {
	return &EventPlayerJoin{
		Event:  Event{},
		player: player,
		Reason: reason,
	}
}

func (j *EventPlayerJoin) GetPlayer() soulsand.SyncPlayer {
	return j.player
}

func (j *EventPlayerJoin) Disconnect(reason string) {
	j.Reason = reason
	j.Cancel()
}

type EventPlayerLeave struct {
	Event

	player soulsand.SyncPlayer
}

func NewLeave(player soulsand.SyncPlayer) *EventPlayerLeave {
	return &EventPlayerLeave{
		Event:  Event{},
		player: player,
	}
}

func (l *EventPlayerLeave) GetPlayer() soulsand.SyncPlayer {
	return l.player
}
