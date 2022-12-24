package temp

import (
	"encoding/json"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type tempInfo struct {
	Uuid string
	Name string
	Size int64
}

func post(w http.ResponseWriter, r *http.Request) {
	output, _ := exec.Command("uuidgen").Output()
	uuid := strings.TrimSuffix(string(output), "\n")
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
	os.Create(os.Getenv("STTORAGE_ROOT") + "/temp/" + tmp.Uuid + ".dat")
	w.Write([]byte(uuid))
}

// $STORAGE_ROOT/temp/下创建名为<uuid>的文件，将tempInfo编码后存储，来保存临时对象信息
func (t *tempInfo) writeToFile() error {
	f, err := os.Create(os.Getenv("STORAGE_ROOT") + "/temp/" + t.Uuid)
	if err != nil {
		return err
	}
	defer f.Close()
	b, _ := json.Marshal(t)
	f.Write(b)
	return nil
}
