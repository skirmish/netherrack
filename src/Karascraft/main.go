package main

import (
	_ "Netherrack"
	"Soulsand"
	"Soulsand/event"
	"flag"
	"log"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	port := flag.Int("port", 25565, "Server port")
	offline := flag.Bool("offline", false, "Offline mode")
	randNames := flag.Bool("rnames", false, "Random names")
	ip := flag.String("ip", "", "Ip to bind to")
	flag.Parse()

	server := Soulsand.GetServer()

	server.SetFlag(Soulsand.OFFLINE_MODE, *offline)
	server.SetFlag(Soulsand.RANDOM_NAMES, *randNames)

	server.Start(*ip, *port)

	server.SetDefaultGamemode(Soulsand.GAMEMODE_ADVENTURE)

	server.SetMessageOfTheDay(Soulsand.ColourRed + "Netherrack " + Soulsand.ChatReset + "Server")
	server.SetMaxPlayers(100)

	test := make(chan Soulsand.Event, 1000)
	eID := server.Register(event.PLAYER_MESSAGE, test)
	for {
		e := (<-test).(Soulsand.EventPlayerMessage)
		log.Println("Got message: ", e.GetMessage())
		if e.GetMessage() == "stop" {
			e.Cancel()
			e.Remove(eID)
		}
		e.Done()
	}
}
