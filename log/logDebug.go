// +build debug

package log

import ()

func Debug(str string) {
	Println(str, 3)
}
