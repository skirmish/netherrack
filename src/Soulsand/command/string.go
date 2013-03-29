package command

import (
	"Soulsand/locale"
	"errors"
	"fmt"
	"reflect"
)

type caString struct {
	MaxLength int
}

func (ca *caString) Parse(in, loc string) (interface{}, error) {
	if ca.MaxLength != 0 && len(in) > ca.MaxLength {
		return nil, errors.New(fmt.Sprintf(locale.Get(loc, "command.error.string.length"), ca.MaxLength))
	}
	return in, nil
}

func (ca *caString) TabComplete(in string) ([]string, bool) {
	return []string{}, false
}

func (ca *caString) IsConst() bool {
	return false
}

func (ca *caString) Printable(loc string) string {
	if ca.MaxLength == 0 {
		return locale.Get(loc, "command.usage.string.norange")
	}
	return fmt.Sprintf(locale.Get(loc, "command.usage.string.range"), ca.MaxLength)
}

func (ca *caString) Type() reflect.Type {
	return reflect.TypeOf("")
}
