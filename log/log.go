// +build !linux,!windows

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NetherrackDev/soulsand"
	"github.com/NetherrackDev/soulsand/chat"
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

func MCPrintln(chatMsg *chat.Message) {
	mCPrintln(chatMsg, 3)
}

func mCPrintln(chatMsg *chat.Message, level int) {
	globalLock.Lock()
	defer globalLock.Unlock()
	printFileInfo(level)
	fmt.Fprintln(os.Stdout, chatMsg.String())
	return

	globalLock.Lock()
	defer globalLock.Unlock()
	printFileInfo(level)
	msg := map[string]interface{}{}
	json.Unmarshal(chatMsg.Bytes(), msg)
	text, ok := msg["text"].([]interface{})
	if !ok {
		panic("Not implemented")
	}
	for _, i := range text {
		switch i := i.(type) {
		case string:
			os.Stdout.WriteString(i)
		case map[string]interface{}:
			os.Stdout.WriteString(i["text"].(string))
		}
	}
	os.Stdout.WriteString("\n")
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
