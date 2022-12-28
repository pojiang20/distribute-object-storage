package temp

import (
	"github.com/pojiang20/distribute-object-storage/data_server/locate"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
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
// 将"对象哈希"+"."+"分片编号"的临时分片文件转为名字形式为"对象哈希"+"."+"分片编号"+"."+"分片哈希"的正式文件
func commitTempObject(tempFilePath string, tempinfo *tempInfo) {
	shardFile, _ := os.Open(tempFilePath)
	shardHash := url.PathEscape(utils.CalculateHash(shardFile))
	shardFile.Close()
	os.Rename(tempFilePath, path.Join(os.Getenv("STORAGE_ROOT"), "objects", tempinfo.Name+"."+shardHash))
	locate.Add(tempinfo.hash(), tempinfo.id())
}

func (t *tempInfo) hash() string {
	return strings.Split(t.Name, ".")[0]
}

func (t *tempInfo) id() int {
	id, _ := strconv.Atoi(strings.Split(t.Name, ".")[1])
	return id
}
