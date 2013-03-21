package command

import (
	"Soulsand/locale"
	"errors"
	"fmt"
	"strconv"
)

type ca_Float struct {
	HasLimits bool
	Min       float64
	Max       float64
}

func (ca *ca_Float) Parse(in, loc string) (interface{}, error) {
	if ca.HasLimits {
		i, err := strconv.ParseFloat(in, 64)
		if err != nil {
			return nil, err
		}
		if i < ca.Min || i > ca.Max {
			return nil, errors.New(fmt.Sprintf(locale.Get(loc, "command.error.float.range"), ca.Min, ca.Max))
		}
		return i, nil
	} else {
		i, err := strconv.ParseFloat(in, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	}
	return nil, nil
}

func (ca *ca_Float) TabComplete(in string) ([]string, bool) {
	return []string{}, false
}

func (ca *ca_Float) IsConst() bool {
	return false
}

func (ca *ca_Float) Printable(loc string) string {
	if ca.HasLimits {
		return fmt.Sprintf(locale.Get(loc, "command.usage.float.range"), ca.Min, ca.Max)
	}
	return locale.Get(loc, "command.usage.float.norange")
}
