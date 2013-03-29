package Soulsand

import (

)

type CommandSender interface {
	//Sends a message to the command sender
	SendMessageSync(string)
	//Returns the command sender's locale
	GetLocaleSync() string
}