//More infomation can be found at http://netherrackdev.github.io/commandhandling.html
package command

import (
	"bytes"
	"fmt"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/chat"
	"github.com/NetherrackDev/soulsand/locale"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

func init() {
	go consoleWatcher()
}

//Executes the the command for the player. The command should
//be in the format:
//	commandName arg1 arg2 `arg 3` ...
func Exec(com string, caller soulsand.CommandSender) {
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
		caller.SendMessageSync(chat.New().Colour(chat.Red).Text(fmt.Sprintf(locale.Get(caller.LocaleSync(), "command.error.unknown"), comName)))
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
		caller.SendMessageSync(chat.New().Colour(chat.Grey).Text(fmt.Sprintf(locale.Get(caller.LocaleSync(), "command.usage.command"), comName)))
		for _, c := range command {
			msg := chat.New()
			msg.Colour(chat.Grey)
			msg.Text("/" + comName)
			for _, a := range c.Arguments {
				msg.Colour(chat.Gold)
				msg.Text(" " + a.Printable(caller.LocaleSync()))
			}
			caller.SendMessageSync(msg)
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
			if cst, ok := a.(*caConst); ok {
				if cst.Value != args[i] {
					continue comLoop
				}
			} else {
				res, err := a.Parse(args[i], caller.LocaleSync())
				if err != nil {
					lastError = err
					continue comLoop
				}
				outArgs = append(outArgs, reflect.ValueOf(res))
			}
		}
		c.Function.Call(outArgs)
		return
	}

	if lastError != nil {
		caller.SendMessageSync(chat.New().Colour(chat.Red).Text(fmt.Sprintf(locale.Get(caller.LocaleSync(), "command.error.parse"), lastError)))
	} else {
		///Print usage
		caller.SendMessageSync(chat.New().Colour(chat.Grey).Text(fmt.Sprintf(locale.Get(caller.LocaleSync(), "command.usage.command"), comName)))
		for _, c := range command {
			msg := chat.New()
			msg.Colour(chat.Grey)
			msg.Text("/" + comName)
			for _, a := range c.Arguments {
				msg.Colour(chat.Gold)
				msg.Text(" " + a.Printable(caller.LocaleSync()))
			}
			caller.SendMessageSync(msg)
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
				if cst, ok := a.(*caConst); ok {
					if cst.Value != args[i] {
						continue comLoop
					}
				} else {
					_, err := a.Parse(args[i], "en_GB")
					if err != nil {
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

//Causes the parsing of all commands added with Add. Should only be
//called by the implementation
func Parse() {
	for com, f := range toParse {
		add(com, f)
	}
	toParse = nil
}

var toParse = map[string]interface{}{}

//Adds a command to the system. Commands must only be added at init time.
//Commands should be in the format:
//	commandName const [type(optArgs)] ...
func Add(com string, f interface{}) {
	toParse[com] = f
}

func add(com string, f interface{}) {
	tokens := tokenize(com)
	commandConst, ok := tokens[0].(constToken)
	if !ok {
		panic(com + " - Command must start with a constant")
	}
	commandName := commandConst.Value

	def := &commandDef{}
	def.Function = reflect.ValueOf(f)
	if _, ok := commands[commandName]; !ok {
		commands[commandName] = make([]*commandDef, 0, 1)
	}
	funcType := def.Function.Type()
	if funcType.NumIn() < 1 || !reflect.TypeOf((*soulsand.CommandSender)(nil)).Elem().AssignableTo(funcType.In(0)) {
		panic(com + " - First argument of command should be a CommandSender")
	}

	commands[commandName] = append(commands[commandName], def)
	def.Arguments = make([]CommandArgument, 0, 10)
	if len(tokens) == 1 {
		if funcType.NumIn() != 1 {
			panic(com + " - Argument count mis-match (Too many arguments)")
		}
		return
	}
	tokens = tokens[1:]
	argumentPosition := 1

	for _, token := range tokens {
		if c, ok := token.(constToken); ok {
			def.Arguments = append(def.Arguments, &caConst{c.Value})
		} else {
			a := token.(argToken)
			cAT, ok := commandArgsTypes[a.Type]
			if !ok {
				panic(com + " - Invalid command argument type")
			}
			if funcType.NumIn() < argumentPosition+1 {
				panic(com + " - Argument count mis-match (Not enough arguments)")
			}
			argType := cAT(a.Args)
			if !argType.Type().AssignableTo(funcType.In(argumentPosition)) {
				panic(fmt.Sprintf("%s - '%s' cannot be used as '%s'", com, funcType.In(argumentPosition).Name(), argType.Type().Name()))
			}
			argumentPosition++
			def.Arguments = append(def.Arguments, argType)
		}
	}
	if funcType.NumIn() > argumentPosition {
		panic(com + " - Argument count mis-match (Too many arguments)")
	}
}

//Adds the type to system. The type can be referanced by the name. When the
//type is used the function will be called with all arguments as the parameter.
func RegisterType(name string, f func([]string) CommandArgument) {
	commandArgsTypes[name] = f
}

var (
	commands         = make(map[string][]*commandDef)
	commandArgsTypes = map[string]func([]string) CommandArgument{
		"string": func(a []string) CommandArgument {
			var maxLen int
			if len(a) == 0 {
				maxLen = 0
			} else {
				maxLen, _ = strconv.Atoi(a[0])
			}
			return &caString{maxLen}
		},
		"int": func(a []string) CommandArgument {
			out := &caInt{}
			if len(a) == 0 {
				out.HasLimits = false
			} else if len(a) == 2 {
				out.HasLimits = true
				min, err := strconv.Atoi(a[0])
				if err != nil {
					panic(err)
				}
				out.Min = min
				max, err := strconv.Atoi(a[1])
				if err != nil {
					panic(err)
				}
				out.Max = max
			} else {
				panic("caInt limit error")
			}
			return out
		},
		"float": func(a []string) CommandArgument {
			out := &caFloat{}
			if len(a) == 0 {
				out.HasLimits = false
			} else if len(a) == 2 {
				out.HasLimits = true
				min, err := strconv.ParseFloat(a[0], 64)
				if err != nil {
					panic(err)
				}
				out.Min = min
				max, err := strconv.ParseFloat(a[1], 64)
				if err != nil {
					panic(err)
				}
				out.Max = max
			} else {
				panic("caFloat limit error")
			}
			return out
		},
		"player": func(a []string) CommandArgument {
			out := &caPlayer{}
			return out
		},
	}
)

type commandDef struct {
	Function  reflect.Value
	Arguments []CommandArgument
}

type CommandArgument interface {
	//Parses the input in the passed locale
	Parse(input, locale string) (interface{}, error)
	//Returns a slice of possible completions for the passed input, should return true if the input was valid
	TabComplete(input string) ([]string, bool)
	//Returns a printable version of the argument that will be displayed in the usage message
	Printable(locale string) string
	//Returns the type that Parse will return
	Type() reflect.Type
}

type constToken struct {
	Value string
}

type argToken struct {
	Type string
	Args []string
}

func tokenize(str string) []interface{} {
	out := make([]interface{}, 0, 5)
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '[' { //constToken
			if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
				panic("Invalid rune " + string(r))
			}
			c := make([]rune, 0, 1)
			c = append(c, r)
			i++
			r = runes[i]
			for unicode.IsLetter(r) || unicode.IsDigit(r) {
				c = append(c, r)
				i++
				if i >= len(str) {
					r = 0
					break
				}
				r = runes[i]
			}
			if r != ' ' && r != 0 {
				panic("Space or end of string expected, found " + string(r))
			}
			out = append(out, constToken{string(c)})
		} else {
			i++
			r = runes[i]
			if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
				panic("Invalid rune " + string(r))
			}
			c := make([]rune, 0, 1)
			c = append(c, r)
			i++
			r = runes[i]
			for unicode.IsLetter(r) || unicode.IsDigit(r) {
				c = append(c, r)
				i++
				if i >= len(str) {
					r = 0
					break
				}
				r = runes[i]
			}
			if r != ']' && r != '(' {
				panic("] or ( expected, found " + string(r))
			}
			arg := argToken{}
			arg.Type = string(c)
			arg.Args = make([]string, 0)
			if r == '(' {
				i++
				r = runes[i]
				for {
					c := make([]rune, 0, 1)
					for r != ',' && r != ')' {
						if r == ' ' && len(c) == 0 {
							i++
							r = runes[i]
							continue
						}
						if r == '\\' {
							i++
							r = runes[i]
						}
						c = append(c, r)
						i++
						if i >= len(str) {
							r = 0
							break
						}
						r = runes[i]
					}
					if r == ')' {
						arg.Args = append(arg.Args, string(c))
						break
					}
					if r == ',' {
						arg.Args = append(arg.Args, string(c))
						i++
						r = runes[i]
					} else {
						panic("Error")
					}
				}
				i++
			}
			out = append(out, arg)
			i++
		}
	}
	return out
}
