package objects

import (
	"fmt"
	"github.com/pojiang20/distribute-object-storage/data_server/locate"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	//模糊搜索文件名以["对象哈希"+"."+"切片编号"]开头的文件，只能匹配到一个文件
	files, _ := filepath.Glob(path.Join(os.Getenv("STORAGE_ROOT"), "objects", fmt.Sprintf("%s.*", hash)))
	//一个数据节点，某一个对象只存在一个分片
	if len(files) != 1 {
		return ""
	}
	shardFileName := files[0]
	f, _ := os.Open(shardFileName)
	shardFileHash := url.PathEscape(utils.CalculateHash(f))
	f.Close()
	expectedShardHash := strings.Split(shardFileName, ".")[2]
	// 校验：校验接口层中ES存储的哈希值与实际存储的内容的哈希值是否一致，若是发生了变化则不一致，并且删除该对象数据
	// 数据存放久了可能会发生数据降解等问题，因此有必要做一致性校验
	if shardFileHash != expectedShardHash {
		log.Println("object hash mismatch,remove ", hash)
		locate.Del(hash)
		os.Remove(shardFileName)
		return ""
	}
	return shardFileName
}

func sendFile(w io.Writer, file string) {
	f, _ := os.Open(file)
	defer f.Close()
	io.Copy(w, f)
}
