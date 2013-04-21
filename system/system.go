package system

import (
	"github.com/thinkofdeath/netherrack/event"
	"github.com/thinkofdeath/soulsand"
	"github.com/thinkofdeath/soulsand/locale"
	"fmt"
	"log"
)

func init() {
	go watcher()
}

var channel chan func() = make(chan func(), 1000)
var EventSource event.Source

func watcher() {
	for {
		f := <-channel
		f()
	}
}

var (
	entitysById         = make(map[int32]soulsand.Entity)
	playersByName       = make(map[string]soulsand.Player)
	currentEID    int32 = 0
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
		log.Println(message)
		for _, p := range playersByName {
			p.SendMessage(message)
		}
	}
}

func AddPlayer(p soulsand.Player) {
	channel <- func() {
		playersByName[p.GetName()] = p
		displayName, err := p.GetDisplayName()
		if err != nil {
			return
		}
		for _, player := range playersByName {
			playerLocale, err := player.GetLocale()
			if err != nil {
				continue
			}
			message := fmt.Sprintf(locale.Get(playerLocale, "message.player.connect"), displayName)
			player.SendMessage(message)
		}
	}
}

func RemovePlayer(p soulsand.Player) {
	displayName := (p.(soulsand.SyncPlayer)).GetDisplayNameSync()
	channel <- func() {
		delete(playersByName, p.GetName())
		for _, player := range playersByName {
			playerLocale, err := player.GetLocale()
			if err != nil {
				continue
			}
			message := fmt.Sprintf(locale.Get(playerLocale, "message.player.disconnect"), displayName)
			player.SendMessage(message)
		}
	}
}

func GetPlayer(name string) soulsand.Player {
	res := make(chan soulsand.Player)
	channel <- func() {
		res <- playersByName[name]
	}
	return <-res
}

func GetFreeEntityID(e soulsand.Entity) int32 {
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

func FreeEntityID(e soulsand.Entity) {
	channel <- func() {
		delete(entitysById, e.GetID())
	}
}
