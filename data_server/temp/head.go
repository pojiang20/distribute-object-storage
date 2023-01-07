package temp

import (
	"fmt"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"log"
	"net/http"
	"os"
)

// 返回数据内容长度
func head(w http.ResponseWriter, r *http.Request) {
	uuid := utils.GetObjectName(r.URL.EscapedPath())
	f, err := os.Open(os.Getenv("STORAGE_ROOT") + "/temp/" + uuid + ".dat")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-length", fmt.Sprintf("%d", info.Size()))
}
