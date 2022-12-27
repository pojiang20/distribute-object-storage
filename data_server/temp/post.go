package temp

import (
	"encoding/json"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type tempInfo struct {
	UUID string
	Name string
	Size int64
}

func post(w http.ResponseWriter, r *http.Request) {
	output, _ := exec.Command("uuidgen").Output()
	uuid := strings.ReplaceAll(string(output), "\n", "")
	name := utils.GetObjectName(r.URL.EscapedPath())
	size, err := strconv.ParseInt(r.Header.Get("size"), 0, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmp := tempInfo{uuid, name, size}
	err = tmp.writeToFile()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = os.Create(path.Join(os.Getenv("STORAGE_ROOT"), "temp", tmp.UUID+".dat"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte(uuid))
}

// $STORAGE_ROOT/temp/下创建名为<uuid>的文件，将tempInfo编码后存储，来保存临时对象信息
func (t *tempInfo) writeToFile() error {
	path1 := path.Join(os.Getenv("STORAGE_ROOT"), "temp", t.UUID)
	log.Printf("[%s]: %+v", path1, t)
	f, err := os.Create(path1)
	if err != nil {
		return err
	}
	defer f.Close()
	b, _ := json.Marshal(t)
	f.Write(b)
	return nil
}
