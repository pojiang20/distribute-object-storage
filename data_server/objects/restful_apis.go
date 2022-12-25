package objects

import (
	"github.com/pojiang20/distribute-object-storage/data_server/locate"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
)

const (
	STORAGE_ROOT = "STORAGE_ROOT"
)

func get(w http.ResponseWriter, r *http.Request) {
	file := getFile(utils.GetObjectName(r.URL.EscapedPath()))
	if file == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	sendFile(w, file)
}

func getFile(hash string) string {
	file := path.Join(os.Getenv("STORAGE_ROOT"), "objects", hash)
	f, _ := os.Open(file)
	d := url.PathEscape(utils.CalculateHash(f))
	f.Close()
	// 校验：校验接口层中ES存储的哈希值与实际存储的内容的哈希值是否一致，若是发生了变化则不一致，并且删除该对象数据
	// 数据存放久了可能会发生数据降解等问题，因此有必要做一致性校验
	if d != hash {
		log.Println("object hash mismatch,remove ", file)
		locate.Add(hash)
		os.Remove(file)
		return ""
	}
	return file
}

func sendFile(w io.Writer, file string) {
	f, _ := os.Open(file)
	defer f.Close()
	io.Copy(w, f)
}
