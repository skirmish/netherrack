package command

import (
	"Soulsand/locale"
	"errors"
	"fmt"
)

type ca_String struct {
	MaxLength int
}

func (ca *ca_String) Parse(in, loc string) (interface{}, error) {
	if ca.MaxLength != 0 && len(in) > ca.MaxLength {
		return nil, errors.New(fmt.Sprintf(locale.Get(loc, "command.error.string.length"), ca.MaxLength))
	}
	return in, nil
}

func (ca *ca_String) TabComplete(in string) ([]string, bool) {
	return []string{}, false
}

func (ca *ca_String) IsConst() bool {
	return false
}

func (ca *ca_String) Printable(loc string) string {
	if ca.MaxLength == 0 {
		return locale.Get(loc, "command.usage.string.norange")
	}
	return fmt.Sprintf(locale.Get(loc, "command.usage.string.range"), ca.MaxLength)
}
