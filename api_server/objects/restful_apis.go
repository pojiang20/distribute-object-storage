package objects

import (
	"fmt"
	"github.com/pojiang20/distribute-object-storage/api_server/heartbeat"
	"github.com/pojiang20/distribute-object-storage/api_server/locate"
	"github.com/pojiang20/distribute-object-storage/src/es"
	"github.com/pojiang20/distribute-object-storage/src/rs"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func put(w http.ResponseWriter, r *http.Request) {
	hash := utils.GetHashFromHeader(r.Header)
	if hash == "" {
		log.Println("missing object hash in digest header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	size := utils.GetSizeFromHeader(r.Header)
	respCode, err := storeObject(r.Body, hash, size)
	if err != nil {
		log.Println(err)
		w.WriteHeader(respCode)
	}
	if respCode != http.StatusOK {
		w.WriteHeader(respCode)
		return
	}
	name := utils.GetObjectName(r.URL.EscapedPath())
	err = es.AddVersion(name, size, hash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	name := utils.GetObjectName(r.URL.EscapedPath())
	versionId := r.URL.Query().Get("version")
	version := 0
	var err error
	if len(versionId) != 0 {
		version, err = strconv.Atoi(versionId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	meta, err := es.GetMetadata(name, version)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if meta.Hash == "" {
		log.Printf("ES INFO: object [%s] not found", name)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	stream, err := GetStream(url.PathEscape(meta.Hash), meta.Size)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	offset, end := utils.GetOffsetFromHeader(r.Header)
	log.Printf("apiServer INFO: in get(), get object data range [%d, %d]\n", offset, end)
	contentLength := meta.Size
	if offset != 0 {
		//截取ooffset到结尾的长度
		contentLength = end - offset + 1
		stream.Seek(offset, io.SeekCurrent)
		w.Header().Set("content-range", fmt.Sprintf("bytes_%d-%d/%d", offset, end, meta.Size))
		w.WriteHeader(http.StatusPartialContent)
	}
	n, _ := io.CopyN(w, stream, contentLength)
	log.Println("Wrote to response length:", n)
	stream.Close()
}

func del(w http.ResponseWriter, r *http.Request) {
	name := utils.GetObjectName(r.URL.EscapedPath())
	version, err := es.SearchLatestVersion(name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = es.PutMetadata(name, version.Version+1, 0, "")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// 构造RSResumablePutStream对象，编码返回
func post(w http.ResponseWriter, r *http.Request) {
	objectName := utils.GetObjectName(r.URL.EscapedPath())
	size, err := strconv.ParseInt(r.Header.Get("size"), 0, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hash := utils.GetHashFromHeader(r.Header)
	if hash == "" {
		log.Printf("apiServer Error: missing object [%s] hash\n", objectName)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if locate.Exist(url.PathEscape(hash)) {
		err = es.AddVersion(objectName, size, hash)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	dataServers := heartbeat.ChooseServers(rs.ALL_SHARDS, nil)
	if len(dataServers) != rs.ALL_SHARDS {
		log.Println("apiServer Error: dataServer is not enough")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	stream, err := rs.NewRSResumablePutStream(dataServers, objectName, size, hash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", stream.ToToken())
	w.WriteHeader(http.StatusCreated)
}
