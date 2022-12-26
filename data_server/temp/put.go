package temp

import (
	"github.com/pojiang20/distribute-object-storage/data_server/locate"
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
	infoFile := path.Join(os.Getenv("STORAGE_ROOT"), "temp", uuid)
	datFile := infoFile + ".dat"
	f, err := os.OpenFile(datFile, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	actualInfo, err := os.Stat(datFile)
	os.Remove(infoFile)
	if actualInfo.Size() != tempinfo.Size {
		os.Remove(datFile)
		log.Printf("Error: the actual uploaded file`s size [%d] is dismatched with expected size [%d]\n",
			actualInfo.Size(), tempinfo.Size)
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
