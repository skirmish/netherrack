package command

import (
	"errors"
	"fmt"
	"github.com/NetherrackDev/soulsand/locale"
	"reflect"
	"strconv"
)

type caInt struct {
	HasLimits bool
	Min       int
	Max       int
}

func (ca *caInt) Parse(in, loc string) (interface{}, error) {
	if ca.HasLimits {
		i, err := strconv.Atoi(in)
		if err != nil {
			return nil, err
		}
		if i < ca.Min || i > ca.Max {
			return nil, errors.New(fmt.Sprintf(locale.Get(loc, "command.error.int.range"), ca.Min, ca.Max))
		}
		return i, nil
	}
	i, err := strconv.Atoi(in)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (ca *caInt) TabComplete(in string) ([]string, bool) {
	return []string{}, false
}

func (ca *caInt) Printable(loc string) string {
	if ca.HasLimits {
		return fmt.Sprintf(locale.Get(loc, "command.usage.int.range"), ca.Min, ca.Max)
	}
	return locale.Get(loc, "command.usage.int.norange")
}

func (ca *caInt) Type() reflect.Type {
	return reflect.TypeOf(int(0))
}
