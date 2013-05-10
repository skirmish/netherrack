package chunk

import (
	"compress/gzip"
	"github.com/NetherrackDev/netherrack/nbt"
	"os"
	"path/filepath"
)

func (world *World) loadLevel() {
	os.MkdirAll(filepath.Join("worlds", world.Name), os.ModeDir|os.ModePerm)

	levelDataFile, err := os.Open(filepath.Join("worlds", world.Name, "level.dat"))

	var levelData nbt.Type
	if err != nil {
		levelData = nbt.NewNBT()
	} else {
		defer levelDataFile.Close()
		gz, err := gzip.NewReader(levelDataFile)
		defer gz.Close()
		if err != nil {
			panic(err)
		}
		levelData = nbt.Parse(gz)
	}
	world.settings, _ = levelData.GetCompound("Data", true)
}

func (world *World) save() {
	levelData := nbt.NewNBT()
	levelData.Set("Data", world.settings)
	levelDataFile, err := os.Create(filepath.Join("worlds", world.Name, "level.dat"))
	defer levelDataFile.Close()
	if err != nil {
		panic(err)
	}
	gz := gzip.NewWriter(levelDataFile)
	defer gz.Close()
	levelData.WriteTo(gz, "")
}
