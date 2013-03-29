package command

import (
	"Soulsand"
	"Soulsand/locale"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"reflect"
)

//Executes the the command for the player. The command should
//be in the format:
//	commandName arg1 arg2 `arg 3` ...
func Exec(com string, caller Soulsand.CommandSender) {
	com = strings.TrimSpace(com)
	if len(com) == 0 {
		return
	}
	var comName string
	pos := strings.Index(com, " ")
	if pos == -1 {
		comName = com
	} else {
		comName = com[:pos]
	}
	com = com[pos+1:]

	command, ok := commands[comName]

	if !ok {
		caller.SendMessageSync(fmt.Sprintf(locale.Get(caller.GetLocaleSync(), "command.error.unknown"), comName))
		return
	}

	if pos == -1 {
		for _, c := range command {
			if len(c.Arguments) == 0 {
				c.Function.Call([]reflect.Value{reflect.ValueOf(caller)})
				return
			}
		}
		//Print usage
		caller.SendMessageSync(fmt.Sprintf(locale.Get(caller.GetLocaleSync(), "command.usage.command"), comName))
		for _, c := range command {
			var buf bytes.Buffer
			buf.WriteString(Soulsand.ColourGray)
			buf.WriteString(comName)
			for _, a := range c.Arguments {
				buf.WriteString(" ")
				buf.WriteString(Soulsand.ColourGold)
				buf.WriteString(a.Printable(caller.GetLocaleSync()))
			}
			caller.SendMessageSync(buf.String())
		}
		return
	}
	args := make([]string, 0, 10)
	for true {
		var end int
		if len(com) == 0 {
			break
		}
		quote := false
		if com[0] == '`' {
			com = com[1:]
			end = strings.Index(com, "`")
			quote = true
		} else {
			end = strings.Index(com, " ")
		}
		if end == -1 {
			args = append(args, com)
		} else {
			args = append(args, com[:end])
		}
		if quote {
			com = com[end+1:]
			end = strings.Index(com, " ")
		}
		if end != -1 {
			com = com[end+1:]
		} else {
			break
		}
	}
	var lastError error
	
	callerValue := reflect.ValueOf(caller)
comLoop:
	for _, c := range command {
		if len(c.Arguments) != len(args) {
			continue
		}
		outArgs := make([]reflect.Value, 0, 5)
		outArgs = append(outArgs, callerValue)
		for i, a := range c.Arguments {
			if !a.IsConst() {
				res, err := a.Parse(args[i], caller.GetLocaleSync())
				if err != nil {
					lastError = err
					continue comLoop
				}
				outArgs = append(outArgs, reflect.ValueOf(res))
			} else {
				cst := a.(*caConst)
				if cst.Value != args[i] {
					continue comLoop
				}
			}
		}
		c.Function.Call(outArgs)
		return
	}

	if lastError != nil {
		caller.SendMessageSync(fmt.Sprintf(locale.Get(caller.GetLocaleSync(), "command.error.parse"), lastError))
	} else {
		//Print usage
		caller.SendMessageSync(fmt.Sprintf(locale.Get(caller.GetLocaleSync(), "command.usage.command"), comName))
		for _, c := range command {
			var buf bytes.Buffer
			buf.WriteString(Soulsand.ColourGray)
			buf.WriteString(comName)
			for _, a := range c.Arguments {
				buf.WriteString(" ")
				buf.WriteString(Soulsand.ColourGold)
				buf.WriteString(a.Printable(caller.GetLocaleSync()))
			}
			caller.SendMessageSync(buf.String())
		}
	}
}

//Returns a \x00 seperated string containing possible options for completing
//the next argument in the command
func Complete(com string) string {
	com = strings.TrimSpace(com)
	if len(com) == 0 {
		return ""
	}
	var comName string
	pos := strings.Index(com, " ")
	if pos == -1 {
		comName = com
	} else {
		comName = com[:pos]
	}
	com = com[pos+1:]

	command, ok := commands[comName]

	if !ok {
		if pos == -1 {
			out := make([]string, 0, 1)
			for n, _ := range commands {
				if strings.HasPrefix(n, comName) {
					out = append(out, "/"+n)
				}
			}
			return strings.Join(out, "\x00")
		}
		return ""
	}

	args := make([]string, 0, 10)
	for true {
		var end int
		if len(com) == 0 {
			break
		}
		quote := false
		if com[0] == '`' {
			com = com[1:]
			end = strings.Index(com, "`")
			quote = true
		} else {
			end = strings.Index(com, " ")
		}
		if end == -1 {
			args = append(args, com)
		} else {
			args = append(args, com[:end])
		}
		if quote {
			com = com[end+1:]
			end = strings.Index(com, " ")
		}
		if end != -1 {
			com = com[end+1:]
		} else {
			break
		}
	}

	out := make(map[string]bool)

comLoop:
	for _, c := range command {
		for i, a := range c.Arguments {
			if i == len(args)-1 {
				tabs, e := a.TabComplete(args[i])
				if e {
					for _, res := range tabs {
						if _, ok := out[res]; !ok {
							out[res] = true
						}
					}
				}
				continue comLoop
			} else {
				if !a.IsConst() {
					_, err := a.Parse(args[i], "en_GB")
					if err != nil {
						continue comLoop
					}
				} else {
					cst := a.(*caConst)
					if cst.Value != args[i] {
						continue comLoop
					}
				}
			}
		}
	}
	var buf bytes.Buffer
	for n, _ := range out {
		buf.WriteString(n)
		buf.WriteString("\x00")
	}
	if buf.Len() != 0 {
		return buf.String()[:buf.Len()-1]
	}
	return ""
}

