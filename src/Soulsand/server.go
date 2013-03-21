package Soulsand

import ()

//These flags will control how the server will run. Set with SetFlag
const (
	//Stops the server from connecting to the auth servers
	OFFLINE_MODE uint64 = 1 << iota
	//Changes every player's name to contain random numbers. Useful for testing with a single account.
	RANDOM_NAMES
)

type Gamemode int

//Gamemodes
const (
	GAMEMODE_SURVIVAL Gamemode = iota
	GAMEMODE_CREATIVE
	GAMEMODE_ADVENTURE
)

type Server interface {
	//Start the server on the specified ip and port
	Start(ip string, port int)
	//Changes a flag on the server. See constants
	SetFlag(flag uint64, value bool)
	//Returns the value of the flag for the server
	GetFlag(flag uint64) bool
	//Returns data required to show a list ping
	GetListPingData() []string
	//Changes the server's message of the day
	SetMessageOfTheDay(message string)
	//Changes the server's max player count. This may only effect the list ping
	SetMaxPlayers(max int)
	//Gets a player by name. Returns nil if the player doesn't exist
	GetPlayer(name string) Player
	//Gets a world by name
	GetWorld(name string) World
	//Gets the number of entitys on the server
	GetEntityCount() int
	//Sets the default gamemode for the server
	SetDefaultGamemode(mode Gamemode)
	//Gets the default gamemode for the server
	GetDefaultGamemode() Gamemode
}

var server Server

//Returns the current server implementation
func GetServer() Server {
	return server
}

//Sets the current server implementation. Should only be called by the implementation
func SetServer(s Server) {
	server = s
}
