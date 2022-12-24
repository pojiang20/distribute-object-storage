package temp

import (
	"os"
	"path"
	"time"
)

func CleanEvery12hour() {
	dir := path.Join(os.Getenv("STORAGE_ROOT"), "temp")
	ticker := time.NewTicker(12 * time.Hour)
	for {
		<-ticker.C
		files, _ := os.ReadDir(dir)
		for _, item := range files {
			info, _ := item.Info()
			dif := int(info.ModTime().Sub(time.Now()).Minutes())
			if dif >= 30 {
				os.Remove(path.Join(os.Getenv("STORAGE_ROOT"), "temp", info.Name()))
			}
		}
	}
}
