package log

import (
	"bytes"
	"fmt"
	"github.com/NetherrackDev/soulsand"
	"hash/adler32"
	"os"
	"runtime"
	"strconv"
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
			os.Stdout.WriteString(buf.String())
			buf.Reset()
			i++
			if msgRunes[i] == 'r' {
				setConsoleAttribute(colourWhite)
				continue
			}
			colourId, err := strconv.ParseInt(string(msgRunes[i]), 16, 8)
			if err != nil {
				continue
			}
			setConsoleAttribute(uintptr(colourId))
			continue
		}
		buf.WriteRune(c)
	}
	buf.WriteRune('\n')
	os.Stdout.WriteString(buf.String())
	setConsoleAttribute(colourWhite)
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

var fileColours = []uintptr{colourRed, colourCyan, colourGreen, colourBlue, colourMagenta}
var filePathMap = map[string]uintptr{}

func printFileInfo() {
	_, file, line, _ := runtime.Caller(2)

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
