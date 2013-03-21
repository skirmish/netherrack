package command

import (
	"strings"
)

type ca_Const struct {
	Value string
}

func (ca *ca_Const) TabComplete(in string) ([]string, bool) {
	if strings.HasPrefix(ca.Value, in) {
		return []string{ca.Value}, true
	}
	return []string{}, false
}

func (ca *ca_Const) Parse(in, loc string) (interface{}, error) {
	return nil, nil
}

func (ca *ca_Const) IsConst() bool {
	return true
}

func (ca *ca_Const) Printable(loc string) string {
	return ca.Value
}
