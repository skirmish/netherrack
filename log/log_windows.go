package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NetherrackDev/soulsand/chat"
	"hash/adler32"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
)

const (
	stdOutputHandle uintptr = 4294967285

	colourBlack       uintptr = 0
	colourDarkBlue            = 1
	colourDarkGreen           = 2
	colourDarkCyan            = 3
	colourDarkRed             = 4
	colourDarkMagenta         = 5
	colourGold                = 6
	colourGrey                = 7
	colourDarkGrey            = 8
	colourBlue                = 9
	colourGreen               = 10
	colourCyan                = 11
	colourRed                 = 12
	colourMagenta             = 13
	colourYellow              = 14
	colourWhite               = 15
)

var (
	dll                     = syscall.MustLoadDLL("Kernel32.dll")
	getStdHandle            = dll.MustFindProc("GetStdHandle")
	setConsoleTextAttribute = dll.MustFindProc("SetConsoleTextAttribute")
	stdOutput               uintptr
	globalLock              sync.Mutex
)

func init() {
	stdOutput, _, _ = getStdHandle.Call(uintptr(stdOutputHandle))
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

func Fatalln(args ...interface{}) {
	Println(args...)
	os.Exit(-1)
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
	err := json.Unmarshal(chatMsg.Bytes(), &msg)
	if err != nil {
		panic(err)
	}
	text, ok := msg["text"].([]interface{})
	if !ok {
		panic("Not implemented")
	}
	for _, i := range text {
		setConsoleAttribute(colourWhite)
		switch i := i.(type) {
		case string:
			os.Stdout.WriteString(stripColourCodes(i))
		case map[string]interface{}:
			if col, ok := i["color"]; ok {
				switch col.(string) {
				case "black":
					setConsoleAttribute(colourBlack)
				case "dark_blue":
					setConsoleAttribute(colourDarkBlue)
				case "dark_green":
					setConsoleAttribute(colourDarkGreen)
				case "dark_aqua":
					setConsoleAttribute(colourDarkCyan)
				case "dark_red":
					setConsoleAttribute(colourDarkRed)
				case "dark_purple":
					setConsoleAttribute(colourDarkMagenta)
				case "gold":
					setConsoleAttribute(colourGold)
				case "gray":
					setConsoleAttribute(colourGrey)
				case "dark_gray":
					setConsoleAttribute(colourDarkGrey)
				case "blue":
					setConsoleAttribute(colourBlue)
				case "green":
					setConsoleAttribute(colourGreen)
				case "aqua":
					setConsoleAttribute(colourCyan)
				case "red":
					setConsoleAttribute(colourRed)
				case "light_purple":
					setConsoleAttribute(colourMagenta)
				case "yellow":
					setConsoleAttribute(colourYellow)
				case "white":
					setConsoleAttribute(colourWhite)
				}
			}
			os.Stdout.WriteString(stripColourCodes(i["text"].(string)))
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

var fileColours = []uintptr{colourRed, colourCyan, colourGreen, colourBlue, colourMagenta}
var filePathMap = map[string]uintptr{}

func printFileInfo(level int) {
	_, file, line, _ := runtime.Caller(level)

	path := file[:strings.LastIndex(file, "/")]
	colour, ok := filePathMap[path]
	if !ok {
		pos := adler32.Checksum([]byte(path)) % uint32(len(fileColours))
		colour = fileColours[pos]
		filePathMap[path] = colour
	}
	setConsoleAttribute(colour)
	fmt.Fprint(os.Stdout, file[strings.LastIndex(file, "/")+1:], ":")
	setConsoleAttribute(colourYellow)
	fmt.Fprint(os.Stdout, line, " ")
	setConsoleAttribute(colourWhite)
}

func setConsoleAttribute(attrib uintptr) {
	setConsoleTextAttribute.Call(stdOutput, attrib)
}
