package command

import (
	"errors"
	"fmt"
	"github.com/NetherrackDev/soulsand/locale"
	"reflect"
	"strconv"
)

type caFloat struct {
	HasLimits bool
	Min       float64
	Max       float64
}

func (ca *caFloat) Parse(in, loc string) (interface{}, error) {
	if ca.HasLimits {
		i, err := strconv.ParseFloat(in, 64)
		if err != nil {
			return nil, err
		}
		if i < ca.Min || i > ca.Max {
			return nil, errors.New(fmt.Sprintf(locale.Get(loc, "command.error.float.range"), ca.Min, ca.Max))
		}
		return i, nil
	}
	i, err := strconv.ParseFloat(in, 64)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (ca *caFloat) TabComplete(in string) ([]string, bool) {
	return []string{}, false
}

func (ca *caFloat) Printable(loc string) string {
	if ca.HasLimits {
		return fmt.Sprintf(locale.Get(loc, "command.usage.float.range"), ca.Min, ca.Max)
	}
	return locale.Get(loc, "command.usage.float.norange")
}

func (ca *caFloat) Type() reflect.Type {
	return reflect.TypeOf(float64(0))
}
