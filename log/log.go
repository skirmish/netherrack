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

// +build !linux,!windows

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const ()

var (
	globalLock sync.Mutex
)

func stripColourCodes(str string) string {
	var buf bytes.Buffer
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if c == '§' {
			i++
			continue
		}
		buf.WriteRune(c)
	}
	return buf.String()
}

func Println(args ...interface{}) {
	globalLock.Lock()
	defer globalLock.Unlock()
	printFileInfo(2)
	fmt.Fprintln(os.Stdout, args...)
}

func Printf(fmtStr string, args ...interface{}) {
	globalLock.Lock()
	defer globalLock.Unlock()
	printFileInfo(2)
	fmt.Fprintf(os.Stdout, fmtStr, args...)
	if fmtStr[len(fmtStr)-1] != '\n' {
		os.Stdout.WriteString("\n")
	}
}

func printFileInfo(level int) {
	_, file, line, _ := runtime.Caller(level)

	fmt.Fprint(os.Stdout, file[strings.LastIndex(file, "/")+1:], ":")
	fmt.Fprint(os.Stdout, line, " ")
}
