package log

import (
	"bytes"
	"fmt"
	"hash/adler32"
	"os"
	"runtime"
	"strings"
	"sync"
)

const (
	colourBlack       = "\x1b[0;30m"
	colourDarkRed     = "\x1b[0;31m"
	colourDarkGreen   = "\x1b[0;32m"
	colourGold        = "\x1b[0;33m"
	colourDarkBlue    = "\x1b[0;34m"
	colourDarkMagenta = "\x1b[0;35m"
	colourDarkCyan    = "\x1b[0;36m"
	colourGrey        = "\x1b[0;37m"
	colourDarkGrey    = "\x1b[0;30;1m"
	colourRed         = "\x1b[0;31;1m"
	colourGreen       = "\x1b[0;32;1m"
	colourYellow      = "\x1b[0;33;1m"
	colourBlue        = "\x1b[0;34;1m"
	colourMagenta     = "\x1b[0;35;1m"
	colourCyan        = "\x1b[0;36;1m"
	colourWhite       = "\x1b[0;37;1m"
)

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

var fileColours = []string{colourRed, colourCyan, colourGreen, colourBlue, colourMagenta}
var filePathMap = map[string]string{}

func printFileInfo(level int) {
	_, file, line, _ := runtime.Caller(level)

	path := file[:strings.LastIndex(file, "/")]
	colour, ok := filePathMap[path]
	if !ok {
		pos := adler32.Checksum([]byte(path)) % uint32(len(fileColours))
		colour = fileColours[pos]
		filePathMap[path] = colour
	}
	os.Stdout.WriteString(colour)
	fmt.Fprint(os.Stdout, file[strings.LastIndex(file, "/")+1:], ":")
	os.Stdout.WriteString(colourYellow)
	fmt.Fprint(os.Stdout, line, " ")
	os.Stdout.WriteString(colourWhite)
}