package command

import (
	"Soulsand/locale"
	"errors"
	"fmt"
	"strconv"
)

type ca_Int struct {
	HasLimits bool
	Min       int
	Max       int
}

func (ca *ca_Int) Parse(in, loc string) (interface{}, error) {
	if ca.HasLimits {
		i, err := strconv.Atoi(in)
		if err != nil {
			return nil, err
		}
		if i < ca.Min || i > ca.Max {
			return nil, errors.New(fmt.Sprintf(locale.Get(loc, "command.error.int.range"), ca.Min, ca.Max))
		}
		return i, nil
	} else {
		i, err := strconv.Atoi(in)
		if err != nil {
			return nil, err
		}
		return i, nil
	}
	return nil, nil
}

func (ca *ca_Int) TabComplete(in string) ([]string, bool) {
	return []string{}, false
}

func (ca *ca_Int) IsConst() bool {
	return false
}

func (ca *ca_Int) Printable(loc string) string {
	if ca.HasLimits {
		return fmt.Sprintf(locale.Get(loc, "command.usage.int.range"), ca.Min, ca.Max)
	}
	return locale.Get(loc, "command.usage.int.norange")
}
