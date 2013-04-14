package event

import (
	"bitbucket.org/Thinkofdeath/soulsand"
)

var _ soulsand.EventPlayerMessage = &PlayerMessage{}
var _ soulsand.EventPlayerJoin = &PlayerJoin{}
var _ soulsand.EventPlayerLeave = &PlayerLeave{}

type PlayerMessage struct {
	Event

	player  soulsand.SyncPlayer
	message string
}

func NewMessage(player soulsand.SyncPlayer, message string) (string, *PlayerMessage) {
	return "EventPlayerMessage",&PlayerMessage{
		Event:   Event{},
		player:  player,
		message: message,
	}
}

func (m *PlayerMessage) SetMessage(msg string) {
	m.message = msg
}

func (m *PlayerMessage) GetMessage() string {
	return m.message
}

func (m *PlayerMessage) GetPlayer() soulsand.SyncPlayer {
	return m.player
}

type PlayerJoin struct {
	Event

	player soulsand.SyncPlayer
	Reason string
}

func NewJoin(player soulsand.SyncPlayer, reason string) (string, *PlayerJoin) {
	return "EventPlayerJoin", &PlayerJoin{
		Event:  Event{},
		player: player,
		Reason: reason,
	}
}

func (j *PlayerJoin) GetPlayer() soulsand.SyncPlayer {
	return j.player
}

func (j *PlayerJoin) Disconnect(reason string) {
	j.Reason = reason
	j.Cancel()
}

type PlayerLeave struct {
	Event

	player soulsand.SyncPlayer
}

func NewLeave(player soulsand.SyncPlayer) (string, *PlayerLeave) {
	return "EventPlayerLeave", &PlayerLeave{
		Event:  Event{},
		player: player,
	}
}

func (l *PlayerLeave) GetPlayer() soulsand.SyncPlayer {
	return l.player
}