//Adds a command to the system. Commands must only be added at init time.
//Commands should be in the format:
//	commandName const $t[optArgs] ...
func Add(com string, f interface{}) {
	com = strings.TrimSpace(com)
	pos := strings.Index(com, " ")
	var comName string
	if pos == -1 {
		comName = com
	} else {
		comName = com[:pos]
	}

	def := &commandDef{}
	def.Function = reflect.ValueOf(f)
	if _, ok := commands[comName]; !ok {
		commands[comName] = make([]*commandDef, 0, 1)
	}
	funcType := def.Function.Type()
	if funcType.NumIn() < 1 || !reflect.TypeOf((*Soulsand.CommandSender)(nil)).Elem().AssignableTo(funcType.In(0)) {
		panic("Commands needs a CommandSender as the first argument. " + funcType.In(0).String())
	}
	
	commands[comName] = append(commands[comName], def)
	def.Arguments = make([]commandArgument, 0, 10)
	if pos == -1 {
		if funcType.NumIn() != 1 {
			panic("Argument count mis-match (Too many arguments)")
		}
		return
	}
	com = com[pos+1:]
	
	argPos := 1
	
	for true {
		pos = strings.Index(com, " ")
		var a string
		if pos == -1 {
			a = com
		} else {
			a = com[:pos]
			com = com[pos+1:]
		}
		if a[0] == '$' { //Variable
			t := a[1]
			cAT, ok := commandArgsTypes[t]
			if !ok {
				panic("Invalid command argument type")
				return
			}
			if funcType.NumIn() < argPos + 1 {
				panic("Argument count mis-match (Too little arguments)")			
			}
			argType := cAT(a[3:len(a)-1])
			if !argType.Type().AssignableTo(funcType.In(argPos)) {
				panic(fmt.Sprintf("'%s' cannot be used as '%s'", funcType.In(argPos).Name(), argType.Type().Name()))
			}
			def.Arguments = append(def.Arguments, argType)
			argPos++
		} else { //Constant
			def.Arguments = append(def.Arguments, &caConst{a})
		}
		if pos == -1 {
			break
		}
	}
	if funcType.NumIn() > argPos {
		panic("Argument count mis-match (Too many arguments)")
	}
}

var (
	commands         map[string][]*commandDef              = make(map[string][]*commandDef)
	commandArgsTypes map[byte]func(string) commandArgument = map[byte]func(string) commandArgument{
		's': func(a string) commandArgument {
			var maxLen int
			if len(a) == 0 {
				maxLen = 0
			} else {
				maxLen, _ = strconv.Atoi(a)
			}
			return &caString{maxLen}
		},
		'i': func(a string) commandArgument {
			out := &caInt{}
			if len(a) == 0 {
				out.HasLimits = false
			} else {
				out.HasLimits = true
				args := strings.Split(a, ",")
				if len(args) != 2 {
					panic("ca_Int Limit error")
				}
				min, err := strconv.Atoi(args[0])
				if err != nil {
					panic(err)
				}
				out.Min = min
				max, err := strconv.Atoi(args[1])
				if err != nil {
					panic(err)
				}
				out.Max = max
			}
			return out
		},
		'f': func(a string) commandArgument {
			out := &caFloat{}
			if len(a) == 0 {
				out.HasLimits = false
			} else {
				out.HasLimits = true
				args := strings.Split(a, ",")
				if len(args) != 2 {
					panic("ca_Float Limit error")
				}
				min, err := strconv.ParseFloat(args[0], 64)
				if err != nil {
					panic(err)
				}
				out.Min = min
				max, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					panic(err)
				}
				out.Max = max
			}
			return out
		},
	}
)

type commandDef struct {
	Function  reflect.Value
	Arguments []commandArgument
}

type commandArgument interface {
	Parse(in, loc string) (interface{}, error)
	TabComplete(in string) ([]string, bool)
	Printable(loc string) string
	IsConst() bool
	Type() reflect.Type
}
