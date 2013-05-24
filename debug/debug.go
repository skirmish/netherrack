package debug

import (
	"fmt"
	"io"
	"strings"
	"time"
)

/*
Output:
Name: (TotalTime)
	Name: (TotalTime)
	Name: (TotalTime)
		Name: (TotalTime)
*/

var (
	main      *step
	current   *step
	startTime time.Time
)

type step struct {
	Name      string
	Subs      []*step
	Prev      *step
	TimeTaken int64
}

func (s *step) Print(w io.Writer, level int, hideZero bool) {
	tabs := strings.Repeat("\t", level)
	if hideZero && s.TimeTaken == 0 {
		return
	}
	io.WriteString(w, fmt.Sprintf("%s%s: (Time: %dns)\n", tabs, s.Name, s.TimeTaken))
	for _, sub := range s.Subs {
		sub.Print(w, level+1, hideZero)
	}
}

func Start(name string) {
	main = &step{
		Name: name,
		Subs: make([]*step, 0),
	}
	current = main
	startTime = time.Now()
}

func Print(w io.Writer, hideZero bool) {
	io.WriteString(w, fmt.Sprintf("Output for %s:\n", main.Name))
	main.Print(w, 0, hideZero)
}

func Stop() {
	main.TimeTaken += time.Now().Sub(startTime).Nanoseconds()
}

func StepIn(name string) {
	current.TimeTaken += time.Now().Sub(startTime).Nanoseconds()
	current = &step{
		Name: name,
		Subs: make([]*step, 0),
		Prev: current,
	}
	current.Prev.Subs = append(current.Prev.Subs, current)
	startTime = time.Now()
}

func StepOut() {
	current.TimeTaken += time.Now().Sub(startTime).Nanoseconds()
	current.Prev.TimeTaken += current.TimeTaken
	current = current.Prev
	startTime = time.Now()
}
