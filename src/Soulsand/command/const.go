package command

import (
	"reflect"
	"strings"
)

type caConst struct {
	Value string
}

func (ca *caConst) TabComplete(in string) ([]string, bool) {
	if strings.HasPrefix(ca.Value, in) {
		return []string{ca.Value}, true
	}
	return []string{}, false
}

func (ca *caConst) Parse(in, loc string) (interface{}, error) {
	return nil, nil
}

func (ca *caConst) IsConst() bool {
	return true
}

func (ca *caConst) Printable(loc string) string {
	return ca.Value
}

func (ca *caConst) Type() reflect.Type {
	return nil
}
