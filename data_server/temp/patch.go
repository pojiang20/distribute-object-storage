package temp

import (
	"encoding/json"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

// 缓存对象数据，初步校验数据的大小是否匹配
// patch接口以uuid标识对象
func patch(w http.ResponseWriter, r *http.Request) {
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
	log.Printf("patch: infoFile[%s],dataFile[%s]", infoFile, datFile)
	f, err := os.OpenFile(datFile, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	_, err = io.Copy(f, r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	info, err := f.Stat()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	actualSize := info.Size()
	if actualSize != tempinfo.Size {
		//os.Remove(datFile)
		//os.Remove(infoFile)
		log.Printf("acutal size %d,not equal %d\n", actualSize, tempinfo.Size)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 读取临时对象元数据
func readFromFile(uuid string) (*tempInfo, error) {
	f, err := os.Open(path.Join(os.Getenv("STORAGE_ROOT"), "temp", uuid))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, _ := io.ReadAll(f)
	var info *tempInfo
	json.Unmarshal(b, &info)
	return info, nil
}
