package main

import (
	_ "Netherrack"
	"Soulsand"
	"Soulsand/event"
	"flag"
	"log"
	"runtime"
	"strings"
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
	server.Register(event.PLAYER_JOIN, test)
	for {
		e := (<-test).(Soulsand.EventPlayerJoin)
		log.Println("A player joined: ", e.GetPlayer().GetName())
		if strings.Contains(e.GetPlayer().GetName(), "think") {
			e.Disconnect("Nope: " + e.GetPlayer().GetName())
		}
		e.Done()
	}
}
