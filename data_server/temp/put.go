package temp

import (
	"github.com/pojiang20/distribute-object-storage/api_server/locate"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"log"
	"net/http"
	"os"
	"path"
)

func put(w http.ResponseWriter, r *http.Request) {
	uuid := utils.GetObjectName(r.URL.EscapedPath())
	tempinfo, err := readFromFile(uuid)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	//info存放了元数据、dat存放了对象内容数据
	infoFile := path.Join(os.Getenv("STORAGE_ROOT"), "tmp", uuid)
	datFile := infoFile + ".dat"
	f, err := os.Open(datFile)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	actualSize := info.Size()
	os.Remove(infoFile)
	if actualSize != tempinfo.Size {
		os.Remove(datFile)
		log.Printf("acutal size %d,not equal %d\n", actualSize, tempinfo.Size)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	commitTempObject(datFile, tempinfo)
}

// 将临时对象转正
func commitTempObject(datFile string, tempinfo *tempInfo) {
	os.Rename(datFile, path.Join(os.Getenv("STORAGE_ROOT"), "objects", tempinfo.Name))
	locate.Add(tempinfo.Name)
}
