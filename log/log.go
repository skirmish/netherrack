// +build !linux,!windows

package log

import (
	"bytes"
	"fmt"
	"github.com/NetherrackDev/soulsand"
	"os"
	"sync"
)

const ()

var (
	globalLock sync.Mutex
)

func Fatalln(args ...interface{}) {
	Println(args...)
	os.Exit(-1)
}

func MCPrintln(msg string) {
	printFileInfo()
	var buf bytes.Buffer
	msgRunes := []rune(msg)
	for i := 0; i < len(msgRunes); i++ {
		c := msgRunes[i]
		if c == soulsand.ChatModifierRune {
			i++
			continue
		}
		buf.WriteRune(c)
	}
	os.Stdout.WriteString(buf.String())
}

func Println(args ...interface{}) {
	globalLock.Lock()
	defer globalLock.Unlock()
	printFileInfo()
	fmt.Fprintln(os.Stdout, args...)
}

func Printf(fmtStr string, args ...interface{}) {
	globalLock.Lock()
	defer globalLock.Unlock()
	printFileInfo()
	fmt.Fprintf(os.Stdout, fmtStr, args...)
	if fmtStr[len(fmtStr)-1] != '\n' {
		os.Stdout.WriteString("\n")
	}
}

func printFileInfo() {
	_, file, line, _ := runtime.Caller(2)

	fmt.Fprint(os.Stdout, file[strings.LastIndex(file, "/")+1:], ":")
	fmt.Fprint(os.Stdout, line, " ")
}
