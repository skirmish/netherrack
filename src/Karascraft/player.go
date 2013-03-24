package main

import (
	"Soulsand"
	"Soulsand/event"
	"fmt"
	"log"
	"runtime"
)

func playerWatcher(player Soulsand.SyncPlayer) {
	chatEvent := make(chan Soulsand.Event, 10)
	player.Register(event.PLAYER_MESSAGE, chatEvent)
	leaveEvent := make(chan Soulsand.Event, 1)
	player.Register(event.PLAYER_LEAVE, leaveEvent)
	for {
		select {
		case e := <-chatEvent:
			ev := e.(Soulsand.EventPlayerMessage)
			msg := ev.GetMessage()
			ev.SetMessage(fmt.Sprintf("["+Soulsand.ColourCyan+"%s"+Soulsand.ChatReset+"]: %s", player.GetName(), msg))
			log.Printf("[%s]: %s", player.GetName(), msg)
			ev.Done()
		case e := <-leaveEvent:
			ev := e.(Soulsand.EventPlayerLeave)
			log.Println("Player leave event")
			ev.Done()
			runtime.Goexit()
		}
	}
}
