//Stores and loads text and its translation
package locale

import (
	"bufio"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

var (
	data map[string]map[string]string = make(map[string]map[string]string)
	lock sync.RWMutex
)

func Load(pathDir string) {
	lock.Lock()
	defer lock.Unlock()
	dir, err := os.Open(pathDir)
	if err != nil {
		log.Printf("Failed to load %s - %s", pathDir, err)
		return
	}
	files, err := dir.Readdir(0)
	if err != nil {
		log.Printf("Failed to load %s - %s", pathDir, err)
		return
	}
	for _, f := range files {
		if path.Ext(f.Name()) != ".lang" {
			continue
		}
		file, err := os.Open(path.Join(pathDir, f.Name()))
		if err != nil {
			log.Println("Failed to load %s", f.Name())
			continue
		}
		r := bufio.NewReader(file)
		locale := f.Name()[:len(f.Name())-5]
		for true {
			line, err := r.ReadString('\n')
			if strings.Contains(line, "=") {
				args := strings.Split(line, "=")
				args[0] = strings.TrimSpace(args[0])
				args[1] = strings.TrimSpace(args[1])
				put(locale, args[0], args[1])
			}
			if err != nil {
				break
			}
		}
	}
}

func put(locale, name, value string) {
	if _, ok := data[locale]; !ok {
		data[locale] = make(map[string]string)
	}
	data[locale][name] = value
}

func Get(locale, name string) string {
	lock.RLock()
	defer lock.RUnlock()
	if l, ok := data[locale]; ok {
		if v, ok := l[name]; ok {
			return v
		}
	}
	return get("en_GB", name)
}

func get(locale, name string) string {
	if l, ok := data[locale]; ok {
		if v, ok := l[name]; ok {
			return v
		}
	}
	return "!#!" + name + "!#!"
}
