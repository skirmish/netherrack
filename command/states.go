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

package command

import (
	"unicode"
)

func lexCommandName(l *lexer) stateFunc {
	for {
		r := l.next()
		if r == eof {
			l.emit(tokenString)
			l.emit(tokenEOF)
			return nil
		}
		if !unicode.IsLetter(r) {
			l.unread()
			l.emit(tokenString)
			return lexFindArgument
		}
	}
}

func lexFindArgument(l *lexer) stateFunc {
	for {
		r := l.next()
		if r == eof {
			l.emit(tokenEOF)
			return nil
		}
		if unicode.IsSpace(r) {
			l.start = l.pos
			continue
		}
		if r == '`' {
			l.start = l.pos
			return lexQuotedString
		}
		if r == '$' {
			l.start = l.pos
			return lexNestedCommand
		}
		if unicode.IsDigit(r) {
			l.unread()
			return lexNumber
		}
		l.unread()
		return lexString
		l.errorf("Unknown rune %c", r)
		return nil
	}
}

func lexString(l *lexer) stateFunc {
	for {
		r := l.next()
		if r == eof {
			l.emit(tokenString)
			l.emit(tokenEOF)
			return nil
		}
		if unicode.IsSpace(r) {
			l.unread()
			l.emit(tokenString)
			return lexFindArgument(l)
		}
	}
}

func lexNumber(l *lexer) stateFunc {
	for {
		r := l.next()
		if r == eof {
			l.emit(tokenNumber)
			l.emit(tokenEOF)
			return nil
		}
		if !unicode.IsDigit(r) && r != '.' {
			l.errorf("Failed to parse %c as a number", r)
			return nil
		}
		if unicode.IsSpace(r) {
			l.unread()
			l.emit(tokenNumber)
			return lexFindArgument(l)
		}
	}
}

func lexQuotedString(l *lexer) stateFunc {
	for {
		r := l.next()
		if r == eof {
			l.errorf("Unexpected end of command")
			l.emit(tokenEOF)
			return nil
		}
		if r == '`' {
			l.unread()
			l.emit(tokenString)
			l.next()
			return lexFindArgument(l)
		}
	}
}

func lexNestedCommand(l *lexer) stateFunc {
	r := l.next()
	l.start = l.pos
	if r != '(' {
		l.errorf("Nested commands must start with a (")
		l.emit(tokenEOF)
		return nil
	}
	for {
		r := l.next()
		if r == eof {
			l.errorf("Unexpected end of command")
			l.emit(tokenEOF)
			return nil
		}
		if r == ')' {
			l.unread()
			l.emit(tokenCommand)
			l.next()
			return lexFindArgument(l)
		}
	}
}
