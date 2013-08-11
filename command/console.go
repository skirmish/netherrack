package command

import (
	"bufio"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/chat"
	"github.com/NetherrackDev/soulsand/log"
	"os"
)

var _ soulsand.CommandSender = consoleSender{}

type consoleSender struct {
}

func (consoleSender) LocaleSync() string {
	return "en_GB"
}

func (consoleSender) SendMessageSync(msg *chat.Message) {
	log.MCPrintln(msg)
}

func consoleWatcher() {
	buf := bufio.NewReader(os.Stdin)
	cs := consoleSender{}
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return
		}
		Exec(line, cs)
	}
}
