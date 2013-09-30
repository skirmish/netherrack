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
	"strings"
	"testing"
)

type testCaller struct{}

func init() {
	RegisterCommand("echo_num #number", test_echo)
	RegisterCommand("echo upper #string", test_subecho)
	RegisterCommand("echo upper #string smile", test_subecho_smile)
}

func test_echo(caller Caller, msg float64) (float64, string) {
	return msg, ""
}

func test_subecho(caller Caller, msg string) (string, string) {
	return strings.ToUpper(msg), ""
}

func test_subecho_smile(caller Caller, msg string) (string, string) {
	return strings.ToUpper(msg) + " :)", ""
}

func (testCaller) CanCall(string) bool { return true }

func TestBasic(t *testing.T) {
	val, err := Exec(testCaller{}, "echo hello")
	if err != "" {
		t.Fatal(err)
	}
	valS := val.(string)
	if valS != "hello" {
		t.Fail()
	}
}

func TestNumber(t *testing.T) {
	val, err := Exec(testCaller{}, "echo_num 5456")
	if err != "" {
		t.Fatal(err)
	}
	valF := val.(float64)
	if valF != 5456 {
		t.Fail()
	}
}

func TestQuoted(t *testing.T) {
	val, err := Exec(testCaller{}, "echo `hello world`")
	if err != "" {
		t.Fatal(err)
	}
	valS := val.(string)
	if valS != "hello world" {
		t.Fail()
	}
}

func TestNested(t *testing.T) {
	val, err := Exec(testCaller{}, "echo $(echo hello)")
	if err != "" {
		t.Fatal(err)
	}
	valS := val.(string)
	if valS != "hello" {
		t.Fail()
	}
}

func TestSub1(t *testing.T) {
	val, err := Exec(testCaller{}, "echo upper hello")
	if err != "" {
		t.Fatal(err)
	}
	valS := val.(string)
	if valS != "HELLO" {
		t.Fail()
	}
}

func TestSub2(t *testing.T) {
	val, err := Exec(testCaller{}, "echo upper hello smile")
	if err != "" {
		t.Fatal(err)
	}
	valS := val.(string)
	if valS != "HELLO :)" {
		t.Fail()
	}
}
