/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

//Package command provides methods to execute commands from strings
package command

import (
	"fmt"
	"github.com/NetherrackDev/netherrack/message"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

var commands = map[string]*commandDef{}

type commandDef struct {
	callback reflect.Value
	sub      map[string]*commandDef
}

var (
	stringType = reflect.TypeOf((*string)(nil)).Elem()
	floatType  = reflect.TypeOf((*float64)(nil)).Elem()
	callerType = reflect.TypeOf((*Caller)(nil)).Elem()
)

//The format is a follows:
//    command #argType1 const2
//Must not be called after init
func Register(def string, callback interface{}) {
	l := lex(def)

	cb := reflect.ValueOf(callback)
	cbType := cb.Type()

	if cbType.NumOut() != 2 || (!stringType.AssignableTo(cbType.Out(0)) && !floatType.AssignableTo(cbType.Out(0))) || !stringType.AssignableTo(cbType.Out(1)) {
		panic("Callback must return a string/float64 and a string")
	}

	if cbType.NumOut() == 0 || !callerType.AssignableTo(cbType.In(0)) {
		panic("Callback must take a command.Caller as its first argument")
	}

	comToken := l.NextToken()
	if comToken.tokenType != tokenString {
		panic("Command name must be a string")
	}
	commandName := comToken.value

	currentPart := commands[commandName]
	if currentPart == nil {
		currentPart = &commandDef{
			sub: map[string]*commandDef{},
		}
		commands[commandName] = currentPart
	}

	i := 1
parse:
	for {
		t := l.NextToken()
		switch t.tokenType {
		case tokenEOF:
			break parse
		case tokenError:
			panic(t.value)
		case tokenCommand:
			panic("Nested commands cannot be used in RegisterCommand")
		case tokenNumber:
			panic("Numbers cannot be used in RegisterCommand")
		case tokenString:
			if strings.HasPrefix(t.value, "#") { //Type
				if cbType.NumIn() <= i {
					panic("Argument count mismatch")
				}
				ty := t.value[1:]
				switch ty {
				case "string":
					if !stringType.AssignableTo(cbType.In(i)) {
						panic(fmt.Sprintf("%s != %s for argument %d", ty, cbType.In(i), i))
					}
				case "number":
					if !floatType.AssignableTo(cbType.In(i)) {
						panic(fmt.Sprintf("%s != %s for argument %d", ty, cbType.In(i), i))
					}
				}
				temp := currentPart
				currentPart = temp.sub[t.value]
				if currentPart == nil {
					currentPart = &commandDef{
						sub: map[string]*commandDef{},
					}
					temp.sub[t.value] = currentPart
				}
				i++
				continue parse
			}
			//Const
			temp := currentPart
			currentPart = temp.sub[t.value]
			if currentPart == nil {
				currentPart = &commandDef{
					sub: map[string]*commandDef{},
				}
				temp.sub[t.value] = currentPart
			}
		}
	}
	if i != cbType.NumIn() {
		panic("Incorrect number of arguments")
	}

	currentPart.callback = cb
}

//
type Caller interface {
	CanCall(string) bool
	SendMessage(*message.Message)
}

//The format of the command is as follows:
//    command arg1 2 `arg 3` $(echo arg4)
//
//'command' is the first part identifies the command tree to check
//this must be a constant value.
//
//'arg1' is a string type
//
//'2' is a number type (maybe float or int depending on the command)
//
//'$(echo arg4)' This will execute the command contained in the brackets
//up to one deep
func Exec(caller Caller, command string) (ret interface{}, err string) {
	l := lex(command)

	comToken := l.NextToken()
	if comToken.tokenType != tokenString {
		err = "Unknown command"
		return
	}
	commandName := comToken.value
	if !caller.CanCall(commandName) {
		err = "You cannot call this command"
		return
	}

	arguments := []reflect.Value{reflect.ValueOf(caller)}

	currentPart := commands[commandName]
	if currentPart == nil {
		err = "Unknown command"
		return
	}

argLoop:
	for {
		t := l.NextToken()
		switch t.tokenType {
		case tokenEOF:
			break argLoop
		case tokenError:
			err = t.value
			return
		case tokenString:
			temp := currentPart
			currentPart = temp.sub[t.value]
			if currentPart == nil {
				currentPart = temp.sub["#string"]
				if currentPart == nil {
					err = "Unknown command" //TODO: Backtrack and look for a better option
					return
				}
				arguments = append(arguments, reflect.ValueOf(t.value))
			}
		case tokenNumber:
			num, e := strconv.ParseFloat(t.value, 64)
			if e != nil {
				err = e.Error()
				return
			}
			temp := currentPart
			currentPart = temp.sub["#number"]
			if currentPart == nil {
				err = "Unknown command"
				return
			}
			arguments = append(arguments, reflect.ValueOf(num))
		case tokenCommand:
			val, e := Exec(caller, t.value)
			if e != "" {
				err = e
				return
			}
			temp := currentPart
			switch val.(type) {
			case string:
				currentPart = temp.sub["#string"]
			case float64:
				currentPart = temp.sub["#number"]
			}
			if currentPart == nil {
				err = "Unknown command"
				return
			}
			arguments = append(arguments, reflect.ValueOf(val))
		}
	}

	if !currentPart.callback.IsValid() {
		err = "Unknown command"
		return
	}

	reflectReturn := currentPart.callback.Call(arguments)
	ret = reflectReturn[0].Interface()
	err = reflectReturn[1].String()

	return
}

//
func Complete(caller Caller, command string) []string {
	return []string{}
}

type stateFunc func(*lexer) stateFunc

type lexer struct {
	str    string
	start  int
	pos    int
	width  int
	state  stateFunc
	tokens chan token
}

func lex(str string) *lexer {
	return &lexer{
		str:    str,
		tokens: make(chan token, 2),
		state:  lexCommandName,
	}
}

func (l *lexer) NextToken() (t token) {
	for {
		select {
		case t = <-l.tokens:
			return
		default:
			if l.state == nil {
				t = token{tokenType: tokenEOF}
				return
			}
			l.state = l.state(l)
		}
	}
}

var eof rune = -1

func (l *lexer) next() rune {
	if l.pos >= len(l.str) {
		return eof
	}
	r, width := utf8.DecodeRuneInString(l.str[l.pos:])
	l.width = width
	l.pos += width
	return r
}

func (l *lexer) unread() {
	l.pos -= l.width
}

func (l *lexer) accept(runes string) bool {
	r := l.next()
	if strings.ContainsRune(runes, r) {
		return true
	}
	l.unread()
	return false
}

func (l *lexer) emit(t tokenType) {
	l.tokens <- token{
		tokenType: t,
		value:     l.str[l.start:l.pos],
	}
	l.start = l.pos
}

func (l *lexer) errorf(format string, vals ...interface{}) {
	l.tokens <- token{
		tokenType: tokenError,
		value:     fmt.Sprintf(format, vals...),
	}
}

type tokenType int

const (
	tokenError tokenType = iota
	tokenEOF
	tokenString
	tokenNumber
	tokenCommand
)

type token struct {
	tokenType tokenType
	value     string
}

func (t token) String() string {
	switch t.tokenType {
	case tokenError:
		return fmt.Sprintf("Error: %s", t.value)
	case tokenEOF:
		return "EOF"
	default:
		return t.value
	}
}
