package event

import (
	"Soulsand"
	"Soulsand/locale"
	"fmt"
)

func init() {
	go watcher()
}

var channel chan func() = make(chan func(), 1000)

func watcher() {
	for {
		f := <-channel
		f()
	}
}

var (
	entitysById   map[int32]Soulsand.Entity  = make(map[int32]Soulsand.Entity)
	playersByName map[string]Soulsand.Player = make(map[string]Soulsand.Player)
	currentEID    int32                      = 0
)

func GetEntityCount() int {
	res := make(chan int, 1)
	channel <- func() {
		res <- len(entitysById)
	}
	return <-res
}

func Broadcast(message string) {
	channel <- func() {
		for _, p := range playersByName {
			p.SendMessage(message)
		}
	}
}

func AddPlayer(p Soulsand.Player) {
	channel <- func() {
		playersByName[p.GetName()] = p
		for _, player := range playersByName {
			message := fmt.Sprintf(locale.Get(player.GetLocale(), "message.player.connect"), p.GetDisplayName())
			player.SendMessage(message)
		}
	}
}

func RemovePlayer(p Soulsand.Player) {
	channel <- func() {
		delete(playersByName, p.GetName())
		for _, player := range playersByName {
			message := fmt.Sprintf(locale.Get(player.GetLocale(), "message.player.disconnect"), p.GetDisplayName())
			player.SendMessage(message)
		}
	}
}

func GetPlayer(name string) Soulsand.Player {
	res := make(chan Soulsand.Player)
	channel <- func() {
		res <- playersByName[name]
	}
	return <-res
}

func GetFreeEntityID(e Soulsand.Entity) int32 {
	res := make(chan int32, 1)
	channel <- func() {
		for {
			_, ok := entitysById[currentEID]
			if !ok {
				entitysById[currentEID] = e
				res <- currentEID
				currentEID++
				return
			}
			currentEID++
		}
	}
	return <-res
}

func FreeEntityID(e Soulsand.Entity) {
	channel <- func() {
		delete(entitysById, e.GetID())
	}
}
