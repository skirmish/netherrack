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
