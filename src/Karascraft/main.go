package main

import (
	_ "Netherrack"
	"Soulsand"
	"flag"
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

	var test chan bool
	<-test
}
