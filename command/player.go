package command

import (
	"errors"
	"fmt"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/locale"
	"github.com/NetherrackDev/soulsand/server"
	"reflect"
	"strings"
)

type caPlayer struct {
}

func (ca *caPlayer) Parse(in, loc string) (interface{}, error) {
	player := server.Player(in)
	if player == nil {
		return nil, errors.New(fmt.Sprintf(locale.Get(loc, "command.error.player"), in))
	}
	return player, nil
}

func (ca *caPlayer) TabComplete(in string) ([]string, bool) {
	players := server.Players()
	out := make([]string, 0, 1)
	in = strings.ToLower(in)
	for _, p := range players {
		if strings.HasPrefix(strings.ToLower(p.Name()), in) {
			out = append(out, p.Name())
		}
	}
	return out, len(out) != 0
}

func (ca *caPlayer) Printable(loc string) string {
	return locale.Get(loc, "command.usage.player")
}

func (ca *caPlayer) Type() reflect.Type {
	return reflect.TypeOf((*soulsand.Player)(nil)).Elem()
}
