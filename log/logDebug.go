// +build debug

package log

import (
	"github.com/NetherrackDev/soulsand/chat"
)

func Debug(str string) {
	mCPrintln(chat.New().Colour(chat.Yellow).Text(str), 3)
}
