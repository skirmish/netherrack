package debug

import (
	"Netherrack/event"
	"Soulsand"
	"fmt"
	"math"
	"runtime"
	"time"
)

func init() {
	if false {
		go monitor()
	}
}

func GetMemory() uint64 {
	runtime.ReadMemStats(&stats)
	return stats.Alloc / uint64(math.Pow(10, 6))
}

var stats runtime.MemStats

func monitor() {
	t := time.Tick(5 * time.Second)
	for {
		<-t
		runtime.ReadMemStats(&stats)
		message(fmt.Sprintf(Soulsand.ColourDarkGray + "[" + Soulsand.ColourDarkBlue + "Debug" + Soulsand.ColourDarkGray + "]"))
		message(fmt.Sprintf(Soulsand.ColourDarkBlue+"Goroutines: "+Soulsand.ColourCyan+"%d", runtime.NumGoroutine()))
		message(fmt.Sprintf(Soulsand.ColourDarkBlue+"Memory: "+Soulsand.ColourCyan+"%s", formatBytes(stats.Alloc)))
	}
}

func message(str string) {
	event.Broadcast(str)
}

func formatBytes(size uint64) string {
	switch {
	case size/uint64(math.Pow(10, 9)) >= 1: //Giga
		return fmt.Sprintf("%dGB", size/uint64(math.Pow(10, 9)))
	case size/uint64(math.Pow(10, 6)) >= 1: //Mega
		return fmt.Sprintf("%dMB", size/uint64(math.Pow(10, 6)))
	case size/uint64(math.Pow(10, 3)) >= 1: //Kili
		return fmt.Sprintf("%dKB", size/uint64(math.Pow(10, 3)))
	}
	return fmt.Sprintf("%dB", size)
}
