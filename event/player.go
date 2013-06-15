package event

import (
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/chat"
)

var _ soulsand.EventPlayerMessage = &PlayerMessage{}
var _ soulsand.EventPlayerJoin = &PlayerJoin{}
var _ soulsand.EventPlayerLeave = &PlayerLeave{}
var _ soulsand.EventPlayerLeftClick = &PlayerLeftClick{}
var _ soulsand.EventPlayerBlockPlace = &PlayerBlockPlace{}
var _ soulsand.EventPlayerRightClick = &PlayerRightClick{}

type PlayerMessage struct {
	Event

	player      soulsand.SyncPlayer
	message     string
	chatMessage *chat.Message
}

func NewMessage(player soulsand.SyncPlayer, message string) (string, *PlayerMessage) {
	return "EventPlayerMessage", &PlayerMessage{
		player:  player,
		message: message,
	}
}

func (m *PlayerMessage) SetMessage(msg *chat.Message) {
	m.chatMessage = msg
}

func (m *PlayerMessage) Message() *chat.Message {
	return m.chatMessage
}

func (m *PlayerMessage) SetMessageString(msg string) {
	m.message = msg
}

func (m *PlayerMessage) MessageString() string {
	return m.message
}

func (m *PlayerMessage) Player() soulsand.SyncPlayer {
	return m.player
}

type PlayerJoin struct {
	Event

	player soulsand.SyncPlayer
	Reason string
}

func NewJoin(player soulsand.SyncPlayer, reason string) (string, *PlayerJoin) {
	return "EventPlayerJoin", &PlayerJoin{
		player: player,
		Reason: reason,
	}
}

func (j *PlayerJoin) Player() soulsand.SyncPlayer {
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
		player: player,
	}
}

func (l *PlayerLeave) Player() soulsand.SyncPlayer {
	return l.player
}

type PlayerLeftClick struct {
	Event

	player soulsand.SyncPlayer
}

func NewPlayerLeftClick(player soulsand.SyncPlayer) (string, *PlayerLeftClick) {
	return "EventPlayerLeftClick", &PlayerLeftClick{
		player: player,
	}
}

func (l *PlayerLeftClick) Player() soulsand.SyncPlayer {
	return l.player
}

type PlayerBlockPlace struct {
	Event

	player  soulsand.SyncPlayer
	x, y, z int
	block   soulsand.ItemStack
}

func NewPlayerBlockPlace(player soulsand.SyncPlayer, x, y, z int, block soulsand.ItemStack) (string, *PlayerBlockPlace) {
	return "EventPlayerBlockPlace", &PlayerBlockPlace{
		player: player,
		x:      x,
		y:      y,
		z:      z,
		block:  block,
	}
}

func (e *PlayerBlockPlace) Player() soulsand.SyncPlayer {
	return e.player
}

func (e *PlayerBlockPlace) Position() (int, int, int) {
	return e.x, e.y, e.z
}

func (e *PlayerBlockPlace) Block() soulsand.ItemStack {
	return e.block
}

func (e *PlayerBlockPlace) SetBlock(block soulsand.ItemStack) {
	e.block = block
}

type PlayerRightClick struct {
	Event

	player soulsand.SyncPlayer
}

func NewPlayerRightClick(player soulsand.SyncPlayer) (string, *PlayerRightClick) {
	return "EventPlayerRightClick", &PlayerRightClick{
		player: player,
	}
}

func (e *PlayerRightClick) Player() soulsand.SyncPlayer {
	return e.player
}

type PlayerBlockBreak struct {
	Event

	player  soulsand.SyncPlayer
	x, y, z int
	face    int8
	status  int8
}

func NewPlayerBlockBreak(player soulsand.SyncPlayer, x, y, z int, face, status int8) (string, *PlayerBlockBreak) {
	return "EventPlayerBlockBreak", &PlayerBlockBreak{
		player: player,
		x:      x,
		y:      y,
		z:      z,
		face:   face,
		status: status,
	}
}

func (e *PlayerBlockBreak) Player() soulsand.SyncPlayer {
	return e.player
}

func (e *PlayerBlockBreak) Position() (int, int, int) {
	return e.x, e.y, e.z
}

func (e *PlayerBlockBreak) Face() int8 {
	return e.face
}

func (e *PlayerBlockBreak) Status() int8 {
	return e.status
}
